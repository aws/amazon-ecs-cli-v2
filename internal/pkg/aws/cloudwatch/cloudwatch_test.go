// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudwatch

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/copilot-cli/internal/pkg/aws/cloudwatch/mocks"
	rg "github.com/aws/copilot-cli/internal/pkg/aws/resourcegroups"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type cloudWatchMocks struct {
	cw *mocks.Mockapi
	rg *mocks.MockresourceGetter
}

func TestCloudWatch_AlarmsWithTags(t *testing.T) {
	const (
		svcName      = "mockSvc"
		envName      = "mockEnv"
		appName      = "mockApp"
		mockAlarmArn = "arn:aws:cloudwatch:us-west-2:1234567890:alarm:mockAlarmName"
		mockArn1     = mockAlarmArn + "1"
		mockArn2     = mockAlarmArn + "2"
	)
	mockTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+00:00")
	mockError := errors.New("some error")
	testTags := map[string]string{
		"copilot-application": appName,
	}

	testCases := map[string]struct {
		setupMocks func(m cloudWatchMocks)

		wantErr         error
		wantAlarmStatus []AlarmStatus
	}{
		"errors if failed to search resources": {
			setupMocks: func(m cloudWatchMocks) {
				m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return(nil, mockError)
			},

			wantErr: mockError,
		},
		"errors if failed to get alarm names because of invalid ARN": {
			setupMocks: func(m cloudWatchMocks) {
				m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return([]*rg.Resource{{ARN: "badArn"}}, nil)
			},

			wantErr: fmt.Errorf("parse alarm ARN badArn: arn: invalid prefix"),
		},
		"errors if failed to get alarm names because of bad ARN resource": {
			setupMocks: func(m cloudWatchMocks) {
				m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return([]*rg.Resource{{ARN: "arn:aws:cloudwatch:us-west-2:1234567890:alarm:badAlarm:Names"}}, nil)
			},

			wantErr: fmt.Errorf("unknown ARN resource format alarm:badAlarm:Names"),
		},
		"errors if failed to describe CloudWatch alarms": {
			setupMocks: func(m cloudWatchMocks) {
				gomock.InOrder(
					m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return([]*rg.Resource{{ARN: mockAlarmArn}}, nil),
					m.cw.EXPECT().DescribeAlarms(&cloudwatch.DescribeAlarmsInput{
						NextToken:  nil,
						AlarmNames: aws.StringSlice([]string{"mockAlarmName"}),
					}).Return(nil, mockError),
				)
			},

			wantErr: fmt.Errorf("describe CloudWatch alarms: some error"),
		},
		"return if no alarms found": {
			setupMocks: func(m cloudWatchMocks) {
				m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return([]*rg.Resource{}, nil)
			},

			wantAlarmStatus: nil,
		},
		"success": {
			setupMocks: func(m cloudWatchMocks) {
				gomock.InOrder(
					m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return([]*rg.Resource{{ARN: mockAlarmArn}}, nil),
					m.cw.EXPECT().DescribeAlarms(&cloudwatch.DescribeAlarmsInput{
						NextToken:  nil,
						AlarmNames: aws.StringSlice([]string{"mockAlarmName"}),
					}).Return(&cloudwatch.DescribeAlarmsOutput{
						NextToken: nil,
						MetricAlarms: []*cloudwatch.MetricAlarm{
							{
								AlarmArn:              aws.String(mockAlarmArn),
								AlarmName:             aws.String("mockAlarmName"),
								StateReason:           aws.String("mockReason"),
								StateValue:            aws.String("mockState"),
								StateUpdatedTimestamp: &mockTime,
							},
						},
					}, nil),
				)
			},

			wantAlarmStatus: []AlarmStatus{
				{
					Arn:          mockAlarmArn,
					Name:         "mockAlarmName",
					Type:         "Metric",
					Reason:       "mockReason",
					Status:       "mockState",
					UpdatedTimes: mockTime,
				},
			},
		},
		"success with pagination": {
			setupMocks: func(m cloudWatchMocks) {
				gomock.InOrder(
					m.rg.EXPECT().GetResourcesByTags(cloudwatchResourceType, gomock.Eq(testTags)).Return([]*rg.Resource{{ARN: mockArn1}, {ARN: mockArn2}}, nil),
					m.cw.EXPECT().DescribeAlarms(&cloudwatch.DescribeAlarmsInput{
						NextToken:  nil,
						AlarmNames: aws.StringSlice([]string{"mockAlarmName1", "mockAlarmName2"}),
					}).Return(&cloudwatch.DescribeAlarmsOutput{
						NextToken: aws.String("mockNextToken"),
						CompositeAlarms: []*cloudwatch.CompositeAlarm{
							{
								AlarmArn:              aws.String(mockArn1),
								AlarmName:             aws.String("mockAlarmName1"),
								StateReason:           aws.String("mockReason"),
								StateValue:            aws.String("mockState"),
								StateUpdatedTimestamp: &mockTime,
							},
						},
					}, nil),
					m.cw.EXPECT().DescribeAlarms(&cloudwatch.DescribeAlarmsInput{
						NextToken:  aws.String("mockNextToken"),
						AlarmNames: aws.StringSlice([]string{"mockAlarmName1", "mockAlarmName2"}),
					}).Return(&cloudwatch.DescribeAlarmsOutput{
						NextToken: nil,
						MetricAlarms: []*cloudwatch.MetricAlarm{
							{
								AlarmArn:              aws.String(mockArn2),
								AlarmName:             aws.String("mockAlarmName2"),
								StateReason:           aws.String("mockReason"),
								StateValue:            aws.String("mockState"),
								StateUpdatedTimestamp: &mockTime,
							},
						},
					}, nil),
				)
			},

			wantAlarmStatus: []AlarmStatus{
				{
					Arn:          mockArn1,
					Name:         "mockAlarmName1",
					Type:         "Composite",
					Reason:       "mockReason",
					Status:       "mockState",
					UpdatedTimes: mockTime,
				},
				{
					Arn:          mockArn2,
					Name:         "mockAlarmName2",
					Type:         "Metric",
					Reason:       "mockReason",
					Status:       "mockState",
					UpdatedTimes: mockTime,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockcwClient := mocks.NewMockapi(ctrl)
			mockrgClient := mocks.NewMockresourceGetter(ctrl)
			mocks := cloudWatchMocks{
				cw: mockcwClient,
				rg: mockrgClient,
			}

			tc.setupMocks(mocks)

			cwSvc := CloudWatch{
				client:   mockcwClient,
				rgClient: mockrgClient,
			}

			gotAlarmStatus, gotErr := cwSvc.AlarmsWithTags(testTags)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantAlarmStatus, gotAlarmStatus)
			}
		})

	}
}
