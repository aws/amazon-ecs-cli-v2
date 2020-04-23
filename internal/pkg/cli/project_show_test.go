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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type showProjectMocks struct {
	storeSvc *climocks.MockstoreReader
	prompt   *climocks.Mockprompter
}

func TestProjectShow_Validate(t *testing.T) {
	testCases := map[string]struct {
		inputProject string
		setupMocks   func(mocks showProjectMocks)

		wantedError error
	}{
		"valid project name and application name": {
			inputProject: "my-project",

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
					Name: "my-project",
				}, nil)
			},

			wantedError: nil,
		},
		"invalid project name": {
			inputProject: "my-project",

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStoreReader := climocks.NewMockstoreReader(ctrl)
			mockPrompter := climocks.NewMockprompter(ctrl)

			mocks := showProjectMocks{
				storeSvc: mockStoreReader,
				prompt:   mockPrompter,
			}
			tc.setupMocks(mocks)

			showProjects := &showProjectOpts{
				showProjectVars: showProjectVars{
					GlobalOpts: &GlobalOpts{
						prompt:      mockPrompter,
						projectName: tc.inputProject,
					},
				},
				storeSvc: mockStoreReader,
			}

			// WHEN
			err := showProjects.Validate()

			// THEN
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestProjectShow_Ask(t *testing.T) {
	testCases := map[string]struct {
		inputProject string

		setupMocks func(mocks showProjectMocks)

		wantedProject string
		wantedError   error
	}{
		"with all flags": {
			inputProject: "my-project",

			setupMocks: func(m showProjectMocks) {},

			wantedProject: "my-project",
			wantedError:   nil,
		},
		"prompt for all input": {
			inputProject: "",

			setupMocks: func(m showProjectMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						&archer.Project{Name: "my-project"},
						&archer.Project{Name: "archer-project"},
					}, nil),

					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("my-project", nil).Times(1),
				)
			},
			wantedProject: "my-project",
			wantedError:   nil,
		},
		"returns error if fail to list project": {
			inputProject: "",

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().ListProjects().Return(nil, errors.New("some error"))
			},
			wantedError: fmt.Errorf("list project: some error"),
		},
		"returns error if no project found": {
			inputProject: "",

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{}, nil)
			},

			wantedError: fmt.Errorf("no project found: run `project init` to set one up, or `cd` into your workspace please"),
		},
		"returns error if fail to select project": {
			inputProject: "",

			setupMocks: func(m showProjectMocks) {
				gomock.InOrder(
					m.storeSvc.EXPECT().ListProjects().Return([]*archer.Project{
						&archer.Project{Name: "my-project"},
						&archer.Project{Name: "archer-project"},
					}, nil),

					m.prompt.EXPECT().SelectOne(applicationShowProjectNamePrompt, applicationShowProjectNameHelpPrompt, []string{"my-project", "archer-project"}).Return("", errors.New("some error")).Times(1),
				)
			},

			wantedError: fmt.Errorf("select project: some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStoreReader := climocks.NewMockstoreReader(ctrl)
			mockPrompter := climocks.NewMockprompter(ctrl)

			mocks := showProjectMocks{
				storeSvc: mockStoreReader,
				prompt:   mockPrompter,
			}
			tc.setupMocks(mocks)

			showProjects := &showProjectOpts{
				showProjectVars: showProjectVars{
					GlobalOpts: &GlobalOpts{
						prompt:      mockPrompter,
						projectName: tc.inputProject,
					},
				},
				storeSvc: mockStoreReader,
			}

			// WHEN
			err := showProjects.Ask()

			// THEN
			if tc.wantedError != nil {
				require.EqualError(t, err, tc.wantedError.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.wantedProject, showProjects.ProjectName(), "expected project name to match")

			}
		})
	}
}

