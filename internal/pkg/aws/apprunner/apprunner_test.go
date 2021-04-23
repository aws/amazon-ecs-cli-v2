// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package apprunner

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/aws/apprunner/mocks"
	"github.com/aws/copilot-cli/internal/pkg/new-sdk-go/apprunner"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAppRunner_DescribeService(t *testing.T) {
	mockTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+00:00")

	testCases := map[string]struct {
		serviceArn          string
		mockAppRunnerClient func(m *mocks.Mockapi)

		wantErr error
		wantSvc Service
	}{
		"success": {
			serviceArn: "mock-svc-arn",
			mockAppRunnerClient: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeService(&apprunner.DescribeServiceInput{
					ServiceArn: aws.String("mock-svc-arn"),
				}).Return(&apprunner.DescribeServiceOutput{
					Service: &apprunner.Service{
						ServiceArn:  aws.String("111111111111.apprunner.us-east-1.amazonaws.com/service/testsvc/test-svc-id"),
						ServiceId:   aws.String("test-svc-id"),
						ServiceName: aws.String("testapp-testenv-testsvc"),
						ServiceUrl:  aws.String("tumkjmvjif.public.us-east-1.apprunner.aws.dev"),
						Status:      aws.String("RUNNING"),
						CreatedAt:   &mockTime,
						UpdatedAt:   &mockTime,
						InstanceConfiguration: &apprunner.InstanceConfiguration{
							Cpu:    aws.String("1024"),
							Memory: aws.String("2048"),
						},
						SourceConfiguration: &apprunner.SourceConfiguration{
							ImageRepository: &apprunner.ImageRepository{
								ImageIdentifier: aws.String("111111111111.dkr.ecr.us-east-1.amazonaws.com/testapp/testsvc:8cdef9a"),
								ImageConfiguration: &apprunner.ImageConfiguration{
									RuntimeEnvironmentVariables: aws.StringMap(map[string]string{
										"LOG_LEVEL":                "info",
										"COPILOT_APPLICATION_NAME": "testapp",
									}),
									Port: aws.String("80"),
								},
							},
						},
					},
				}, nil)
			},
			wantSvc: Service{
				ServiceARN:  "111111111111.apprunner.us-east-1.amazonaws.com/service/testsvc/test-svc-id",
				Name:        "testapp-testenv-testsvc",
				ID:          "test-svc-id",
				Status:      "RUNNING",
				ServiceURL:  "tumkjmvjif.public.us-east-1.apprunner.aws.dev",
				DateCreated: mockTime,
				DateUpdated: mockTime,
				EnvironmentVariables: []*EnvironmentVariable{
					{
						Name:  "COPILOT_APPLICATION_NAME",
						Value: "testapp",
					},
					{
						Name:  "LOG_LEVEL",
						Value: "info",
					},
				},
				CPU:    "1024",
				Memory: "2048",
				Port:   "80",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppRunnerClient := mocks.NewMockapi(ctrl)
			tc.mockAppRunnerClient(mockAppRunnerClient)

			service := AppRunner{
				client: mockAppRunnerClient,
			}

			gotSvc, gotErr := service.DescribeService(tc.serviceArn)

			if gotErr != nil {
				require.EqualError(t, tc.wantErr, gotErr.Error())
			} else {
				require.Equal(t, tc.wantSvc, *gotSvc)
			}
		})
	}
}

func TestAppRunner_ServiceARN(t *testing.T) {
	const (
		mockSvc    = "mockSvc"
		mockSvcARN = "mockSvcArn"
	)
	testError := errors.New("some error")
	testCases := map[string]struct {
		mockAppRunnerClient func(m *mocks.Mockapi)

		wantErr    error
		wantSvcARN string
	}{
		"success": {
			mockAppRunnerClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListServices(&apprunner.ListServicesInput{}).Return(&apprunner.ListServicesOutput{
					ServiceSummaryList: []*apprunner.ServiceSummary{
						{
							ServiceName: aws.String("mockSvc"),
							ServiceArn:  aws.String("mockSvcArn"),
						},
						{
							ServiceName: aws.String("mockSvc2"),
							ServiceArn:  aws.String("mockSvcArn2"),
						},
					},
				}, nil)
			},
			wantSvcARN: mockSvcARN,
		},
		"errors if fail to get services": {
			mockAppRunnerClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListServices(&apprunner.ListServicesInput{}).Return(nil, testError)
			},
			wantErr: fmt.Errorf("list AppRunner services: some error"),
		},
		"errors if no service found": {
			mockAppRunnerClient: func(m *mocks.Mockapi) {
				m.EXPECT().ListServices(&apprunner.ListServicesInput{}).Return(&apprunner.ListServicesOutput{
					ServiceSummaryList: []*apprunner.ServiceSummary{
						{
							ServiceName: aws.String("mockSvc2"),
							ServiceArn:  aws.String("mockSvcArn2"),
						},
					},
				}, nil)
			},
			wantErr: fmt.Errorf("no AppRunner service found for mockSvc"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppRunnerClient := mocks.NewMockapi(ctrl)
			tc.mockAppRunnerClient(mockAppRunnerClient)

			service := AppRunner{
				client: mockAppRunnerClient,
			}

			svcArn, err := service.ServiceARN("mockSvc")

			if err != nil {
				require.EqualError(t, tc.wantErr, err.Error())
			} else {
				require.Equal(t, tc.wantSvcARN, svcArn)
			}
		})
	}
}

