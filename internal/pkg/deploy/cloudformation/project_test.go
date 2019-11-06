// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/deploy/cloudformation/stack"
	"github.com/aws/amazon-ecs-cli-v2/templates"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/gobuffalo/packd"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCreateProjectResources(t *testing.T) {
	mockProject := archer.Project{
		Name:      "testproject",
		AccountID: "1234",
	}
	testCases := map[string]struct {
		mockCFClient func() *mockCloudFormation
		want         error
	}{
		"Infrastructure Roles Stack Fails": {
			mockCFClient: func() *mockCloudFormation {
				return &mockCloudFormation{
					t: t,
					mockCreateChangeSet: func(t *testing.T, in *cloudformation.CreateChangeSetInput) (*cloudformation.CreateChangeSetOutput, error) {
						return nil, fmt.Errorf("error creating stack")
					},
				}
			},
			want: fmt.Errorf("failed to create changeSet for stack testproject-infrastructure-roles: error creating stack"),
		},
		"Infrastructure Roles Stack Already Exists": {
			mockCFClient: func() *mockCloudFormation {
				return &mockCloudFormation{
					t: t,
					mockCreateChangeSet: func(t *testing.T, in *cloudformation.CreateChangeSetInput) (*cloudformation.CreateChangeSetOutput, error) {
						msg := fmt.Sprintf("Stack [%s-%s] already exists and cannot be created again with the changeSet [ecscli-%s]", mockProjectName, mockEnvironmentName, mockChangeSetID)
						return nil, awserr.New("ValidationError", msg, nil)
					},
					mockCreateStackSet: func(t *testing.T, in *cloudformation.CreateStackSetInput) (*cloudformation.CreateStackSetOutput, error) {
						return nil, nil
					},
				}
			},
		},
		"StackSet Already Exists": {
			mockCFClient: func() *mockCloudFormation {
				client := getMockSuccessfulDeployCFClient(t, "stackname")
				client.mockCreateStackSet = func(t *testing.T, in *cloudformation.CreateStackSetInput) (*cloudformation.CreateStackSetOutput, error) {
					return nil, awserr.New(cloudformation.ErrCodeNameAlreadyExistsException, "StackSetAlreadyExiststs", nil)
				}
				return client
			},
		},
		"Infrastructure Roles StackSet Created": {
			mockCFClient: func() *mockCloudFormation {
				client := getMockSuccessfulDeployCFClient(t, "stackname")
				client.mockCreateStackSet = func(t *testing.T, in *cloudformation.CreateStackSetInput) (*cloudformation.CreateStackSetOutput, error) {
					require.Equal(t, "ECS CLI Project Resources (ECR repos, KMS keys, S3 buckets)", *in.Description)
					require.Equal(t, "testproject-infrastructure", *in.StackSetName)
					require.NotZero(t, *in.TemplateBody, "TemplateBody should not be empty")
					require.Equal(t, "testproject-executionrole", *in.ExecutionRoleName)
					require.Equal(t, "arn:aws:iam::1234:role/testproject-adminrole", *in.AdministrationRoleARN)
					require.True(t, len(in.Tags) == 1, "There should be one tag for the project")
					require.Equal(t, "ecs-project", *in.Tags[0].Key)
					require.Equal(t, mockProject.Name, *in.Tags[0].Value)

					return nil, nil
				}
				return client
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cf := CloudFormation{
				client: tc.mockCFClient(),
				box:    templates.Box(),
			}
			got := cf.CreateProjectResources(&mockProject)

			if tc.want != nil {
				require.EqualError(t, tc.want, got.Error())
			} else {
				require.NoError(t, got)
			}
		})
	}
}