func TestProjectShow_Execute(t *testing.T) {
	projectName := "my-project"
	testCases := map[string]struct {
		shouldOutputJSON bool

		setupMocks func(mocks showProjectMocks)

		wantedContent string
		wantedError   error
	}{
		"correctly shows json output": {
			shouldOutputJSON: true,

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
					Name:   "my-project",
					Domain: "example.com",
				}, nil)
				m.storeSvc.EXPECT().ListApplications("my-project").Return([]*archer.Application{
					&archer.Application{
						Name: "my-app",
						Type: "lb-web-app",
					},
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
					&archer.Environment{
						Name:      "test",
						Region:    "us-west-2",
						AccountID: "123456789",
						Prod:      false,
					},
					&archer.Environment{
						Name:      "prod",
						AccountID: "123456789",
						Region:    "us-west-1",
						Prod:      true,
					},
				}, nil)
			},

			wantedContent: "{\"name\":\"my-project\",\"uri\":\"example.com\",\"environments\":[{\"project\":\"\",\"name\":\"test\",\"region\":\"us-west-2\",\"accountID\":\"123456789\",\"prod\":false,\"registryURL\":\"\",\"executionRoleARN\":\"\",\"managerRoleARN\":\"\"},{\"project\":\"\",\"name\":\"prod\",\"region\":\"us-west-1\",\"accountID\":\"123456789\",\"prod\":true,\"registryURL\":\"\",\"executionRoleARN\":\"\",\"managerRoleARN\":\"\"}],\"applications\":[{\"project\":\"\",\"name\":\"my-app\",\"type\":\"lb-web-app\"}]}\n",
		},
		"correctly shows human output": {
			shouldOutputJSON: false,

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
					Name:   "my-project",
					Domain: "example.com",
				}, nil)
				m.storeSvc.EXPECT().ListApplications("my-project").Return([]*archer.Application{
					&archer.Application{
						Name: "my-app",
						Type: "lb-web-app",
					},
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
					&archer.Environment{
						Name:      "test",
						Region:    "us-west-2",
						AccountID: "123456789",
					},
					&archer.Environment{
						Name:      "prod",
						AccountID: "123456789",
						Region:    "us-west-1",
					},
				}, nil)
			},

			wantedContent: `About

  Name              my-project
  URI               example.com

Environments

  Name              AccountID           Region
  test              123456789           us-west-2
  prod              123456789           us-west-1

Applications

  Name              Type
  my-app            lb-web-app
`,
		},
		"returns error if fail to get project": {
			shouldOutputJSON: false,

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("get project: some error"),
		},
		"returns error if fail to list environment": {
			shouldOutputJSON: false,

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
					Name:   "my-project",
					Domain: "example.com",
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("list environment: some error"),
		},
		"returns error if fail to list application": {
			shouldOutputJSON: false,

			setupMocks: func(m showProjectMocks) {
				m.storeSvc.EXPECT().GetProject("my-project").Return(&archer.Project{
					Name:   "my-project",
					Domain: "example.com",
				}, nil)
				m.storeSvc.EXPECT().ListEnvironments("my-project").Return([]*archer.Environment{
					&archer.Environment{
						Name:      "test",
						Region:    "us-west-2",
						AccountID: "123456789",
					},
					&archer.Environment{
						Name:      "prod",
						AccountID: "123456789",
						Region:    "us-west-1",
					},
				}, nil)
				m.storeSvc.EXPECT().ListApplications("my-project").Return(nil, errors.New("some error"))
			},

			wantedError: fmt.Errorf("list application: some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			b := &bytes.Buffer{}
			mockStoreReader := climocks.NewMockstoreReader(ctrl)

			mocks := showProjectMocks{
				storeSvc: mockStoreReader,
			}
			tc.setupMocks(mocks)

			showProjects := &showProjectOpts{
				showProjectVars: showProjectVars{
					shouldOutputJSON: tc.shouldOutputJSON,
					GlobalOpts: &GlobalOpts{
						projectName: projectName,
					},
				},
				storeSvc: mockStoreReader,
				w:        b,
			}

			// WHEN
			err := showProjects.Execute()

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
