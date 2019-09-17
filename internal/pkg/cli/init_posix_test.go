// +build !windows

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/Netflix/go-expect"
	"github.com/aws/PRIVATE-amazon-ecs-archer/internal/pkg/archer"
	cli_mocks "github.com/aws/PRIVATE-amazon-ecs-archer/internal/pkg/cli/mocks"
	"github.com/aws/PRIVATE-amazon-ecs-archer/mocks"
	"github.com/golang/mock/gomock"
	"github.com/hinshun/vt10x"
	"github.com/stretchr/testify/require"
)

func TestInit_Ask(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockProjectStore := mocks.NewMockProjectStore(ctrl)
	defer ctrl.Finish()

	testCases := map[string]struct {
		inputProject     string
		inputApp         string
		inputType        string
		input            func(c *expect.Console)
		wantedProject    string
		wantedApp        string
		wantedType       string
		existingProjects []string
	}{
		"with no flags set and no projects": {
			input: func(c *expect.Console) {
				c.ExpectString("What is your project's name?")
				c.SendLine("heartbeat")
				c.ExpectString("What is your application's name?")
				c.SendLine("api")
				c.ExpectString("Which template would you like to use?")
				c.SendLine(string(terminal.KeyEnter))
				c.ExpectString("Would you like to set up a test environment")
				c.SendLine("n")
				c.ExpectEOF()
			},
			wantedProject: "heartbeat",
			wantedApp:     "api",
			wantedType:    "Load Balanced Web App",
		},
		"with no flags set and existing projects": {
			existingProjects: []string{"heartbeat"},
			input: func(c *expect.Console) {
				c.ExpectString("Which project should we use?")
				c.SendLine(string(terminal.KeyEnter))
				c.ExpectString("What is your application's name?")
				c.SendLine("api")
				c.ExpectString("Which template would you like to use?")
				c.SendLine(string(terminal.KeyEnter))
				c.ExpectString("Would you like to set up a test environment")
				c.SendLine("n")
				c.ExpectEOF()
			},
			wantedProject: "heartbeat",
			wantedApp:     "api",
			wantedType:    "Load Balanced Web App",
		},
		"with only project flag set": {
			inputProject: "heartbeat",
			input: func(c *expect.Console) {
				c.ExpectString("What is your application's name?")
				c.SendLine("api")
				c.ExpectString("Which template would you like to use?")
				c.SendLine(string(terminal.KeyEnter))
				c.ExpectString("Would you like to set up a test environment")
				c.SendLine("n")
				c.ExpectEOF()
			},
			wantedProject: "heartbeat",
			wantedApp:     "api",
			wantedType:    "Load Balanced Web App",
		},
		"with project and app flag set": {
			inputProject: "heartbeat",
			inputApp:     "api",
			input: func(c *expect.Console) {
				c.ExpectString("Which template would you like to use?")
				c.SendLine(string(terminal.KeyEnter))
				c.ExpectString("Would you like to set up a test environment")
				c.SendLine("n")
				c.ExpectEOF()
			},
			wantedProject: "heartbeat",
			wantedApp:     "api",
			wantedType:    "Load Balanced Web App",
		},
		"with project, app and template flag set": {
			inputProject: "heartbeat",
			inputApp:     "api",
			inputType:    "Load Balanced Web App",
			input: func(c *expect.Console) {
				c.ExpectString("Would you like to set up a test environment")
				c.SendLine("n")
				c.ExpectEOF()
			},
			wantedProject: "heartbeat",
			wantedApp:     "api",
			wantedType:    "Load Balanced Web App",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			mockTerminal, _, err := vt10x.NewVT10XConsole()
			require.NoError(t, err)
			defer mockTerminal.Close()
			app := &InitAppOpts{
				Project: tc.inputProject,
				Name:    tc.inputApp,
				Type:    tc.inputType,
				prompt: terminal.Stdio{
					In:  mockTerminal.Tty(),
					Out: mockTerminal.Tty(),
					Err: mockTerminal.Tty(),
				},
				projStore:        mockProjectStore,
				existingProjects: tc.existingProjects,
			}
			// Write inputs to the terminal
			done := make(chan struct{})
			go func() {
				defer close(done)
				tc.input(mockTerminal)
			}()

			// WHEN
			err = app.Ask()

			// Wait until the terminal receives the input
			mockTerminal.Tty().Close()
			<-done

			// THEN
			require.NoError(t, err)
			require.Equal(t, tc.wantedProject, app.Project, "expected project names to match")
			require.Equal(t, tc.wantedApp, app.Name, "expected app names to match")
			require.Equal(t, tc.wantedType, app.Type, "expected template names to match")

		})
	}
}

