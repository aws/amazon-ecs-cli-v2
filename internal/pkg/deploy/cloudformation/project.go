// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/deploy/cloudformation/stack"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// CreateProjectResources sets up everything required for our project-wide resources.
// These resources include things that are regional, rather than scoped to a particular
// environment, such as ECR Repos, CodePipeline KMS keys & S3 buckets.
// We deploy project resources through StackSets - that way we can have one
// template that we update and all regional stacks are updated.
func (cf CloudFormation) CreateProjectResources(project *archer.Project) error {
	projectConfig := stack.NewProjectStackConfig(project, cf.box)

	// First deploy the project roles needed by StackSets. These roles
	// allow the stack set to set up our regional stacks.
	if err := cf.deploy(projectConfig); err == nil {
		_, err := cf.waitForStackCreation(projectConfig)
		if err != nil {
			return err
		}
	} else {
		// If the stack already exists - we can move on
		// to creating the StackSet.
		var alreadyExists *ErrStackAlreadyExists
		if !errors.As(err, &alreadyExists) {
			return err
		}
	}

	blankProjectTemplate, err := projectConfig.ResourceTemplate(&stack.ProjectResourcesConfig{
		Project: projectConfig.Project.Name,
	})

	if err != nil {
		return err
	}

	_, err = cf.client.CreateStackSet(&cloudformation.CreateStackSetInput{
		Description:           aws.String(projectConfig.StackSetDescription()),
		StackSetName:          aws.String(projectConfig.StackSetName()),
		TemplateBody:          aws.String(blankProjectTemplate),
		ExecutionRoleName:     aws.String(projectConfig.StackSetExecutionRoleName()),
		AdministrationRoleARN: aws.String(projectConfig.StackSetAdminRoleARN()),
		Tags:                  projectConfig.Tags(),
	})

	if err != nil && !stackSetExists(err) {
		return err
	}

	return nil
}

// AddAppToProject attempts to add new App specific resources to the Project resource stack.
// Currently, this means that we'll set up an ECR repo with a policy for all envs to be able
// to pull from it.
func (cf CloudFormation) AddAppToProject(project *archer.Project, newApp *archer.Application) error {
	projectConfig := stack.NewProjectStackConfig(project, cf.box)
	previouslyDeployedConfig, err := cf.getLastDeployedProjectConfig(projectConfig)
	if err != nil {
		return fmt.Errorf("adding %s app resources to project %s: %w", newApp.Name, project.Name, err)
	}

	// We'll generate a new list of Accounts to add to our project
	// infrastructure by appending the environment's account if it
	// doesn't already exist.
	appList := []string{}
	shouldAddNewApp := true
	for _, app := range previouslyDeployedConfig.Apps {
		appList = append(appList, app)
		if app == newApp.Name {
			shouldAddNewApp = false
		}
	}

	if !shouldAddNewApp {
		return nil
	}

	appList = append(appList, newApp.Name)

	newDeploymentConfig := stack.ProjectResourcesConfig{
		Version:  previouslyDeployedConfig.Version + 1,
		Apps:     appList,
		Accounts: previouslyDeployedConfig.Accounts,
		Project:  projectConfig.Project.Name,
	}
	if err := cf.deployProjectConfig(projectConfig, &newDeploymentConfig); err != nil {
		return fmt.Errorf("adding %s app resources to project: %w", newApp.Name, err)
	}

	return nil
}

// AddEnvToProject takes a new environment and updates the Project configuration
// with new Account IDs in resource policies (KMS Keys and ECR Repos) - and
// sets up a new stack instance if the environment is in a new region.
func (cf CloudFormation) AddEnvToProject(project *archer.Project, env *archer.Environment) error {
	projectConfig := stack.NewProjectStackConfig(project, cf.box)
	previouslyDeployedConfig, err := cf.getLastDeployedProjectConfig(projectConfig)
	if err != nil {
		return fmt.Errorf("getting previous deployed stackset %w", err)
	}

	// We'll generate a new list of Accounts to add to our project
	// infrastructure by appending the environment's account if it
	// doesn't already exist.
	accountList := []string{}
	shouldAddNewAccountID := true
	for _, accountID := range previouslyDeployedConfig.Accounts {
		accountList = append(accountList, accountID)
		if accountID == env.AccountID {
			shouldAddNewAccountID = false
		}
	}

	if shouldAddNewAccountID {
		accountList = append(accountList, env.AccountID)
	}

	newDeploymentConfig := stack.ProjectResourcesConfig{
		Version:  previouslyDeployedConfig.Version + 1,
		Apps:     previouslyDeployedConfig.Apps,
		Accounts: accountList,
		Project:  projectConfig.Project.Name,
	}

	if err := cf.deployProjectConfig(projectConfig, &newDeploymentConfig); err != nil {
		return fmt.Errorf("adding %s environment resources to project: %w", env.Name, err)
	}

	if err := cf.addNewProjectStackInstances(projectConfig, env); err != nil {
		return fmt.Errorf("adding new stack instance for environment %s: %w", env.Name, err)
	}

	return nil
}