func TestAddEnvToProject(t *testing.T) {
	mockProject := archer.Project{
		Name:      "testproject",
		AccountID: "1234",
	}
	testCases := map[string]struct {
		cf      CloudFormation
		project *archer.Project
		env     *archer.Environment
		want    error
	}{
		"with no existing deployments and adding an env": {
			project: &mockProject,
			env:     &archer.Environment{Name: "test", AccountID: "1234", Region: "us-west-2"},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					// Given there hasn't been a StackSet update - the metadata in the stack body will be empty.
					mockDescribeStackSet: func(t *testing.T, in *cloudformation.DescribeStackSetInput) (*cloudformation.DescribeStackSetOutput, error) {
						body, err := yaml.Marshal(stack.DeployedProjectMetadata{})
						require.NoError(t, err)
						return &cloudformation.DescribeStackSetOutput{
							StackSet: &cloudformation.StackSet{
								TemplateBody: aws.String(string(body)),
							},
						}, nil
					},
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						require.Equal(t, "ECS CLI Project Resources (ECR repos, KMS keys, S3 buckets)", *in.Description)
						require.Equal(t, "testproject-infrastructure", *in.StackSetName)
						require.Equal(t, "testproject-executionrole", *in.ExecutionRoleName)
						require.Equal(t, "arn:aws:iam::1234:role/testproject-adminrole", *in.AdministrationRoleARN)
						require.True(t, len(in.Tags) == 1, "There should be one tag for the project")
						require.Equal(t, "ecs-project", *in.Tags[0].Key)
						require.Equal(t, mockProject.Name, *in.Tags[0].Value)

						require.Equal(t, "1", *in.OperationId)

						require.NotZero(t, *in.TemplateBody, "TemplateBody should not be empty")
						configToDeploy, err := stack.ProjectConfigFrom(in.TemplateBody)
						require.NoError(t, err)
						require.ElementsMatch(t, []string{mockProject.AccountID}, configToDeploy.Accounts)
						require.Empty(t, configToDeploy.Apps, "There should be no new apps to deploy")
						require.Equal(t, 1, configToDeploy.Version)
						return &cloudformation.UpdateStackSetOutput{
							OperationId: aws.String("1"),
						}, nil
					},
					mockListStackInstances: func(t *testing.T, in *cloudformation.ListStackInstancesInput) (*cloudformation.ListStackInstancesOutput, error) {
						return &cloudformation.ListStackInstancesOutput{
							Summaries: []*cloudformation.StackInstanceSummary{},
						}, nil
					},
					mockCreateStackInstances: func(t *testing.T, in *cloudformation.CreateStackInstancesInput) (*cloudformation.CreateStackInstancesOutput, error) {
						require.ElementsMatch(t, []*string{aws.String(mockProject.AccountID)}, in.Accounts)
						require.ElementsMatch(t, []*string{aws.String("us-west-2")}, in.Regions)
						require.Equal(t, "testproject-infrastructure", *in.StackSetName)
						return &cloudformation.CreateStackInstancesOutput{
							OperationId: aws.String("1"),
						}, nil
					},
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("SUCCEEDED"),
							},
						}, nil
					},
				},
				box: templates.Box(),
			},
		},

		"with no new account ID added": {
			project: &mockProject,
			env:     &archer.Environment{Name: "test", AccountID: "1234", Region: "us-west-2"},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					// Given there hasn't been a StackSet update - the metadata in the stack body will be empty.
					mockDescribeStackSet: func(t *testing.T, in *cloudformation.DescribeStackSetInput) (*cloudformation.DescribeStackSetOutput, error) {
						body, err := yaml.Marshal(stack.DeployedProjectMetadata{
							Metadata: stack.ProjectResourcesConfig{
								Accounts: []string{"1234"},
							},
						})
						require.NoError(t, err)
						return &cloudformation.DescribeStackSetOutput{
							StackSet: &cloudformation.StackSet{
								TemplateBody: aws.String(string(body)),
							},
						}, nil
					},
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						require.Equal(t, "ECS CLI Project Resources (ECR repos, KMS keys, S3 buckets)", *in.Description)
						require.Equal(t, "testproject-infrastructure", *in.StackSetName)
						require.Equal(t, "testproject-executionrole", *in.ExecutionRoleName)
						require.Equal(t, "arn:aws:iam::1234:role/testproject-adminrole", *in.AdministrationRoleARN)
						require.True(t, len(in.Tags) == 1, "There should be one tag for the project")
						require.Equal(t, "ecs-project", *in.Tags[0].Key)
						require.Equal(t, mockProject.Name, *in.Tags[0].Value)

						require.Equal(t, "1", *in.OperationId)

						require.NotZero(t, *in.TemplateBody, "TemplateBody should not be empty")
						configToDeploy, err := stack.ProjectConfigFrom(in.TemplateBody)
						require.NoError(t, err)
						// Ensure there are no duplicate accounts
						require.ElementsMatch(t, []string{mockProject.AccountID}, configToDeploy.Accounts)
						require.Empty(t, configToDeploy.Apps, "There should be no new apps to deploy")
						require.Equal(t, 1, configToDeploy.Version)
						return &cloudformation.UpdateStackSetOutput{
							OperationId: aws.String("1"),
						}, nil
					},
					mockListStackInstances: func(t *testing.T, in *cloudformation.ListStackInstancesInput) (*cloudformation.ListStackInstancesOutput, error) {
						return &cloudformation.ListStackInstancesOutput{
							Summaries: []*cloudformation.StackInstanceSummary{},
						}, nil
					},
					mockCreateStackInstances: func(t *testing.T, in *cloudformation.CreateStackInstancesInput) (*cloudformation.CreateStackInstancesOutput, error) {
						require.ElementsMatch(t, []*string{aws.String(mockProject.AccountID)}, in.Accounts)
						require.ElementsMatch(t, []*string{aws.String("us-west-2")}, in.Regions)
						require.Equal(t, "testproject-infrastructure", *in.StackSetName)
						return &cloudformation.CreateStackInstancesOutput{
							OperationId: aws.String("1"),
						}, nil
					},

					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("SUCCEEDED"),
							},
						}, nil
					},
				},
				box: templates.Box(),
			},
		},

		"with existing stack instances in same region but different account (no new stack instances, but update stackset)": {
			project: &mockProject,
			env:     &archer.Environment{Name: "test", AccountID: "1234", Region: "us-west-2"},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					// Given there hasn't been a StackSet update - the metadata in the stack body will be empty.
					mockDescribeStackSet: func(t *testing.T, in *cloudformation.DescribeStackSetInput) (*cloudformation.DescribeStackSetOutput, error) {
						body, err := yaml.Marshal(stack.DeployedProjectMetadata{Metadata: stack.ProjectResourcesConfig{
							Accounts: []string{"4567"},
							Version:  1,
						}})
						require.NoError(t, err)
						return &cloudformation.DescribeStackSetOutput{
							StackSet: &cloudformation.StackSet{
								TemplateBody: aws.String(string(body)),
							},
						}, nil
					},
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						require.NotZero(t, *in.TemplateBody, "TemplateBody should not be empty")
						configToDeploy, err := stack.ProjectConfigFrom(in.TemplateBody)
						require.NoError(t, err)
						require.ElementsMatch(t, []string{mockProject.AccountID, "4567"}, configToDeploy.Accounts)
						require.Empty(t, configToDeploy.Apps, "There should be no new apps to deploy")
						require.Equal(t, 2, configToDeploy.Version)

						return &cloudformation.UpdateStackSetOutput{
							OperationId: aws.String("2"),
						}, nil
					},
					mockListStackInstances: func(t *testing.T, in *cloudformation.ListStackInstancesInput) (*cloudformation.ListStackInstancesOutput, error) {
						return &cloudformation.ListStackInstancesOutput{
							Summaries: []*cloudformation.StackInstanceSummary{
								&cloudformation.StackInstanceSummary{
									Region:  aws.String("us-west-2"),
									Account: aws.String(mockProject.AccountID),
								},
							},
						}, nil
					},
					mockCreateStackInstances: func(t *testing.T, in *cloudformation.CreateStackInstancesInput) (*cloudformation.CreateStackInstancesOutput, error) {
						t.FailNow()
						return nil, nil
					},
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("SUCCEEDED"),
							},
						}, nil
					},
				},
				box: templates.Box(),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.cf.AddEnvToProject(tc.project, tc.env)

			if tc.want != nil {
				require.EqualError(t, got, tc.want.Error())
			} else {
				require.NoError(t, got)
			}
		})
	}
}

