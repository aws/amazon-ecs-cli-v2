// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	sdkcloudformation "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/aws/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/aws/ecs"
	"github.com/aws/copilot-cli/internal/pkg/deploy/cloudformation/mocks"
	awsecs "github.com/aws/copilot-cli/internal/pkg/new-sdk-go/ecs"
	"github.com/aws/copilot-cli/internal/pkg/term/progress"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type mockFileWriter struct {
	io.Writer
}

func (m mockFileWriter) Fd() uintptr { return 0 }

func testDeployWorkload_OnCreateChangeSetFailure(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wantedErr := errors.New("some error")
	m := mocks.NewMockcfnClient(ctrl)
	m.EXPECT().Create(gomock.Any()).Return("", wantedErr)
	m.EXPECT().ErrorEvents(gomock.Any()).Return(nil, nil)
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.True(t, errors.Is(err, wantedErr), `expected returned error to be wrapped with "some error"`)
}

func testDeployWorkload_OnUpdateChangeSetFailure(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wantedErr := errors.New("some error")
	m := mocks.NewMockcfnClient(ctrl)
	m.EXPECT().Create(gomock.Any()).Return("", &cloudformation.ErrStackAlreadyExists{})
	m.EXPECT().Update(gomock.Any()).Return("", wantedErr)
	m.EXPECT().ErrorEvents(gomock.Any()).Return(nil, nil)
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.True(t, errors.Is(err, wantedErr), `expected returned error to be wrapped with "some error"`)
}

func testDeployWorkload_OnDescribeChangeSetFailure(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockcfnClient(ctrl)
	m.EXPECT().Create(gomock.Any()).Return("1234", nil)
	m.EXPECT().DescribeChangeSet(gomock.Any(), gomock.Any()).Return(nil, errors.New("DescribeChangeSet error"))
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.EqualError(t, err, "DescribeChangeSet error")
}

func testDeployWorkload_OnTemplateBodyFailure(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockcfnClient(ctrl)
	m.EXPECT().Create(gomock.Any()).Return("1234", nil)
	m.EXPECT().DescribeChangeSet(gomock.Any(), gomock.Any()).Return(&cloudformation.ChangeSetDescription{}, nil)
	m.EXPECT().TemplateBodyFromChangeSet(gomock.Any(), gomock.Any()).Return("", errors.New("TemplateBody error"))
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.EqualError(t, err, "TemplateBody error")
}

func testDeployWorkload_StackStreamerFailureShouldCancelRenderer(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	wantedErr := errors.New("streamer error")
	m := mocks.NewMockcfnClient(ctrl)
	m.EXPECT().Create(gomock.Any()).Return("1234", nil)
	m.EXPECT().DescribeChangeSet(gomock.Any(), gomock.Any()).Return(&cloudformation.ChangeSetDescription{}, nil)
	m.EXPECT().TemplateBodyFromChangeSet(gomock.Any(), gomock.Any()).Return("", nil)
	m.EXPECT().DescribeStackEvents(gomock.Any()).Return(nil, wantedErr)
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.True(t, errors.Is(err, wantedErr), "expected streamer error to be wrapped and returned")
}

func testDeployWorkload_StreamUntilStackCreationFails(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockcfnClient(ctrl)
	m.EXPECT().Create(gomock.Any()).Return("1234", nil)
	m.EXPECT().DescribeChangeSet(gomock.Any(), gomock.Any()).Return(&cloudformation.ChangeSetDescription{}, nil)
	m.EXPECT().TemplateBodyFromChangeSet(gomock.Any(), gomock.Any()).Return("", nil)
	m.EXPECT().DescribeStackEvents(gomock.Any()).Return(&sdkcloudformation.DescribeStackEventsOutput{
		StackEvents: []*sdkcloudformation.StackEvent{
			{
				EventId:            aws.String("2"),
				LogicalResourceId:  aws.String("hello"),
				PhysicalResourceId: aws.String("AWS::CloudFormation::Stack"),
				ResourceStatus:     aws.String("CREATE_FAILED"), // Send failure event for stack.
				Timestamp:          aws.Time(time.Now()),
			},
		},
	}, nil).AnyTimes()
	m.EXPECT().Describe("hello").Return(&cloudformation.StackDescription{
		StackStatus: aws.String("CREATE_FAILED"),
	}, nil)
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.EqualError(t, err, "stack hello did not complete successfully and exited with status CREATE_FAILED")
}