//TODO this test currently doesn't mock out the manifest writer.
// Since that part will change soon, I don't have tests for the
// manifest writer parts yet.
func TestInit_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectStore := mocks.NewMockProjectStore(ctrl)
	mockEnvStore := mocks.NewMockEnvironmentStore(ctrl)
	mockSpinner := cli_mocks.NewMockspinner(ctrl)
	mockDeployer := mocks.NewMockEnvironmentDeployer(ctrl)

	mockError := fmt.Errorf("error")

	testCases := map[string]struct {
		inputOpts    InitAppOpts
		consoleInput func(c *expect.Console)
		mocking      func()
		want         error
	}{
		"should not prompt to create test environment given existing environments": {
			inputOpts: InitAppOpts{
				Name:             "frontend",
				Project:          "project1",
				Type:             "Empty",
				existingProjects: []string{"project1", "project2"},
			},
			consoleInput: func(c *expect.Console) {
				c.ExpectEOF()
			},
			mocking: func() {
				mockProjectStore.
					EXPECT().
					CreateProject(gomock.Any()).
					Times(0)
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("project1")).
					Return([]*archer.Environment{
						{Name: "test"},
					}, nil).
					Times(1)
			},
			want: nil,
		},
		"should create a new project without a test environment": {
			inputOpts: InitAppOpts{
				Name:             "frontend",
				Project:          "project3",
				Type:             "Empty",
				existingProjects: []string{"project1", "project2"},
			},
			consoleInput: func(c *expect.Console) {
				c.ExpectString("Would you like to set up a test environment?")
				c.SendLine("n")
				c.ExpectEOF()
			},
			mocking: func() {
				mockProjectStore.
					EXPECT().
					CreateProject(gomock.Eq(&archer.Project{Name: "project3"})).
					Return(nil).
					Times(1)
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("project3")).
					Return(make([]*archer.Environment, 0), nil).
					Times(1)
			},
			want: nil,
		},
		"should echo error returned from call to CreateProject": {
			inputOpts: InitAppOpts{
				Project:          "project3",
				existingProjects: []string{"project1", "project2"},
			},
			consoleInput: func(c *expect.Console) {
				c.ExpectEOF()
			},
			mocking: func() {
				mockProjectStore.
					EXPECT().
					CreateProject(gomock.Eq(&archer.Project{Name: "project3"})).
					Return(mockError)
			},
			want: mockError,
		},
		"should echo error returned from call to deployer.DeployEnvironment": {
			inputOpts: InitAppOpts{
				Name:             "frontend",
				Project:          "project3",
				Type:             "Empty",
				existingProjects: []string{"project1", "project2"},
			},
			consoleInput: func(c *expect.Console) {
				c.ExpectString("Would you like to set up a test environment?")
				c.SendLine("Y")
				c.ExpectEOF()
			},
			mocking: func() {
				mockProjectStore.
					EXPECT().
					CreateProject(gomock.Eq(&archer.Project{Name: "project3"})).
					Return(nil).
					Times(1)
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("project3")).
					Return([]*archer.Environment{}, nil).
					Times(1)
				mockSpinner.EXPECT().Start("Deploying env...").Times(1)
				mockDeployer.EXPECT().
					DeployEnvironment(gomock.Eq(archer.Environment{
						Project: "project3",
						Name:    defaultEnvironmentName,
					}), gomock.Eq(true)).
					Return(mockError).
					Times(1)
				mockSpinner.EXPECT().Stop("Error!").Times(1)
			},
			want: mockError,
		},
		"should echo error returned from call to deployer.Wait": {
			inputOpts: InitAppOpts{
				Name:             "frontend",
				Project:          "project3",
				Type:             "Empty",
				existingProjects: []string{"project1", "project2"},
			},
			consoleInput: func(c *expect.Console) {
				c.ExpectString("Would you like to set up a test environment?")
				c.SendLine("Y")
				c.ExpectEOF()
			},
			mocking: func() {
				mockProjectStore.
					EXPECT().
					CreateProject(gomock.Eq(&archer.Project{Name: "project3"})).
					Return(nil).
					Times(1)
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("project3")).
					Return([]*archer.Environment{}, nil).
					Times(1)
				mockSpinner.EXPECT().Start("Deploying env...").Times(1)
				mockDeployer.EXPECT().
					DeployEnvironment(gomock.Eq(archer.Environment{
						Project: "project3",
						Name:    defaultEnvironmentName,
					}), gomock.Eq(true)).
					Return(nil).
					Times(1)
				mockDeployer.EXPECT().
					Wait(gomock.Eq(archer.Environment{
						Project: "project3",
						Name:    defaultEnvironmentName,
					})).
					Return(mockError).
					Times(1)
				mockSpinner.EXPECT().Stop("Error!").Times(1)
			},
			want: mockError,
		},
		"should create a new test environment": {
			inputOpts: InitAppOpts{
				Name:             "frontend",
				Project:          "project3",
				Type:             "Empty",
				existingProjects: []string{"project1", "project2"},
			},
			consoleInput: func(c *expect.Console) {
				c.ExpectString("Would you like to set up a test environment?")
				c.SendLine("Y")
				c.ExpectEOF()
			},
			mocking: func() {
				mockProjectStore.
					EXPECT().
					CreateProject(gomock.Eq(&archer.Project{Name: "project3"})).
					Return(nil).
					Times(1)
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("project3")).
					Return([]*archer.Environment{}, nil).
					Times(1)
				mockSpinner.EXPECT().Start("Deploying env...").Times(1)
				mockDeployer.EXPECT().
					DeployEnvironment(gomock.Eq(archer.Environment{
						Project: "project3",
						Name:    defaultEnvironmentName,
					}), gomock.Eq(true)).
					Return(nil).
					Times(1)
				mockDeployer.EXPECT().
					Wait(gomock.Eq(archer.Environment{
						Project: "project3",
						Name:    defaultEnvironmentName,
					})).
					Return(nil).
					Times(1)
				mockSpinner.EXPECT().Stop("Done!").Times(1)
			},
			want: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockTerminal, _, err := vt10x.NewVT10XConsole()
			require.NoError(t, err)
			defer mockTerminal.Close()
			// Write inputs to the terminal
			done := make(chan struct{})
			go func() {
				defer close(done)
				tc.consoleInput(mockTerminal)
			}()

			tc.mocking()
			tc.inputOpts.prompt = terminal.Stdio{
				In:  mockTerminal.Tty(),
				Out: mockTerminal.Tty(),
				Err: mockTerminal.Tty(),
			}
			tc.inputOpts.projStore = mockProjectStore
			tc.inputOpts.envStore = mockEnvStore
			tc.inputOpts.spinner = mockSpinner
			tc.inputOpts.deployer = mockDeployer

			// WHEN
			got := tc.inputOpts.Execute()

			// THEN
			mockTerminal.Tty().Close()
			<-done
			require.Equal(t, tc.want, got)
		})
	}
}
