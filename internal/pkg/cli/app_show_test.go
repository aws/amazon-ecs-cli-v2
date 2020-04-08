// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	climocks "github.com/aws/amazon-ecs-cli-v2/internal/pkg/cli/mocks"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/describe"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type showAppMocks struct {
	storeSvc  *climocks.MockstoreReader
	prompt    *climocks.Mockprompter
	describer *climocks.MockwebAppDescriber
	ws        *climocks.MockwsAppReader
}

func TestAppShow_Validate(t *testing.T) {
	testCases := map[string]struct {
		inputProject     string
		inputApplication string
		setupMocks       func(mocks showAppMocks)

		wantedError error
	}{
		"valid project name and application name": {
			inputProject:     "my-project",
			inputApplication: "my-app",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
						Name: "my-project",
					}, nil),
					m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
						Name: "my-app",
					}, nil),
				)
			},

			wantedError: nil,
		},
		"invalid project name": {
			inputProject:     "my-project",
			inputApplication: "my-app",

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("some error"),
		},
		"invalid application name": {
			inputProject:     "my-project",
			inputApplication: "my-app",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
						Name: "my-project",
					}, nil),
					m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(nil, errors.New("some error")),
				)
			},

			wantedError: fmt.Errorf("some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStoreReader := climocks.NewMockstoreReader(ctrl)

			mocks := showAppMocks{
				storeSvc: mockStoreReader,
			}

			tc.setupMocks(mocks)

			showApps := &showAppOpts{
				showAppVars: showAppVars{
					appName: tc.inputApplication,
					GlobalOpts: &GlobalOpts{
						projectName: tc.inputProject,
					},
				},
				storeSvc: mockStoreReader,
			}

			// WHEN
			err := showApps.Validate()

			// THEN
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestAppShow_Ask(t *testing.T) {
	testCases := map[string]struct {
		inputProject string
		inputApp     string

		setupMocks func(mocks showAppMocks)

		wantedProject string
		wantedApp     string
		wantedError   error
	}{
		"with all flags": {
			inputProject: "my-project",
			inputApp:     "my-app",
			setupMocks:   func(mocks showAppMocks) {},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   nil,
		},
		"retrieve all app names if fail to retrieve app name from local": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					// askProject
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						{Name: "my-project"},
						{Name: "archer-project"},
					}, nil),
					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("my-project", nil).Times(1),

					// askAppName
					m.ws.EXPECT().AppNames().Return(nil, errors.New("some error")),
					m.storeSvc.EXPECT().ListApplications("my-project").Return([]*archer.Application{
						{Name: "my-app"},
						{Name: "archer-app"},
					}, nil),
					m.prompt.EXPECT().SelectOne(fmt.Sprintf(applicationShowAppNamePrompt, "my-project"), applicationShowAppNameHelpPrompt, []string{"my-app", "archer-app"}).Return("my-app", nil).Times(1),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   nil,
		},
		"retrieve all app names if no app found in local dir": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					// askProject
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						{Name: "my-project"},
						{Name: "archer-project"},
					}, nil),
					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("my-project", nil).Times(1),

					// askAppName
					m.ws.EXPECT().AppNames().Return([]string{}, nil),
					m.storeSvc.EXPECT().ListApplications("my-project").Return([]*archer.Application{
						{Name: "my-app"},
						{Name: "archer-app"},
					}, nil),

					m.prompt.EXPECT().SelectOne(fmt.Sprintf(applicationShowAppNamePrompt, "my-project"), applicationShowAppNameHelpPrompt, []string{"my-app", "archer-app"}).Return("my-app", nil).Times(1),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   nil,
		},
		"retrieve local app names": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					// askProject
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						{Name: "my-project"},
						{Name: "archer-project"},
					}, nil),
					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("my-project", nil).Times(1),

					// askAppName
					m.ws.EXPECT().AppNames().Return([]string{"my-app", "archer-app"}, nil),
					m.prompt.EXPECT().SelectOne(fmt.Sprintf(applicationShowAppNamePrompt, "my-project"), applicationShowAppNameHelpPrompt, []string{"my-app", "archer-app"}).Return("my-app", nil).Times(1),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   nil,
		},
		"skip selecting if only one application found": {
			inputProject: "my-project",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.ws.EXPECT().AppNames().Return(nil, errors.New("some error")),
					m.storeSvc.EXPECT().ListApplications("my-project").Return([]*archer.Application{
						{
							Name: "my-app",
						},
					}, nil),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   nil,
		},
		"returns error when fail to list project": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().ListProjects().Return(nil, errors.New("some error"))
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   fmt.Errorf("list projects: some error"),
		},
		"returns error when no project found": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{}, nil)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   fmt.Errorf("no project found: run `project init` please"),
		},
		"returns error when fail to select project": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					// askProject
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						{Name: "my-project"},
						{Name: "archer-project"},
					}, nil),
					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("", errors.New("some error")).Times(1),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   fmt.Errorf("select projects: some error"),
		},
		"returns error when fail to list applications": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					// askProject
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						{Name: "my-project"},
						{Name: "archer-project"},
					}, nil),
					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("my-project", nil).Times(1),

					// askAskName
					m.ws.EXPECT().AppNames().Return(nil, errors.New("some error")),
					m.storeSvc.EXPECT().ListApplications("my-project").Return(nil, fmt.Errorf("some error")),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   fmt.Errorf("list applications for project my-project: some error"),
		},
		"returns error when fail to select application": {
			inputProject: "",
			inputApp:     "",

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					// askProject
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						{Name: "my-project"},
						{Name: "archer-project"},
					}, nil),
					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("my-project", nil).Times(1),

					// askAppName
					m.ws.EXPECT().AppNames().Return(nil, errors.New("some error")),
					m.storeSvc.EXPECT().ListApplications("my-project").Return([]*archer.Application{
						{Name: "my-app"},
						{Name: "archer-app"},
					}, nil),

					m.prompt.EXPECT().SelectOne(fmt.Sprintf(applicationShowAppNamePrompt, "my-project"), applicationShowAppNameHelpPrompt, []string{"my-app", "archer-app"}).Return("", fmt.Errorf("some error")).Times(1),
				)
			},

			wantedProject: "my-project",
			wantedApp:     "my-app",
			wantedError:   fmt.Errorf("select applications for project my-project: some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStoreReader := climocks.NewMockstoreReader(ctrl)
			mockPrompter := climocks.NewMockprompter(ctrl)
			mockWorkspace := climocks.NewMockwsAppReader(ctrl)

			mocks := showAppMocks{
				storeSvc: mockStoreReader,
				prompt:   mockPrompter,
				ws:       mockWorkspace,
			}

			tc.setupMocks(mocks)

			showApps := &showAppOpts{
				showAppVars: showAppVars{
					appName: tc.inputApp,
					GlobalOpts: &GlobalOpts{
						prompt:      mockPrompter,
						projectName: tc.inputProject,
					},
				},
				storeSvc: mockStoreReader,
				ws:       mockWorkspace,
			}

			// WHEN
			err := showApps.Ask()

			// THEN
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.wantedProject, showApps.ProjectName(), "expected project name to match")
				require.Equal(t, tc.wantedApp, showApps.appName, "expected application name to match")
			}
		})
	}
}

