// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ecs

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/copilot-cli/internal/pkg/aws/ecs/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestECS_TaskDefinition(t *testing.T) {
	mockError := errors.New("error")

	testCases := map[string]struct {
		taskDefinitionName string
		mockECSClient      func(m *mocks.Mockapi)

		wantErr     error
		wantTaskDef *TaskDefinition
	}{
		"should return wrapped error given error": {
			taskDefinitionName: "task-def",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String("task-def"),
				}).Return(nil, mockError)
			},
			wantErr: fmt.Errorf("describe task definition %s: %w", "task-def", mockError),
		},
		"returns task definition given a task definition name": {
			taskDefinitionName: "task-def",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
					TaskDefinition: aws.String("task-def"),
				}).Return(&ecs.DescribeTaskDefinitionOutput{
					TaskDefinition: &ecs.TaskDefinition{
						ContainerDefinitions: []*ecs.ContainerDefinition{
							{
								Environment: []*ecs.KeyValuePair{
									{
										Name:  aws.String("COPILOT_SERVICE_NAME"),
										Value: aws.String("my-app"),
									},
									{
										Name:  aws.String("COPILOT_ENVIRONMENT_NAME"),
										Value: aws.String("prod"),
									},
								},
							},
						},
					},
				}, nil)
			},
			wantTaskDef: &TaskDefinition{
				ContainerDefinitions: []*ecs.ContainerDefinition{
					{
						Environment: []*ecs.KeyValuePair{
							{
								Name:  aws.String("COPILOT_SERVICE_NAME"),
								Value: aws.String("my-app"),
							},
							{
								Name:  aws.String("COPILOT_ENVIRONMENT_NAME"),
								Value: aws.String("prod"),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}

			gotTaskDef, gotErr := service.TaskDefinition(tc.taskDefinitionName)

			if gotErr != nil {
				require.Equal(t, tc.wantErr, gotErr)
			} else {
				require.Equal(t, tc.wantTaskDef, gotTaskDef)
			}
		})

	}
}

func TestECS_Service(t *testing.T) {
	testCases := map[string]struct {
		clusterName   string
		serviceName   string
		mockECSClient func(m *mocks.Mockapi)

		wantErr error
		wantSvc *Service
	}{
		"success": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  aws.String("mockCluster"),
					Services: aws.StringSlice([]string{"mockService"}),
				}).Return(&ecs.DescribeServicesOutput{
					Services: []*ecs.Service{
						{
							ServiceName: aws.String("mockService"),
						},
					},
				}, nil)
			},
			wantSvc: &Service{
				ServiceName: aws.String("mockService"),
			},
		},
		"errors if failed to describe service": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  aws.String("mockCluster"),
					Services: aws.StringSlice([]string{"mockService"}),
				}).Return(nil, errors.New("some error"))
			},
			wantErr: fmt.Errorf("describe service mockService: some error"),
		},
		"errors if failed to find the service": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeServices(&ecs.DescribeServicesInput{
					Cluster:  aws.String("mockCluster"),
					Services: aws.StringSlice([]string{"mockService"}),
				}).Return(&ecs.DescribeServicesOutput{
					Services: []*ecs.Service{
						{
							ServiceName: aws.String("badMockService"),
						},
					},
				}, nil)
			},
			wantErr: fmt.Errorf("cannot find service mockService"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}

			gotSvc, gotErr := service.Service(tc.clusterName, tc.serviceName)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantSvc, gotSvc)
			}
		})

	}
}