func TestAddAppToProject(t *testing.T) {
	mockProject := archer.Project{
		Name:      "testproject",
		AccountID: "1234",
	}
	testCases := map[string]struct {
		cf      CloudFormation
		project *archer.Project
		app     *archer.Application
		want    error
	}{
		"with no existing deployments and adding an app": {
			project: &mockProject,
			app:     &archer.Application{Name: "TestApp"},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					// Given there hasn't been a StackSet update - the metadata in the stack body will be empty.
					mockDescribeStackSet: func(t *testing.T, in *cloudformation.DescribeStackSetInput) (*cloudformation.DescribeStackSetOutput, error) {
						body, err := yaml.Marshal(stack.DeployedProjectMetadata{})
						require.NoError(t, err)
						return &cloudformation.DescribeStackSetOutput{
							StackSet: &cloudformation.StackSet{
								TemplateBody: aws.String(string(body)),
							},
						}, nil
					},
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						require.Equal(t, "ECS CLI Project Resources (ECR repos, KMS keys, S3 buckets)", *in.Description)
						require.Equal(t, "testproject-infrastructure", *in.StackSetName)
						require.Equal(t, "testproject-executionrole", *in.ExecutionRoleName)
						require.Equal(t, "arn:aws:iam::1234:role/testproject-adminrole", *in.AdministrationRoleARN)
						require.True(t, len(in.Tags) == 1, "There should be one tag for the project")
						require.Equal(t, "ecs-project", *in.Tags[0].Key)
						require.Equal(t, mockProject.Name, *in.Tags[0].Value)
						// We should increment the version
						require.Equal(t, "1", *in.OperationId)

						require.NotZero(t, *in.TemplateBody, "TemplateBody should not be empty")
						configToDeploy, err := stack.ProjectConfigFrom(in.TemplateBody)
						require.NoError(t, err)
						require.ElementsMatch(t, []string{"TestApp"}, configToDeploy.Apps)
						require.Empty(t, configToDeploy.Accounts, "There should be no new accounts to deploy")
						require.Equal(t, 1, configToDeploy.Version)
						return &cloudformation.UpdateStackSetOutput{
							OperationId: aws.String("1"),
						}, nil
					},
					mockListStackInstances: func(t *testing.T, in *cloudformation.ListStackInstancesInput) (*cloudformation.ListStackInstancesOutput, error) {
						return &cloudformation.ListStackInstancesOutput{
							Summaries: []*cloudformation.StackInstanceSummary{},
						}, nil
					},
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("SUCCEEDED"),
							},
						}, nil
					},
				},
				box: templates.Box(),
			},
		},
		"with new app to existing project with existing apps": {
			project: &mockProject,
			app:     &archer.Application{Name: "test"},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					// Given there hasn't been a StackSet update - the metadata in the stack body will be empty.
					mockDescribeStackSet: func(t *testing.T, in *cloudformation.DescribeStackSetInput) (*cloudformation.DescribeStackSetOutput, error) {
						body, err := yaml.Marshal(stack.DeployedProjectMetadata{Metadata: stack.ProjectResourcesConfig{
							Apps:    []string{"firsttest"},
							Version: 1,
						}})
						require.NoError(t, err)
						return &cloudformation.DescribeStackSetOutput{
							StackSet: &cloudformation.StackSet{
								TemplateBody: aws.String(string(body)),
							},
						}, nil
					},
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						require.NotZero(t, *in.TemplateBody, "TemplateBody should not be empty")
						configToDeploy, err := stack.ProjectConfigFrom(in.TemplateBody)
						require.NoError(t, err)
						require.ElementsMatch(t, []string{"test", "firsttest"}, configToDeploy.Apps)
						require.Empty(t, configToDeploy.Accounts, "There should be no new apps to deploy")
						require.Equal(t, 2, configToDeploy.Version)

						return &cloudformation.UpdateStackSetOutput{
							OperationId: aws.String("2"),
						}, nil
					},
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("SUCCEEDED"),
							},
						}, nil
					},
				},
				box: templates.Box(),
			},
		},
		"with ewxisting app to existing project with existing apps": {
			project: &mockProject,
			app:     &archer.Application{Name: "test"},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					// Given there hasn't been a StackSet update - the metadata in the stack body will be empty.
					mockDescribeStackSet: func(t *testing.T, in *cloudformation.DescribeStackSetInput) (*cloudformation.DescribeStackSetOutput, error) {
						body, err := yaml.Marshal(stack.DeployedProjectMetadata{Metadata: stack.ProjectResourcesConfig{
							Apps:    []string{"test"},
							Version: 1,
						}})
						require.NoError(t, err)
						return &cloudformation.DescribeStackSetOutput{
							StackSet: &cloudformation.StackSet{
								TemplateBody: aws.String(string(body)),
							},
						}, nil
					},
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						t.FailNow()
						return nil, nil
					},
				},
				box: templates.Box(),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.cf.AddAppToProject(tc.project, tc.app)

			if tc.want != nil {
				require.EqualError(t, got, tc.want.Error())
			} else {
				require.NoError(t, got)
			}
		})
	}
}

