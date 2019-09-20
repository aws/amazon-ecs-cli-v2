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
	"github.com/aws/PRIVATE-amazon-ecs-archer/mocks"
	"github.com/golang/mock/gomock"
	"github.com/hinshun/vt10x"
	"github.com/stretchr/testify/require"
)

func TestEnvList_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEnvStore := mocks.NewMockEnvironmentStore(ctrl)
	defer ctrl.Finish()

	testCases := map[string]struct {
		listOpts ListEnvOpts
		output   func(c *expect.Console) bool
		mocking  func()
	}{
		"with envs": {
			output: func(c *expect.Console) bool {
				c.ExpectString("test")
				c.ExpectString("test2")
				return true
			},
			listOpts: ListEnvOpts{
				ProjectName: "coolproject",
				manager:     mockEnvStore,
			},
			mocking: func() {
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("coolproject")).
					Return([]*archer.Environment{
						&archer.Environment{Name: "test"},
						&archer.Environment{Name: "test2"},
					}, nil)

			},
		},
		"with invalid project name": {
			output: func(c *expect.Console) bool {
				c.ExpectString("error calling SSM")
				return true
			},
			listOpts: ListEnvOpts{
				ProjectName: "coolproject",
				manager:     mockEnvStore,
			},
			mocking: func() {
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("coolproject")).
					Return(nil, fmt.Errorf("error calling SSM"))

			},
		},
		"with production envs": {
			output: func(c *expect.Console) bool {
				c.ExpectString("test")
				c.ExpectString("test2 (prod)")
				return true
			},
			listOpts: ListEnvOpts{
				ProjectName: "coolproject",
				manager:     mockEnvStore,
			},
			mocking: func() {
				mockEnvStore.
					EXPECT().
					ListEnvironments(gomock.Eq("coolproject")).
					Return([]*archer.Environment{
						&archer.Environment{Name: "test"},
						&archer.Environment{Name: "test2", Prod: true},
					}, nil)

			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			mockTerminal, _, err := vt10x.NewVT10XConsole()
			require.NoError(t, err)
			defer mockTerminal.Close()
			// Prepare mocks
			tc.mocking()

			// Set up fake terminal
			tc.listOpts.prompt = terminal.Stdio{
				In:  mockTerminal.Tty(),
				Out: mockTerminal.Tty(),
				Err: mockTerminal.Tty(),
			}

			// Write inputs to the terminal
			done := make(chan bool)
			go func() { done <- tc.output(mockTerminal) }()

			// WHEN
			tc.listOpts.Execute()
			require.True(t, <-done, "We should print to the terminal")

			// Cleanup our terminals
			mockTerminal.Tty().Close()
			mockTerminal.Close()
		})
	}
}