func (cf CloudFormation) deployProjectConfig(projectConfig *stack.ProjectStackConfig, resources *stack.ProjectResourcesConfig) error {
	newTemplateToDeploy, err := projectConfig.ResourceTemplate(resources)
	if err != nil {
		return err
	}
	// Every time we deploy the StackSet, we include a version field in the stack metadata.
	// When we go to update the StackSet, we include that version + 1 as the "Operation ID".
	// This ensures that we don't overwrite any changes that may have been applied between
	// us reading the stack and actually updating it.
	// As an example:
	//  * We read the stack with Version 1
	//  * Someone else reads the stack with Version 1
	//  * We update the StackSet with Version 2, the update completes.
	//  * Someone else tries to update the StackSet with their stale version 2.
	//  * "2" has already been used as an operation ID, and the stale write fails.
	input := cloudformation.UpdateStackSetInput{
		TemplateBody:          aws.String(newTemplateToDeploy),
		OperationId:           aws.String(fmt.Sprintf("%d", resources.Version)),
		StackSetName:          aws.String(projectConfig.StackSetName()),
		Description:           aws.String(projectConfig.StackSetDescription()),
		ExecutionRoleName:     aws.String(projectConfig.StackSetExecutionRoleName()),
		AdministrationRoleARN: aws.String(projectConfig.StackSetAdminRoleARN()),
		Tags:                  projectConfig.Tags(),
	}
	output, err := cf.client.UpdateStackSet(&input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudformation.ErrCodeOperationIdAlreadyExistsException, cloudformation.ErrCodeOperationInProgressException, cloudformation.ErrCodeStaleRequestException:
				return &ErrStackSetOutOfDate{projectName: projectConfig.Project.Name, parentErr: err}
			}
		}
		return fmt.Errorf("updating project resources: %w", err)
	}

	return cf.waitForStackSetOperation(projectConfig.StackSetName(), *output.OperationId)

}

// addNewStackInstances takes an environment and determines if we need to create a new
// stack instance. We only spin up a new stack instance if the env is in a new region.
func (cf CloudFormation) addNewProjectStackInstances(projectConfig *stack.ProjectStackConfig, env *archer.Environment) error {
	stackInstances, err := cf.client.ListStackInstances(&cloudformation.ListStackInstancesInput{
		StackSetName: aws.String(projectConfig.StackSetName()),
	})

	if err != nil {
		return fmt.Errorf("fetching existing project stack instances: %w", err)
	}

	// We only want to deploy a new StackInstance if we're
	// adding an environment in a new region.
	shouldDeployNewStackInstance := true
	for _, stackInstance := range stackInstances.Summaries {
		if *stackInstance.Region == env.Region {
			shouldDeployNewStackInstance = false
		}
	}

	if !shouldDeployNewStackInstance {
		return nil
	}

	// Set up a new Stack Instance for the new region. The Stack Instance will inherit
	// the latest StackSet template.
	createStacksOutput, err := cf.client.CreateStackInstances(&cloudformation.CreateStackInstancesInput{
		Accounts:     []*string{aws.String(projectConfig.Project.AccountID)},
		Regions:      []*string{aws.String(env.Region)},
		StackSetName: aws.String(projectConfig.StackSetName()),
	})

	if err != nil {
		return fmt.Errorf("creating new project stack instances: %w", err)
	}

	return cf.waitForStackSetOperation(projectConfig.StackSetName(), *createStacksOutput.OperationId)
}

func (cf CloudFormation) getLastDeployedProjectConfig(projectConfig *stack.ProjectStackConfig) (*stack.ProjectResourcesConfig, error) {
	// Check the existing deploy stack template. From that template, we'll parse out the list of apps and accounts that
	// are deployed in the stack.
	describeOutput, err := cf.client.DescribeStackSet(&cloudformation.DescribeStackSetInput{
		StackSetName: aws.String(projectConfig.StackSetName()),
	})
	previouslyDeployedConfig, err := stack.ProjectConfigFrom(describeOutput.StackSet.TemplateBody)
	if err != nil {
		return nil, fmt.Errorf("parsing previous deployed stackset %w", err)
	}
	return previouslyDeployedConfig, nil
}

func (cf CloudFormation) waitForStackSetOperation(stackSetName, operationID string) error {
	for {
		response, err := cf.client.DescribeStackSetOperation(&cloudformation.DescribeStackSetOperationInput{
			OperationId:  aws.String(operationID),
			StackSetName: aws.String(stackSetName),
		})

		if err != nil {
			return fmt.Errorf("fetching stack set operation status: %w", err)
		}

		if *response.StackSetOperation.Status == "STOPPED" {
			return fmt.Errorf("project operation %s in stack set %s was manually stopped", operationID, stackSetName)
		}

		if *response.StackSetOperation.Status == "FAILED" {
			return fmt.Errorf("project operation %s in stack set %s failed", operationID, stackSetName)
		}

		if *response.StackSetOperation.Status == "SUCCEEDED" {
			return nil
		}

		time.Sleep(3 * time.Second)
	}
}
