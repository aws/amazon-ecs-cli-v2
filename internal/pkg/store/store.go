// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Package store implements CRUD operations for project, environment, application and
pipeline configuration. This configuration contains the archer projects
a customer has, and the environments and pipelines associated with each
project.
*/
package store

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/identity"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/aws/aws-sdk-go/service/route53domains/route53domainsiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// Parameter name formats for resources in a project. Projects are laid out in SSM
// based on path - each parameter's key has a certain format, and you can have
// hierarchies based on that format. Projects are at the root of the hierarchy.
// Searching SSM for all parameters with the `rootProjectPath` key will give you
// all the project keys, for example.

// current schema Version for Projects
const schemaVersion = "1.0"

// schema formats supported in current schemaVersion. NOTE: May change to map in the future.
const (
	rootProjectPath  = "/archer/"
	fmtProjectPath   = "/archer/%s"
	rootEnvParamPath = "/archer/%s/environments/"
	fmtEnvParamPath  = "/archer/%s/environments/%s" // path for an environment in a project
	rootAppParamPath = "/archer/%s/applications/"
	fmtAppParamPath  = "/archer/%s/applications/%s" // path for an application in a project
)

type identityService interface {
	Get() (identity.Caller, error)
}

// Store is in charge of fetching and creating projects, environment and pipeline configuration in SSM.
type Store struct {
	idClient      identityService
	domainsClient route53domainsiface.Route53DomainsAPI
	ssmClient     ssmiface.SSMAPI
	sessionRegion string
}

// New returns a Store allowing you to query or create Projects or Environments.
func New() (*Store, error) {
	sess, err := session.Default()

	if err != nil {
		return nil, err
	}

	return &Store{
		idClient: identity.New(sess),
		// See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DNSLimitations.html#limits-service-quotas
		// > To view limits and request higher limits for Route 53, you must change the Region to US East (N. Virginia).
		// So we have to set the region to us-east-1 to be able to find out if a domain name exists in the account.
		domainsClient: route53domains.New(sess, aws.NewConfig().WithRegion("us-east-1")),
		ssmClient:     ssm.New(sess),
		sessionRegion: *sess.Config.Region,
	}, nil
}

// CreateProject instantiates a new project, validates its uniqueness and stores it in SSM.
func (s *Store) CreateProject(project *archer.Project) error {
	projectPath := fmt.Sprintf(fmtProjectPath, project.Name)
	project.Version = schemaVersion

	data, err := marshal(project)
	if err != nil {
		return fmt.Errorf("serializing project %s: %w", project.Name, err)
	}

	if project.Domain != "" {
		in := &route53domains.GetDomainDetailInput{DomainName: aws.String(project.Domain)}
		if _, err := s.domainsClient.GetDomainDetail(in); err != nil {
			return fmt.Errorf("get domain details for %s: %w", project.Domain, err)
		}
	}

	_, err = s.ssmClient.PutParameter(&ssm.PutParameterInput{
		Name:        aws.String(projectPath),
		Description: aws.String("An ECS-CLI Project"),
		Type:        aws.String(ssm.ParameterTypeString),
		Value:       aws.String(data),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterAlreadyExists:
				return &ErrProjectAlreadyExists{
					ProjectName: project.Name,
				}
			}
		}
		return fmt.Errorf("create project %s: %w", project.Name, err)
	}
	return nil
}

// GetProject fetches a project by name. If it can't be found, return a ErrNoSuchProject
func (s *Store) GetProject(projectName string) (*archer.Project, error) {
	projectPath := fmt.Sprintf(fmtProjectPath, projectName)
	projectParam, err := s.ssmClient.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(projectPath),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterNotFound:
				account, region := s.getCallerAccountAndRegion()
				return nil, &ErrNoSuchProject{
					ProjectName: projectName,
					AccountID:   account,
					Region:      region,
				}
			}
		}
		return nil, fmt.Errorf("get project %s: %w", projectName, err)
	}

	var project archer.Project
	if err := json.Unmarshal([]byte(*projectParam.Parameter.Value), &project); err != nil {
		return nil, fmt.Errorf("read details for project %s: %w", projectName, err)
	}
	return &project, nil
}

// ListProjects returns the list of existing projects in the customer's account and region.
func (s *Store) ListProjects() ([]*archer.Project, error) {
	var projects []*archer.Project
	serializedProjects, err := s.listParams(rootProjectPath)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	for _, serializedProject := range serializedProjects {
		var project archer.Project
		if err := json.Unmarshal([]byte(*serializedProject), &project); err != nil {
			return nil, fmt.Errorf("read project details: %w", err)
		}

		projects = append(projects, &project)
	}
	return projects, nil
}

// CreateEnvironment instantiates a new environment within an existing project. Returns ErrEnvironmentAlreadyExists
// if the environment already exists in the project.
func (s *Store) CreateEnvironment(environment *archer.Environment) error {
	if _, err := s.GetProject(environment.Project); err != nil {
		return err
	}

	environmentPath := fmt.Sprintf(fmtEnvParamPath, environment.Project, environment.Name)
	data, err := marshal(environment)
	if err != nil {
		return fmt.Errorf("serializing environment %s: %w", environment.Name, err)
	}

	paramOutput, err := s.ssmClient.PutParameter(&ssm.PutParameterInput{
		Name:        aws.String(environmentPath),
		Description: aws.String(fmt.Sprintf("The %s deployment stage", environment.Name)),
		Type:        aws.String(ssm.ParameterTypeString),
		Value:       aws.String(data),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterAlreadyExists:
				return &ErrEnvironmentAlreadyExists{
					EnvironmentName: environment.Name,
					ProjectName:     environment.Project}
			}
		}
		return fmt.Errorf("create environment %s in project %s: %w", environment.Name, environment.Project, err)
	}

	log.Printf("Created environment with version %v", *paramOutput.Version)
	return nil

}