func TestWaitForStackSetOperation(t *testing.T) {
	waitingForOperation := true
	testCases := map[string]struct {
		cf   CloudFormation
		want error
	}{
		"operation succeeded": {
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("SUCCEEDED"),
							},
						}, nil
					},
				},
				box: boxWithTemplateFile(),
			},
		},
		"operation failed": {
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("FAILED"),
							},
						}, nil
					},
				},
				box: boxWithTemplateFile(),
			},
			want: fmt.Errorf("project operation operation in stack set stackset failed"),
		},
		"operation stopped": {
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String("STOPPED"),
							},
						}, nil
					},
				},
				box: boxWithTemplateFile(),
			},
			want: fmt.Errorf("project operation operation in stack set stackset was manually stopped"),
		},
		"operation non-terminal to succeeded": {
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockDescribeStackSetOperation: func(t *testing.T, in *cloudformation.DescribeStackSetOperationInput) (*cloudformation.DescribeStackSetOperationOutput, error) {
						// First, say the status is running. Then during the next call, set the status to succeeded.
						status := "RUNNING"
						if !waitingForOperation {
							status = "SUCCEEDED"
						}
						waitingForOperation = false
						return &cloudformation.DescribeStackSetOperationOutput{
							StackSetOperation: &cloudformation.StackSetOperation{
								Status: aws.String(status),
							},
						}, nil
					},
				},
				box: boxWithTemplateFile(),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.cf.waitForStackSetOperation("stackset", "operation")

			if tc.want != nil {
				require.EqualError(t, got, tc.want.Error())
			} else {
				require.NoError(t, got)
			}
		})
	}
}