func testDeployWorkload_RenderNewlyCreatedStackWithECSService(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	// GIVEN
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCFN := mocks.NewMockcfnClient(ctrl)
	mockECS := mocks.NewMockecsClient(ctrl)
	deploymentTime := time.Date(2020, time.November, 23, 18, 0, 0, 0, time.UTC)

	mockCFN.EXPECT().Create(gomock.Any()).Return("1234", nil)
	mockCFN.EXPECT().DescribeChangeSet("1234", "hello").Return(&cloudformation.ChangeSetDescription{
		Changes: []*sdkcloudformation.Change{
			{
				ResourceChange: &sdkcloudformation.ResourceChange{
					LogicalResourceId: aws.String("Service"),
					ResourceType:      aws.String("AWS::ECS::Service"),
				},
			},
		},
	}, nil)
	mockCFN.EXPECT().TemplateBodyFromChangeSet("1234", "hello").Return(`
Resources:
  Service:
    Metadata:
      'aws:copilot:description': 'My ECS Service'
    Type: AWS::ECS::Service
`, nil)
	mockCFN.EXPECT().DescribeStackEvents(&sdkcloudformation.DescribeStackEventsInput{
		StackName: aws.String("hello"),
	}).Return(&sdkcloudformation.DescribeStackEventsOutput{
		StackEvents: []*sdkcloudformation.StackEvent{
			{
				EventId:            aws.String("1"),
				LogicalResourceId:  aws.String("Service"),
				PhysicalResourceId: aws.String("arn:aws:ecs:us-west-2:1111:service/cluster/service"),
				ResourceType:       aws.String("AWS::ECS::Service"),
				ResourceStatus:     aws.String("CREATE_IN_PROGRESS"),
				Timestamp:          aws.Time(deploymentTime),
			},
			{
				EventId:            aws.String("2"),
				LogicalResourceId:  aws.String("Service"),
				PhysicalResourceId: aws.String("arn:aws:ecs:us-west-2:1111:service/cluster/service"),
				ResourceType:       aws.String("AWS::ECS::Service"),
				ResourceStatus:     aws.String("CREATE_COMPLETE"),
				Timestamp:          aws.Time(deploymentTime),
			},
			{
				EventId:           aws.String("3"),
				LogicalResourceId: aws.String("hello"),
				ResourceType:      aws.String("AWS::CloudFormation::Stack"),
				ResourceStatus:    aws.String("CREATE_COMPLETE"),
				Timestamp:         aws.Time(deploymentTime),
			},
		},
	}, nil).AnyTimes()
	mockECS.EXPECT().Service("cluster", "service").Return(&ecs.Service{
		Deployments: []*awsecs.Deployment{
			{
				RolloutState:   aws.String("COMPLETED"),
				Status:         aws.String("PRIMARY"),
				TaskDefinition: aws.String("arn:aws:ecs:us-west-2:1111:task-definition/hello:10"),
				UpdatedAt:      aws.Time(deploymentTime),
			},
		},
	}, nil)
	mockCFN.EXPECT().Describe("hello").Return(&cloudformation.StackDescription{
		StackStatus: aws.String("CREATE_COMPLETE"),
	}, nil)
	client := CloudFormation{cfnClient: mockCFN, ecsClient: mockECS}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.NoError(t, err)
	require.Contains(t, buf.String(), "My ECS Service", "resource should be rendered")
	require.Contains(t, buf.String(), "PRIMARY", "Status of the service should be rendered")
	require.Contains(t, buf.String(), "[completed]", "Rollout state of service should be rendered")
}

