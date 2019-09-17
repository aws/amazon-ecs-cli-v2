// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/aws/PRIVATE-amazon-ecs-archer/internal/pkg/archer"
	cli_mocks "github.com/aws/PRIVATE-amazon-ecs-archer/internal/pkg/cli/mocks"
	"github.com/aws/PRIVATE-amazon-ecs-archer/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestEnvAdd_Validate(t *testing.T) {
	testCases := map[string]struct {
		inputOpts       AddEnvOpts
		wantedErrPrefix string
	}{
		"with no project name": {
			inputOpts: AddEnvOpts{
				EnvName: "coolapp",
			},
			wantedErrPrefix: "to add an environment either run the command in your workspace or provide a --project",
		},
		"with valid input": {
			inputOpts: AddEnvOpts{
				EnvName:     "coolapp",
				ProjectName: "coolproject",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.inputOpts.Validate()
			if tc.wantedErrPrefix != "" {
				require.Regexp(t, "^"+tc.wantedErrPrefix+".*", err.Error())
			} else {
				require.NoError(t, err, "There shouldn't have been an error")
			}
		})
	}
}
func TestEnvAdd_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEnvStore := mocks.NewMockEnvironmentStore(ctrl)
	mockDeployer := mocks.NewMockEnvironmentDeployer(ctrl)
	mockSpinner := cli_mocks.NewMockspinner(ctrl)
	var capturedArgument *archer.Environment
	defer ctrl.Finish()

	testCases := map[string]struct {
		addEnvOpts  AddEnvOpts
		expectedEnv archer.Environment
		mocking     func()
	}{
		"with a succesful call to add env": {
			addEnvOpts: AddEnvOpts{
				manager:     mockEnvStore,
				deployer:    mockDeployer,
				ProjectName: "project",
				EnvName:     "env",
				Production:  true,
				spinner:     mockSpinner,
			},
			expectedEnv: archer.Environment{
				Name:    "env",
				Project: "project",
				//TODO update these to real values
				AccountID: "1234",
				Region:    "1234",
				Prod:      true,
			},
			mocking: func() {
				mockEnvStore.
					EXPECT().
					CreateEnvironment(gomock.Any()).
					Do(func(env *archer.Environment) {
						capturedArgument = env
					})
				mockDeployer.EXPECT().DeployEnvironment(gomock.Any(), gomock.Any())
				mockSpinner.EXPECT().Start(gomock.Eq("Deploying env..."))
				// TODO: Assert Wait is called with stack name returned by DeployEnvironment.
				mockDeployer.EXPECT().Wait(gomock.Any())
				mockSpinner.EXPECT().Stop(gomock.Eq("Done!"))
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Setup mocks
			tc.mocking()

			tc.addEnvOpts.Execute()

			require.Equal(t, tc.expectedEnv, *capturedArgument)
		})
	}
}