// GetEnvironment gets an environment belonging to a particular project by name. If no environment is found
// it returns ErrNoSuchEnvironment.
func (s *Store) GetEnvironment(projectName string, environmentName string) (*archer.Environment, error) {
	environmentPath := fmt.Sprintf(fmtEnvParamPath, projectName, environmentName)
	environmentParam, err := s.ssmClient.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(environmentPath),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterNotFound:
				return nil, &ErrNoSuchEnvironment{
					ProjectName:     projectName,
					EnvironmentName: environmentName,
				}
			}
		}
		return nil, fmt.Errorf("get environment %s in project %s: %w", environmentName, projectName, err)
	}

	var env archer.Environment
	err = json.Unmarshal([]byte(*environmentParam.Parameter.Value), &env)
	if err != nil {
		return nil, fmt.Errorf("read details for environment %s in project %s: %w", environmentName, projectName, err)
	}
	return &env, nil
}

// ListEnvironments returns all environments belonging to a particular project.
func (s *Store) ListEnvironments(projectName string) ([]*archer.Environment, error) {
	var environments []*archer.Environment

	environmentsPath := fmt.Sprintf(rootEnvParamPath, projectName)
	serializedEnvs, err := s.listParams(environmentsPath)
	if err != nil {
		return nil, fmt.Errorf("list environments for project %s: %w", projectName, err)
	}
	for _, serializedEnv := range serializedEnvs {
		var env archer.Environment
		if err := json.Unmarshal([]byte(*serializedEnv), &env); err != nil {
			return nil, fmt.Errorf("read environment details for project %s: %w", projectName, err)
		}

		environments = append(environments, &env)
	}
	return environments, nil
}

// CreateApplication instantiates a new application within an existing project. Returns ErrApplicationAlreadyExists
// if the application already exists in the project.
func (s *Store) CreateApplication(app *archer.Application) error {
	if _, err := s.GetProject(app.Project); err != nil {
		return err
	}

	applicationPath := fmt.Sprintf(fmtAppParamPath, app.Project, app.Name)
	data, err := marshal(app)
	if err != nil {
		return fmt.Errorf("serializing application %s: %w", app.Name, err)
	}

	paramOutput, err := s.ssmClient.PutParameter(&ssm.PutParameterInput{
		Name:        aws.String(applicationPath),
		Description: aws.String(fmt.Sprintf("ECS-CLI v2 Application %s", app.Name)),
		Type:        aws.String(ssm.ParameterTypeString),
		Value:       aws.String(data),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterAlreadyExists:
				return &ErrApplicationAlreadyExists{
					ApplicationName: app.Name,
					ProjectName:     app.Project}
			}
		}
		return fmt.Errorf("create application %s in project %s: %w", app.Name, app.Project, err)
	}

	log.Printf("Created Application with version %v", *paramOutput.Version)
	return nil

}

// GetApplication gets an application belonging to a particular project by name. If no app is found
// it returns ErrNoSuchApplication.
func (s *Store) GetApplication(projectName, appName string) (*archer.Application, error) {
	appPath := fmt.Sprintf(fmtAppParamPath, projectName, appName)
	appParam, err := s.ssmClient.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(appPath),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ssm.ErrCodeParameterNotFound:
				return nil, &ErrNoSuchApplication{
					ProjectName:     projectName,
					ApplicationName: appName,
				}
			}
		}
		return nil, fmt.Errorf("get application %s in project %s: %w", appName, projectName, err)
	}

	var app archer.Application
	err = json.Unmarshal([]byte(*appParam.Parameter.Value), &app)
	if err != nil {
		return nil, fmt.Errorf("read details for application %s in project %s: %w", appName, projectName, err)
	}
	return &app, nil
}

// ListApplications returns all applications belonging to a particular project.
func (s *Store) ListApplications(projectName string) ([]*archer.Application, error) {
	var applications []*archer.Application

	applicationsPath := fmt.Sprintf(rootAppParamPath, projectName)
	serializedApps, err := s.listParams(applicationsPath)
	if err != nil {
		return nil, fmt.Errorf("list applications for project %s: %w", projectName, err)
	}
	for _, serializedApp := range serializedApps {
		var app archer.Application
		if err := json.Unmarshal([]byte(*serializedApp), &app); err != nil {
			return nil, fmt.Errorf("read application details for project %s: %w", projectName, err)
		}

		applications = append(applications, &app)
	}
	return applications, nil
}

func (s *Store) listParams(path string) ([]*string, error) {
	var serializedParams []*string

	var nextToken *string = nil
	for {
		params, err := s.ssmClient.GetParametersByPath(&ssm.GetParametersByPathInput{
			Path:      aws.String(path),
			Recursive: aws.Bool(false),
			NextToken: nextToken,
		})

		if err != nil {
			return nil, err
		}

		for _, param := range params.Parameters {
			serializedParams = append(serializedParams, param.Value)
		}

		nextToken = params.NextToken
		if nextToken == nil {
			break
		}
	}
	return serializedParams, nil
}

// Retrieves the caller's Account ID with a best effort. If it fails to fetch the Account ID,
// this returns "unknown".
func (s *Store) getCallerAccountAndRegion() (string, string) {
	identity, err := s.idClient.Get()
	region := s.sessionRegion
	if err != nil {
		log.Printf("Failed to get caller's Account ID %v", err)
		return "unknown", region
	}
	return identity.Account, region
}

func marshal(e interface{}) (string, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