func TestECS_Tasks(t *testing.T) {
	testCases := map[string]struct {
		clusterName   string
		serviceName   string
		mockECSClient func(m *mocks.Mockapi)

		wantErr   error
		wantTasks []*Task
	}{
		"errors if failed to list running tasks": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(nil, errors.New("some error"))
			},
			wantErr: fmt.Errorf("list running tasks of service mockService: some error"),
		},
		"errors if failed to describe running tasks": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: nil,
					TaskArns:  aws.StringSlice([]string{"mockTaskArn"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn"}),
				}).Return(nil, errors.New("some error"))
			},
			wantErr: fmt.Errorf("describe running tasks in cluster mockCluster: some error"),
		},
		"success": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: nil,
					TaskArns:  aws.StringSlice([]string{"mockTaskArn"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("mockTaskArn"),
						},
					},
				}, nil)
			},
			wantTasks: []*Task{
				{
					TaskArn: aws.String("mockTaskArn"),
				},
			},
		},
		"success with pagination": {
			clusterName: "mockCluster",
			serviceName: "mockService",
			mockECSClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: aws.String("mockNextToken"),
					TaskArns:  aws.StringSlice([]string{"mockTaskArn1"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn1"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("mockTaskArn1"),
						},
					},
				}, nil)
				m.EXPECT().ListTasks(&ecs.ListTasksInput{
					Cluster:     aws.String("mockCluster"),
					ServiceName: aws.String("mockService"),
					NextToken:   aws.String("mockNextToken"),
				}).Return(&ecs.ListTasksOutput{
					NextToken: nil,
					TaskArns:  aws.StringSlice([]string{"mockTaskArn2"}),
				}, nil)
				m.EXPECT().DescribeTasks(&ecs.DescribeTasksInput{
					Cluster: aws.String("mockCluster"),
					Tasks:   aws.StringSlice([]string{"mockTaskArn2"}),
				}).Return(&ecs.DescribeTasksOutput{
					Tasks: []*ecs.Task{
						{
							TaskArn: aws.String("mockTaskArn2"),
						},
					},
				}, nil)
			},
			wantTasks: []*Task{
				{
					TaskArn: aws.String("mockTaskArn1"),
				},
				{
					TaskArn: aws.String("mockTaskArn2"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockECSClient := mocks.NewMockapi(ctrl)
			tc.mockECSClient(mockECSClient)

			service := ECS{
				client: mockECSClient,
			}

			gotTasks, gotErr := service.ServiceTasks(tc.clusterName, tc.serviceName)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantTasks, gotTasks)
			}
		})

	}
}

func TestTaskDefinition_EnvVars(t *testing.T) {
	testCases := map[string]struct {
		inContainers []*ecs.ContainerDefinition

		wantEnvVars map[string]string
	}{
		"should return wrapped error given error": {
			inContainers: []*ecs.ContainerDefinition{
				{
					Environment: []*ecs.KeyValuePair{
						{
							Name:  aws.String("COPILOT_SERVICE_NAME"),
							Value: aws.String("my-svc"),
						},
						{
							Name:  aws.String("COPILOT_ENVIRONMENT_NAME"),
							Value: aws.String("prod"),
						},
					},
				},
			},

			wantEnvVars: map[string]string{
				"COPILOT_SERVICE_NAME":     "my-svc",
				"COPILOT_ENVIRONMENT_NAME": "prod",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			taskDefinition := TaskDefinition{
				ContainerDefinitions: tc.inContainers,
			}

			gotEnvVars := taskDefinition.EnvironmentVariables()

			require.Equal(t, tc.wantEnvVars, gotEnvVars)
		})

	}
}