func TestDeployProjectConfig_ErrWrapping(t *testing.T) {
	mockProject := archer.Project{Name: "project", AccountID: "12345"}

	testCases := map[string]struct {
		cf   CloudFormation
		want error
	}{
		"ErrCodeOperationIdAlreadyExistsException": {
			want: &ErrStackSetOutOfDate{projectName: mockProject.Name},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						return nil, awserr.New(cloudformation.ErrCodeOperationIdAlreadyExistsException, "operation already exists", nil)
					},
				},
				box: boxWithTemplateFile(),
			},
		},
		"ErrCodeOperationInProgressException": {
			want: &ErrStackSetOutOfDate{projectName: mockProject.Name},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						return nil, awserr.New(cloudformation.ErrCodeOperationInProgressException, "something is in progres", nil)
					},
				},
				box: boxWithTemplateFile(),
			},
		},
		"ErrCodeStaleRequestException": {
			want: &ErrStackSetOutOfDate{projectName: mockProject.Name},
			cf: CloudFormation{
				client: &mockCloudFormation{
					t: t,
					mockUpdateStackSet: func(t *testing.T, in *cloudformation.UpdateStackSetInput) (*cloudformation.UpdateStackSetOutput, error) {
						return nil, awserr.New(cloudformation.ErrCodeStaleRequestException, "something is stale", nil)
					},
				},
				box: boxWithTemplateFile(),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockProjectResources := stack.ProjectResourcesConfig{}
			got := tc.cf.deployProjectConfig(stack.NewProjectStackConfig(&mockProject, boxWithProjectTemplate()), &mockProjectResources)
			require.NotNil(t, got)
			require.True(t, errors.Is(tc.want, got), "Got %v but expected %v", got, tc.want)
		})
	}
}

// Useful for mocking a successfully deployed stack
func getMockSuccessfulDeployCFClient(t *testing.T, stackName string) *mockCloudFormation {
	return &mockCloudFormation{
		t: t,
		mockCreateChangeSet: func(t *testing.T, in *cloudformation.CreateChangeSetInput) (*cloudformation.CreateChangeSetOutput, error) {
			return &cloudformation.CreateChangeSetOutput{
				Id:      aws.String("changesetID"),
				StackId: aws.String(stackName),
			}, nil
		},
		mockWaitUntilChangeSetCreateComplete: func(t *testing.T, in *cloudformation.DescribeChangeSetInput) error {

			return nil
		},
		mockWaitUntilStackCreateComplete: func(t *testing.T, input *cloudformation.DescribeStacksInput) error {
			return nil
		},
		mockDescribeChangeSet: func(t *testing.T, in *cloudformation.DescribeChangeSetInput) (*cloudformation.DescribeChangeSetOutput, error) {
			return &cloudformation.DescribeChangeSetOutput{
				ExecutionStatus: aws.String(cloudformation.ExecutionStatusAvailable),
			}, nil
		},
		mockExecuteChangeSet: func(t *testing.T, in *cloudformation.ExecuteChangeSetInput) (output *cloudformation.ExecuteChangeSetOutput, e error) {
			return nil, nil
		},
		mockDescribeStacks: func(t *testing.T, input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
			return &cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					&cloudformation.Stack{
						StackId: aws.String(fmt.Sprintf("arn:aws:cloudformation:eu-west-3:902697171733:stack/%s", stackName)),
					},
				},
			}, nil
		},
	}
}

func boxWithProjectTemplate() packd.Box {
	box := packd.NewMemoryBox()

	box.AddString("project/cf.yml", mockTemplate)

	return box
}