func TestAppShow_Execute(t *testing.T) {
	projectName := "my-project"
	testCases := map[string]struct {
		inputApp              string
		shouldOutputJSON      bool
		shouldOutputResources bool

		setupMocks func(mocks showAppMocks)

		wantedContent string
		wantedError   error
	}{
		"noop if app name is empty": {
			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetApplication(gomock.Any(), gomock.Any()).Times(0)
				m.describer.EXPECT().URI(gomock.Any()).Times(0)
			},
		},
		"prompt for all input for json output": {
			inputApp:              "my-app",
			shouldOutputJSON:      true,
			shouldOutputResources: true,

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
						Name: "my-app",
					}, nil),
					m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
						{Name: "test"},
						{Name: "prod"},
					}, nil),

					m.describer.EXPECT().URI("test").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "frontend",
					}, nil),
					m.describer.EXPECT().ECSParams("test").Return(&describe.WebAppECSParams{
						ContainerPort: "80",
						TaskSize: describe.TaskSize{
							CPU:    "256",
							Memory: "512",
						},
						TaskCount: "1",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "test"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "test",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "test",
						},
					}, nil),
					m.describer.EXPECT().URI("prod").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "backend",
					}, nil),
					m.describer.EXPECT().ECSParams("prod").Return(&describe.WebAppECSParams{
						ContainerPort: "5000",
						TaskSize: describe.TaskSize{
							CPU:    "512",
							Memory: "1024",
						},
						TaskCount: "3",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "prod"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "prod",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "prod",
						},
					}, nil),

					// describer#StackResources
					m.describer.EXPECT().StackResources("test").Return([]*describe.CfnResource{
						{
							Type:       "AWS::EC2::SecurityGroup",
							PhysicalID: "sg-0758ed6b233743530",
						},
					}, nil),
					m.describer.EXPECT().StackResources("prod").Return([]*describe.CfnResource{
						{
							Type:       "AWS::EC2::SecurityGroupIngress",
							PhysicalID: "ContainerSecurityGroupIngressFromPublicALB",
						},
					}, nil),
				)
			},

			wantedContent: "{\"appName\":\"my-app\",\"type\":\"\",\"project\":\"my-project\",\"configurations\":[{\"environment\":\"test\",\"port\":\"80\",\"tasks\":\"1\",\"cpu\":\"256\",\"memory\":\"512\"},{\"environment\":\"prod\",\"port\":\"5000\",\"tasks\":\"3\",\"cpu\":\"512\",\"memory\":\"1024\"}],\"routes\":[{\"environment\":\"test\",\"url\":\"http://my-pr-Publi.us-west-2.elb.amazonaws.com/frontend\"},{\"environment\":\"prod\",\"url\":\"http://my-pr-Publi.us-west-2.elb.amazonaws.com/backend\"}],\"variables\":[{\"environment\":\"prod\",\"name\":\"ECS_CLI_ENVIRONMENT_NAME\",\"value\":\"prod\"},{\"environment\":\"test\",\"name\":\"ECS_CLI_ENVIRONMENT_NAME\",\"value\":\"test\"}],\"resources\":{\"prod\":[{\"type\":\"AWS::EC2::SecurityGroupIngress\",\"physicalID\":\"ContainerSecurityGroupIngressFromPublicALB\"}],\"test\":[{\"type\":\"AWS::EC2::SecurityGroup\",\"physicalID\":\"sg-0758ed6b233743530\"}]}}\n",
		},
		"prompt for all input for human output": {
			inputApp:              "my-app",
			shouldOutputJSON:      false,
			shouldOutputResources: true,

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
						Name: "my-app",
					}, nil),
					m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
						{Name: "test"},
						{Name: "prod"},
					}, nil),

					m.describer.EXPECT().URI("test").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "frontend",
					}, nil),
					m.describer.EXPECT().ECSParams("test").Return(&describe.WebAppECSParams{
						ContainerPort: "80",
						TaskSize: describe.TaskSize{
							CPU:    "256",
							Memory: "512",
						},
						TaskCount: "1",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "test"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "test",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "test",
						},
					}, nil),
					m.describer.EXPECT().URI("prod").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "backend",
					}, nil),
					m.describer.EXPECT().ECSParams("prod").Return(&describe.WebAppECSParams{
						ContainerPort: "5000",
						TaskSize: describe.TaskSize{
							CPU:    "512",
							Memory: "1024",
						},
						TaskCount: "3",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "prod"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "prod",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "prod",
						},
					}, nil),

					m.describer.EXPECT().StackResources("test").Return([]*describe.CfnResource{
						{
							Type:       "AWS::EC2::SecurityGroup",
							PhysicalID: "sg-0758ed6b233743530",
						},
					}, nil),
					m.describer.EXPECT().StackResources("prod").Return([]*describe.CfnResource{
						{
							Type:       "AWS::EC2::SecurityGroupIngress",
							PhysicalID: "ContainerSecurityGroupIngressFromPublicALB",
						},
					}, nil),
				)
			},

			wantedContent: `About

  Project           my-project
  Name              my-app
  Type              

Configurations

  Environment       Tasks               CPU (vCPU)          Memory (MiB)        Port
  test              1                   0.25                512                 80
  prod              3                   0.5                 1024                5000

Routes

  Environment       URL
  test              http://my-pr-Publi.us-west-2.elb.amazonaws.com/frontend
  prod              http://my-pr-Publi.us-west-2.elb.amazonaws.com/backend

Variables

  Name                      Environment         Value
  ECS_CLI_ENVIRONMENT_NAME  prod                prod
  -                         test                test

Resources

  test
    AWS::EC2::SecurityGroup  sg-0758ed6b233743530

  prod
    AWS::EC2::SecurityGroupIngress  ContainerSecurityGroupIngressFromPublicALB
`,
		},
		"returns error if fail to get application": {
			inputApp:              "my-app",
			shouldOutputJSON:      false,
			shouldOutputResources: false,

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("get application: some error"),
		},
		"returns error if fail to list environments": {
			inputApp:              "my-app",
			shouldOutputJSON:      false,
			shouldOutputResources: false,

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
					Name: "my-app",
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("list environments: some error"),
		},
		"returns error if fail to retrieve URI": {
			inputApp:              "my-app",
			shouldOutputJSON:      false,
			shouldOutputResources: false,

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
					Name: "my-app",
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
					{Name: "test"},
					{Name: "prod"},
				}, nil)

				m.describer.EXPECT().URI("test").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("retrieve application URI: some error"),
		},
		"returns error if fail to retrieve deploy info": {
			inputApp:         "my-app",
			shouldOutputJSON: false,

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
					Name: "my-app",
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
					{Name: "test"},
					{Name: "prod"},
				}, nil)

				m.describer.EXPECT().URI("test").Return(&describe.WebAppURI{
					DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
					Path:    "frontend",
				}, nil)
				m.describer.EXPECT().ECSParams("test").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("retrieve application deployment configuration: some error"),
		},
		"returns error if fail to retrieve environment variables": {
			inputApp:         "my-app",
			shouldOutputJSON: false,

			setupMocks: func(m showAppMocks) {
				m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
					Name: "my-app",
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
					{Name: "test"},
					{Name: "prod"},
				}, nil)

				m.describer.EXPECT().URI("test").Return(&describe.WebAppURI{
					DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
					Path:    "frontend",
				}, nil)
				m.describer.EXPECT().ECSParams("test").Return(&describe.WebAppECSParams{
					ContainerPort: "80",
					TaskSize: describe.TaskSize{
						CPU:    "256",
						Memory: "512",
					},
					TaskCount: "1",
				}, nil)
				m.describer.EXPECT().EnvVars(&archer.Environment{Name: "test"}).Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("retrieve environment variables: some error"),
		},
		"returns error if fail to retrieve application resources": {
			inputApp:              "my-app",
			shouldOutputJSON:      false,
			shouldOutputResources: true,

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
						Name: "my-app",
					}, nil),
					m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
						{Name: "test"},
						{Name: "prod"},
					}, nil),

					m.describer.EXPECT().URI("test").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "frontend",
					}, nil),
					m.describer.EXPECT().ECSParams("test").Return(&describe.WebAppECSParams{
						ContainerPort: "80",
						TaskSize: describe.TaskSize{
							CPU:    "256",
							Memory: "512",
						},
						TaskCount: "1",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "test"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "test",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "test",
						},
					}, nil),
					m.describer.EXPECT().URI("prod").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "backend",
					}, nil),
					m.describer.EXPECT().ECSParams("prod").Return(&describe.WebAppECSParams{
						ContainerPort: "5000",
						TaskSize: describe.TaskSize{
							CPU:    "512",
							Memory: "1024",
						},
						TaskCount: "3",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "prod"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "prod",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "prod",
						},
					}, nil),
					m.describer.EXPECT().StackResources("test").Return(nil, errors.New("some error")),
				)
			},

			wantedError: fmt.Errorf("retrieve application resources: some error"),
		},
		"do not return error if fail to retrieve URI because of application not deployed": {
			inputApp:              "my-app",
			shouldOutputJSON:      false,
			shouldOutputResources: true,

			setupMocks: func(m showAppMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().GetApplication("my-project", "my-app").Return(&archer.Application{
						Name: "my-app",
					}, nil),
					m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
						{Name: "test"},
						{Name: "prod"},
					}, nil),

					m.describer.EXPECT().URI("test").Return(nil, fmt.Errorf("describe stack my-project-test-my-app: %w", awserr.New("ValidationError", "Stack with id my-project-test-my-app does not exist", nil))),
					m.describer.EXPECT().URI("prod").Return(&describe.WebAppURI{
						DNSName: "my-pr-Publi.us-west-2.elb.amazonaws.com",
						Path:    "backend",
					}, nil),
					m.describer.EXPECT().ECSParams("test").Times(0),
					m.describer.EXPECT().ECSParams("prod").Return(&describe.WebAppECSParams{
						ContainerPort: "5000",
						TaskSize: describe.TaskSize{
							CPU:    "512",
							Memory: "1024",
						},
						TaskCount: "3",
					}, nil),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "test"}).Times(0),
					m.describer.EXPECT().EnvVars(&archer.Environment{Name: "prod"}).Return([]*describe.WebAppEnvVars{
						&describe.WebAppEnvVars{
							Environment: "prod",
							Name:        "ECS_CLI_ENVIRONMENT_NAME",
							Value:       "prod",
						},
					}, nil),
					m.describer.EXPECT().StackResources("test").Return(nil, fmt.Errorf("describe resources for stack my-project-test-my-app: %w", awserr.New("ValidationError", "Stack with id my-project-test-my-app does not exist", nil))),
					m.describer.EXPECT().StackResources("prod").Return([]*describe.CfnResource{
						{
							Type:       "AWS::EC2::SecurityGroupIngress",
							PhysicalID: "ContainerSecurityGroupIngressFromPublicALB",
						},
					}, nil),
				)
			},

			wantedContent: `About

  Project           my-project
  Name              my-app
  Type              

Configurations

  Environment       Tasks               CPU (vCPU)          Memory (MiB)        Port
  prod              3                   0.5                 1024                5000

Routes

  Environment       URL
  prod              http://my-pr-Publi.us-west-2.elb.amazonaws.com/backend

Variables

  Name                      Environment         Value
  ECS_CLI_ENVIRONMENT_NAME  prod                prod

Resources

  prod
    AWS::EC2::SecurityGroupIngress  ContainerSecurityGroupIngressFromPublicALB
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			b := &bytes.Buffer{}
			mockStoreReader := climocks.NewMockstoreReader(ctrl)
			mockWebAppDescriber := climocks.NewMockwebAppDescriber(ctrl)

			mocks := showAppMocks{
				storeSvc:  mockStoreReader,
				describer: mockWebAppDescriber,
			}

			tc.setupMocks(mocks)

			showApps := &showAppOpts{
				showAppVars: showAppVars{
					appName:               tc.inputApp,
					shouldOutputJSON:      tc.shouldOutputJSON,
					shouldOutputResources: tc.shouldOutputResources,
					GlobalOpts: &GlobalOpts{
						projectName: projectName,
					},
				},
				storeSvc:      mockStoreReader,
				describer:     mockWebAppDescriber,
				initDescriber: func(*showAppOpts) error { return nil },
				w:             b,
			}

			// WHEN
			err := showApps.Execute()

			// THEN
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.wantedContent, b.String(), "expected output content match")
			}
		})
	}
}