func testDeployWorkload_RenderNewlyCreatedStackWithAddons(t *testing.T, when func(w progress.FileWriter, cf CloudFormation) error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockcfnClient(ctrl)

	// Mocks for the parent stack.
	m.EXPECT().Create(gomock.Any()).Return("1234", nil)
	m.EXPECT().DescribeChangeSet("1234", "hello").Return(&cloudformation.ChangeSetDescription{
		Changes: []*sdkcloudformation.Change{
			{
				ResourceChange: &sdkcloudformation.ResourceChange{
					LogicalResourceId:  aws.String("Cluster"),
					PhysicalResourceId: aws.String("AWS::ECS::Cluster"),
				},
			},
			{
				ResourceChange: &sdkcloudformation.ResourceChange{
					ChangeSetId:        aws.String("5678"),
					LogicalResourceId:  aws.String("AddonsStack"),
					PhysicalResourceId: aws.String("arn:aws:cloudformation:us-west-2:12345:stack/my-nested-stack/d0a825a0-e4cd-xmpl-b9fb-061c69e99205"),
				},
			},
		},
	}, nil)

	m.EXPECT().TemplateBodyFromChangeSet("1234", "hello").Return(`
Resources:
  Cluster:
    Metadata:
      'aws:copilot:description': 'An ECS cluster'
    Type: AWS::ECS::Cluster
  AddonsStack:
    Metadata:
      'aws:copilot:description': 'An Addons CloudFormation Stack for your additional AWS resources'
    Type: AWS::CloudFormation::Stack
`, nil)

	m.EXPECT().DescribeStackEvents(&sdkcloudformation.DescribeStackEventsInput{
		StackName: aws.String("hello"),
	}).Return(&sdkcloudformation.DescribeStackEventsOutput{
		StackEvents: []*sdkcloudformation.StackEvent{
			{
				EventId:            aws.String("1"),
				LogicalResourceId:  aws.String("Cluster"),
				PhysicalResourceId: aws.String("AWS::ECS::Cluster"),
				ResourceStatus:     aws.String("CREATE_COMPLETE"),
				Timestamp:          aws.Time(time.Now()),
			},
			{
				EventId:            aws.String("2"),
				LogicalResourceId:  aws.String("AddonsStack"),
				PhysicalResourceId: aws.String("AWS::CloudFormation::Stack"),
				ResourceStatus:     aws.String("CREATE_COMPLETE"),
				Timestamp:          aws.Time(time.Now()),
			},
			{
				EventId:            aws.String("3"),
				LogicalResourceId:  aws.String("hello"),
				PhysicalResourceId: aws.String("AWS::CloudFormation::Stack"),
				ResourceStatus:     aws.String("CREATE_COMPLETE"),
				Timestamp:          aws.Time(time.Now()),
			},
		},
	}, nil).AnyTimes()

	m.EXPECT().Describe("hello").Return(&cloudformation.StackDescription{
		StackStatus: aws.String("CREATE_COMPLETE"),
	}, nil)

	// Mocks for the addons stack.
	m.EXPECT().DescribeChangeSet("5678", "my-nested-stack").Return(&cloudformation.ChangeSetDescription{
		Changes: []*sdkcloudformation.Change{
			{
				ResourceChange: &sdkcloudformation.ResourceChange{
					LogicalResourceId:  aws.String("MyTable"),
					PhysicalResourceId: aws.String("AWS::DynamoDB::Table"),
				},
			},
		},
	}, nil)

	m.EXPECT().TemplateBodyFromChangeSet("5678", "my-nested-stack").Return(`
Resources:
  MyTable:
    Metadata:
      'aws:copilot:description': 'A DynamoDB table to store data'
    Type: AWS::DynamoDB::Table`, nil)

	m.EXPECT().DescribeStackEvents(&sdkcloudformation.DescribeStackEventsInput{
		StackName: aws.String("my-nested-stack"),
	}).Return(&sdkcloudformation.DescribeStackEventsOutput{
		StackEvents: []*sdkcloudformation.StackEvent{
			{
				EventId:            aws.String("1"),
				LogicalResourceId:  aws.String("MyTable"),
				PhysicalResourceId: aws.String("AWS::DynamoDB::Table"),
				ResourceStatus:     aws.String("CREATE_COMPLETE"),
				Timestamp:          aws.Time(time.Now()),
			},
			{
				EventId:            aws.String("2"),
				LogicalResourceId:  aws.String("my-nested-stack"),
				PhysicalResourceId: aws.String("AWS::CloudFormation::Stack"),
				ResourceStatus:     aws.String("CREATE_COMPLETE"),
				Timestamp:          aws.Time(time.Now()),
			},
		},
	}, nil).AnyTimes()
	client := CloudFormation{cfnClient: m}
	buf := new(strings.Builder)

	// WHEN
	err := when(mockFileWriter{Writer: buf}, client)

	// THEN
	require.NoError(t, err)
	require.Contains(t, buf.String(), "An ECS cluster")
	require.Contains(t, buf.String(), "An Addons CloudFormation Stack for your additional AWS resources")
	require.Contains(t, buf.String(), "A DynamoDB table to store data")
}
