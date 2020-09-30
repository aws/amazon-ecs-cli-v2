// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/config"

	"testing"

	"github.com/aws/copilot-cli/internal/pkg/cli/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestDeleteJobOpts_Validate(t *testing.T) {
	mockError := errors.New("some error")

	tests := map[string]struct {
		inAppName  string
		inEnvName  string
		inName     string
		setupMocks func(m *mocks.Mockstore)

		want error
	}{
		"should return errNoAppInWorkspace": {
			setupMocks: func(m *mocks.Mockstore) {},
			inName:     "api",
			want:       errNoAppInWorkspace,
		},
		"with no flag set": {
			inAppName:  "phonetool",
			setupMocks: func(m *mocks.Mockstore) {},
			want:       nil,
		},
		"with all flag set": {
			inAppName: "phonetool",
			inName:    "api",
			setupMocks: func(m *mocks.Mockstore) {
				m.EXPECT().GetJob("phonetool", "api").Times(1).Return(&config.Workload{
					Name: "api",
				}, nil)
			},
			want: nil,
		},
		"with env flag set": {
			inAppName: "phonetool",
			inEnvName: "test",
			setupMocks: func(m *mocks.Mockstore) {
				m.EXPECT().GetEnvironment("phonetool", "test").
					Return(&config.Environment{Name: "test"}, nil)
			},
			want: nil,
		},
		"with unknown environment": {
			inAppName: "phonetool",
			inEnvName: "test",
			setupMocks: func(m *mocks.Mockstore) {
				m.EXPECT().GetEnvironment("phonetool", "test").Return(nil, errors.New("unknown env"))
			},
			want: errors.New("get environment test from config store: unknown env"),
		},
		"should return error if fail to get job name": {
			inAppName: "phonetool",
			inName:    "resizer",
			setupMocks: func(m *mocks.Mockstore) {
				m.EXPECT().GetJob("phonetool", "resizer").Times(1).Return(nil, mockError)
			},
			want: errors.New("some error"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockstore := mocks.NewMockstore(ctrl)

			test.setupMocks(mockstore)

			opts := deleteJobOpts{
				deleteJobVars: deleteJobVars{
					appName: test.inAppName,
					name:    test.inName,
					envName: test.inEnvName,
				},
				store: mockstore,
			}

			err := opts.Validate()

			if test.want != nil {
				require.EqualError(t, err, test.want.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteJobOpts_Ask(t *testing.T) {
	const (
		testAppName = "phonetool"
		testJobName = "resizer"
	)
	mockError := errors.New("mockError")

	tests := map[string]struct {
		skipConfirmation bool
		inName           string
		envName          string

		mockPrompt func(m *mocks.Mockprompter)
		mockSel    func(m *mocks.MockwsSelector)

		wantedName  string
		wantedError error
	}{
		"should ask for job name": {
			inName:           "",
			skipConfirmation: true,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job("Select a job to delete", "").Return(testJobName, nil)
			},
			mockPrompt: func(m *mocks.Mockprompter) {},

			wantedName: testJobName,
		},
		"returns error if no jobs found": {
			inName:           "",
			skipConfirmation: true,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job("Select a job to delete", "").Return("", mockError)
			},
			mockPrompt: func(m *mocks.Mockprompter) {},

			wantedError: fmt.Errorf("select job: %w", mockError),
		},
		"returns error if fail to select job": {
			inName:           "",
			skipConfirmation: true,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job("Select a job to delete", "").Return("", mockError)
			},
			mockPrompt: func(m *mocks.Mockprompter) {},

			wantedError: fmt.Errorf("select job: %w", mockError),
		},
		"should skip confirmation": {
			inName:           testJobName,
			skipConfirmation: true,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job(gomock.Any(), gomock.Any()).Times(0)
			},
			mockPrompt: func(m *mocks.Mockprompter) {},

			wantedName: testJobName,
		},
		"should wrap error returned from prompter confirmation": {
			inName:           testJobName,
			skipConfirmation: false,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job(gomock.Any(), gomock.Any()).Times(0)
			},
			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					fmt.Sprintf(fmtJobDeleteConfirmPrompt, testJobName, testAppName),
					jobDeleteConfirmHelp,
				).Times(1).Return(true, mockError)
			},

			wantedError: fmt.Errorf("job delete confirmation prompt: %w", mockError),
		},
		"should return error if user does not confirm job delete": {
			inName:           testJobName,
			skipConfirmation: false,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job(gomock.Any(), gomock.Any()).Times(0)
			},
			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					fmt.Sprintf(fmtJobDeleteConfirmPrompt, testJobName, testAppName),
					jobDeleteConfirmHelp,
				).Times(1).Return(false, nil)
			},

			wantedError: errJobDeleteCancelled,
		},
		"should return error nil if user confirms job delete": {
			inName:           testJobName,
			skipConfirmation: false,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job(gomock.Any(), gomock.Any()).Times(0)
			},
			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					fmt.Sprintf(fmtJobDeleteConfirmPrompt, testJobName, testAppName),
					jobDeleteConfirmHelp,
				).Times(1).Return(true, nil)
			},

			wantedName: testJobName,
		},
		"should return error nil if user confirms job delete --env": {
			inName:           testJobName,
			envName:          "test",
			skipConfirmation: false,
			mockSel: func(m *mocks.MockwsSelector) {
				m.EXPECT().Job(gomock.Any(), gomock.Any()).Times(0)
			},
			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					fmt.Sprintf(fmtJobDeleteFromEnvConfirmPrompt, testJobName, "test"),
					fmt.Sprintf(jobDeleteFromEnvConfirmHelp, "test"),
				).Times(1).Return(true, nil)
			},

			wantedName: testJobName,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPrompter := mocks.NewMockprompter(ctrl)
			mockSel := mocks.NewMockwsSelector(ctrl)
			test.mockPrompt(mockPrompter)
			test.mockSel(mockSel)

			opts := deleteJobOpts{
				deleteJobVars: deleteJobVars{
					skipConfirmation: test.skipConfirmation,
					appName:          testAppName,
					name:             test.inName,
					envName:          test.envName,
				},
				prompt: mockPrompter,
				sel:    mockSel,
			}

			got := opts.Ask()

			if got != nil {
				require.Equal(t, test.wantedError, got)
			} else {
				require.Equal(t, test.wantedName, opts.name)
			}
		})
	}
}