func TestTask_TaskStatus(t *testing.T) {
	startTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+00:00")
	stopTime, _ := time.Parse(time.RFC3339, "2006-01-02T16:04:05+00:00")
	mockImageDigest := "18f7eb6cff6e63e5f5273fb53f672975fe6044580f66c354f55d2de8dd28aec7"
	testCases := map[string]struct {
		health        *string
		taskArn       *string
		containers    []*ecs.Container
		lastStatus    *string
		startedAt     time.Time
		stoppedAt     time.Time
		stoppedReason *string

		wantTaskStatus *TaskStatus
		wantErr        error
	}{
		"errors if failed to parse task ID": {
			taskArn: aws.String("badTaskArn"),
			wantErr: fmt.Errorf("arn: invalid prefix"),
		},
		"success with a provisioning task": {
			taskArn: aws.String("arn:aws:ecs:us-west-2:123456789:task/my-project-test-Cluster-9F7Y0RLP60R7/4082490ee6c245e09d2145010aa1ba8d"),
			containers: []*ecs.Container{
				{
					Image:       aws.String("mockImageArn"),
					ImageDigest: aws.String("sha256:" + mockImageDigest),
				},
			},
			health:     aws.String("HEALTHY"),
			lastStatus: aws.String("UNKNOWN"),

			wantTaskStatus: &TaskStatus{
				Health: "HEALTHY",
				ID:     "4082490ee6c245e09d2145010aa1ba8d",
				Images: []Image{
					{
						Digest: mockImageDigest,
						ID:     "mockImageArn",
					},
				},
				LastStatus: "UNKNOWN",
			},
		},
		"success with a running task": {
			taskArn: aws.String("arn:aws:ecs:us-west-2:123456789:task/my-project-test-Cluster-9F7Y0RLP60R7/4082490ee6c245e09d2145010aa1ba8d"),
			containers: []*ecs.Container{
				{
					Image:       aws.String("mockImageArn"),
					ImageDigest: aws.String("sha256:" + mockImageDigest),
				},
			},
			health:     aws.String("HEALTHY"),
			lastStatus: aws.String("UNKNOWN"),
			startedAt:  startTime,

			wantTaskStatus: &TaskStatus{
				Health: "HEALTHY",
				ID:     "4082490ee6c245e09d2145010aa1ba8d",
				Images: []Image{
					{
						Digest: mockImageDigest,
						ID:     "mockImageArn",
					},
				},
				LastStatus: "UNKNOWN",
				StartedAt:  startTime,
			},
		},
		"success with a stopped task": {
			taskArn: aws.String("arn:aws:ecs:us-west-2:123456789:task/my-project-test-Cluster-9F7Y0RLP60R7/4082490ee6c245e09d2145010aa1ba8d"),
			containers: []*ecs.Container{
				{
					Image:       aws.String("mockImageArn"),
					ImageDigest: aws.String("sha256:" + mockImageDigest),
				},
			},
			health:        aws.String("HEALTHY"),
			lastStatus:    aws.String("UNKNOWN"),
			startedAt:     startTime,
			stoppedAt:     stopTime,
			stoppedReason: aws.String("some reason"),

			wantTaskStatus: &TaskStatus{
				Health: "HEALTHY",
				ID:     "4082490ee6c245e09d2145010aa1ba8d",
				Images: []Image{
					{
						Digest: mockImageDigest,
						ID:     "mockImageArn",
					},
				},
				LastStatus:    "UNKNOWN",
				StartedAt:     startTime,
				StoppedAt:     stopTime,
				StoppedReason: "some reason",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			task := Task{
				HealthStatus:  tc.health,
				TaskArn:       tc.taskArn,
				Containers:    tc.containers,
				LastStatus:    tc.lastStatus,
				StartedAt:     &tc.startedAt,
				StoppedAt:     &tc.stoppedAt,
				StoppedReason: tc.stoppedReason,
			}

			gotTaskStatus, gotErr := task.TaskStatus()

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantTaskStatus, gotTaskStatus)
			}
		})

	}
}

func TestTaskStatus_HumanString(t *testing.T) {
	startTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+00:00")
	stopTime, _ := time.Parse(time.RFC3339, "2006-01-02T16:04:05+00:00")
	mockImageDigest := "18f7eb6cff6e63e5f5273fb53f672975fe6044580f66c354f55d2de8dd28aec7"
	testCases := map[string]struct {
		id          string
		health      string
		lastStatus  string
		imageDigest string
		startedAt   time.Time
		stoppedAt   time.Time

		wantTaskStatus string
	}{
		"all params": {
			health:      "HEALTHY",
			id:          "aslhfnqo39j8qomimvoiqm89349",
			lastStatus:  "RUNNING",
			startedAt:   startTime,
			stoppedAt:   stopTime,
			imageDigest: mockImageDigest,

			wantTaskStatus: "  aslhfnqo\t18f7eb6c\tRUNNING\tHEALTHY\t14 years ago\t14 years ago\n",
		},
		"missing params": {
			health:     "HEALTHY",
			lastStatus: "RUNNING",

			wantTaskStatus: "  -\t-\tRUNNING\tHEALTHY\t-\t-\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			task := TaskStatus{
				Health: tc.health,
				ID:     tc.id,
				Images: []Image{
					{
						Digest: tc.imageDigest,
					},
				},
				LastStatus: tc.lastStatus,
				StartedAt:  tc.startedAt,
				StoppedAt:  tc.stoppedAt,
			}

			gotTaskStatus := task.HumanString()

			require.Equal(t, tc.wantTaskStatus, gotTaskStatus)
		})

	}
}