func Test_ParseServiceName(t *testing.T) {
	testCases := map[string]struct {
		svcARN string

		wantErr     error
		wantSvcName string
	}{
		"invalid ARN": {
			svcARN:  "mockBadSvcARN",
			wantErr: fmt.Errorf("arn: invalid prefix"),
		},
		"empty svc ARN": {
			svcARN:  "",
			wantErr: fmt.Errorf("arn: invalid prefix"),
		},
		"successfully parse name from service ARN": {
			svcARN:      "arn:aws:apprunner:us-west-2:1234567890:service/my-service/svc-id",
			wantSvcName: "my-service",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			svcName, err := ParseServiceName(tc.svcARN)

			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantSvcName, svcName)
			}
		})
	}
}

func Test_ParseServiceID(t *testing.T) {
	testCases := map[string]struct {
		svcARN string

		wantErr   error
		wantSvcID string
	}{
		"invalid ARN": {
			svcARN:  "mockBadSvcARN",
			wantErr: fmt.Errorf("arn: invalid prefix"),
		},
		"empty svc ARN": {
			svcARN:  "",
			wantErr: fmt.Errorf("arn: invalid prefix"),
		},
		"successfully parse ID from service ARN": {
			svcARN:    "arn:aws:apprunner:us-west-2:1234567890:service/my-service/svc-id",
			wantSvcID: "svc-id",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			svcID, err := ParseServiceID(tc.svcARN)

			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantSvcID, svcID)
			}
		})
	}
}

func Test_LogGroupName(t *testing.T) {
	testCases := map[string]struct {
		svcARN string

		wantErr          error
		wantLogGroupName string
	}{
		"errors if ARN is invalid": {
			svcARN:  "this is not a valid arn",
			wantErr: fmt.Errorf("get service name: arn: invalid prefix"),
		},
		"errors if ARN is missing ID": {
			svcARN:  "arn:aws:apprunner:us-west-2:1234567890:service/my-service",
			wantErr: fmt.Errorf("get service name: cannot parse resource for ARN arn:aws:apprunner:us-west-2:1234567890:service/my-service"),
		},
		"success": {
			svcARN:           "arn:aws:apprunner:us-west-2:1234567890:service/my-service/svc-id",
			wantLogGroupName: "/aws/apprunner/my-service/svc-id/application",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logGroupName, err := LogGroupName(tc.svcARN)

			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.Equal(t, nil, err)
				require.Equal(t, tc.wantLogGroupName, logGroupName)
			}
		})
	}
}

func Test_SystemLogGroupName(t *testing.T) {
	testCases := map[string]struct {
		svcARN string

		wantErr          error
		wantLogGroupName string
	}{
		"errors if ARN is invalid": {
			svcARN:  "this is not a valid arn",
			wantErr: fmt.Errorf("get service name: arn: invalid prefix"),
		},
		"errors if ARN is missing ID": {
			svcARN:  "arn:aws:apprunner:us-west-2:1234567890:service/my-service",
			wantErr: fmt.Errorf("get service name: cannot parse resource for ARN arn:aws:apprunner:us-west-2:1234567890:service/my-service"),
		},
		"success": {
			svcARN:           "arn:aws:apprunner:us-west-2:1234567890:service/my-service/svc-id",
			wantLogGroupName: "/aws/apprunner/my-service/svc-id/service",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			logGroupName, err := SystemLogGroupName(tc.svcARN)

			if tc.wantErr != nil {
				require.EqualError(t, err, tc.wantErr.Error())
			} else {
				require.Equal(t, nil, err)
				require.Equal(t, tc.wantLogGroupName, logGroupName)
			}
		})
	}
}
