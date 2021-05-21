// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/aws/copilot-cli/internal/pkg/exec"
	"github.com/aws/copilot-cli/internal/pkg/repository/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRepository_BuildAndPush(t *testing.T) {
	inRepoName := "my-repo"
	inDockerfilePath := "path/to/dockerfile"

	mockTag1, mockTag2, mockTag3 := "tag1", "tag2", "tag3"
	mockRepoURI := "mockURI"

	defaultDockerArguments := exec.BuildArguments{
		URI:        mockRepoURI,
		Dockerfile: inDockerfilePath,
		Context:    filepath.Dir(inDockerfilePath),
		Tags:       []string{mockTag1, mockTag2, mockTag3},
	}

	testCases := map[string]struct {
		inRepoName       string
		inDockerfilePath string
		inMockDocker     func(m *mocks.MockContainerLoginBuildPusher)

		mockRegistry func(m *mocks.MockRegistry)

		wantedError  error
		wantedDigest string
	}{
		"failed to get auth": {
			mockRegistry: func(m *mocks.MockRegistry) {
				m.EXPECT().Auth().Return("", "", errors.New("error getting auth"))
			},
			inMockDocker: func(m *mocks.MockContainerLoginBuildPusher) {
				m.EXPECT().Build(gomock.Any()).AnyTimes()
				m.EXPECT().GetHelperProviderFromDockerCfg().Return("")
				m.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantedError: errors.New("get auth: error getting auth"),
		},
		"failed to build image": {
			mockRegistry: func(m *mocks.MockRegistry) {
				m.EXPECT().Auth().Return("", "", nil).AnyTimes()
			},
			inMockDocker: func(m *mocks.MockContainerLoginBuildPusher) {
				m.EXPECT().Build(&defaultDockerArguments).Return(errors.New("error building image"))
				m.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
				m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantedError: fmt.Errorf("build Dockerfile at %s: error building image", inDockerfilePath),
		},
		"failed to login": {
			mockRegistry: func(m *mocks.MockRegistry) {
				m.EXPECT().Auth().Return("my-name", "my-pwd", nil)
			},
			inMockDocker: func(m *mocks.MockContainerLoginBuildPusher) {
				m.EXPECT().Build(gomock.Any()).AnyTimes()
				m.EXPECT().GetHelperProviderFromDockerCfg().Return("")
				m.EXPECT().Login(mockRepoURI, "my-name", "my-pwd").Return(errors.New("error logging in"))
				m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			wantedError: fmt.Errorf("login to repo %s: error logging in", inRepoName),
		},
		"failed to push": {
			mockRegistry: func(m *mocks.MockRegistry) {
				m.EXPECT().Auth().Times(1)
			},
			inMockDocker: func(m *mocks.MockContainerLoginBuildPusher) {
				m.EXPECT().Build(&defaultDockerArguments).Times(1)
				m.EXPECT().GetHelperProviderFromDockerCfg().Return("")
				m.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
				m.EXPECT().Push(mockRepoURI, mockTag1, mockTag2, mockTag3).Return("", errors.New("error pushing image"))
			},
			wantedError: errors.New("push to repo my-repo: error pushing image"),
		},
		"push with ecr-login": {
			inMockDocker: func(m *mocks.MockContainerLoginBuildPusher) {
				m.EXPECT().Build(&defaultDockerArguments).Return(nil).Times(1)
				m.EXPECT().GetHelperProviderFromDockerCfg().Return("ecr-login")
				m.EXPECT().Push(mockRepoURI, mockTag1, mockTag2, mockTag3).Return("sha256:f1d4ae3f7261a72e98c6ebefe9985cf10a0ea5bd762585a43e0700ed99863807", nil)
			},
			wantedDigest: "sha256:f1d4ae3f7261a72e98c6ebefe9985cf10a0ea5bd762585a43e0700ed99863807",
		},
		"success": {
			mockRegistry: func(m *mocks.MockRegistry) {
				m.EXPECT().Auth().Return("my-name", "my-pwd", nil).Times(1)
			},
			inMockDocker: func(m *mocks.MockContainerLoginBuildPusher) {
				m.EXPECT().Build(&defaultDockerArguments).Return(nil).Times(1)
				m.EXPECT().GetHelperProviderFromDockerCfg().Return("")
				m.EXPECT().Login(mockRepoURI, "my-name", "my-pwd").Return(nil).Times(1)
				m.EXPECT().Push(mockRepoURI, mockTag1, mockTag2, mockTag3).Return("sha256:f1d4ae3f7261a72e98c6ebefe9985cf10a0ea5bd762585a43e0700ed99863807", nil)
			},
			wantedDigest: "sha256:f1d4ae3f7261a72e98c6ebefe9985cf10a0ea5bd762585a43e0700ed99863807",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepoGetter := mocks.NewMockRegistry(ctrl)
			mockDocker := mocks.NewMockContainerLoginBuildPusher(ctrl)

			if tc.mockRegistry != nil {
				tc.mockRegistry(mockRepoGetter)
			}
			if tc.inMockDocker != nil {
				tc.inMockDocker(mockDocker)
			}

			repo := &Repository{
				name:     inRepoName,
				registry: mockRepoGetter,

				uri: mockRepoURI,
			}

			digest, err := repo.BuildAndPush(mockDocker, &exec.BuildArguments{
				Dockerfile: inDockerfilePath,
				Context:    filepath.Dir(inDockerfilePath),
				Tags:       []string{mockTag1, mockTag2, mockTag3},
			})
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedDigest, digest)
			}
		})
	}
}
