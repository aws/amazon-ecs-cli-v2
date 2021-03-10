// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ec2

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/copilot-cli/internal/pkg/aws/ec2/mocks"
	"github.com/aws/copilot-cli/internal/pkg/deploy"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

var (
	inAppEnvFilters = []Filter{
		{
			Name:   fmt.Sprintf(TagFilterName, deploy.AppTagKey),
			Values: []string{"my-app"},
		},
		{
			Name:   fmt.Sprintf(TagFilterName, deploy.EnvTagKey),
			Values: []string{"my-env"},
		},
	}

	subnet1 = &ec2.Subnet{
		SubnetId:            aws.String("subnet-1"),
		MapPublicIpOnLaunch: aws.Bool(false),
	}
	subnet2 = &ec2.Subnet{
		SubnetId:            aws.String("subnet-2"),
		MapPublicIpOnLaunch: aws.Bool(true),
	}
	subnet3 = &ec2.Subnet{
		SubnetId:            aws.String("subnet-3"),
		MapPublicIpOnLaunch: aws.Bool(true),
	}
)

func TestEC2_ExtractVPC(t *testing.T) {
	testCases := map[string]struct {
		displayString string
		wantedError   error
		wantedVPC     *VPC
	}{
		"returns error if string is empty": {
			displayString: "",
			wantedError:   fmt.Errorf("extract VPC ID from string: "),
		},
		"returns just the VPC ID if no name present": {
			displayString: "vpc-imagr8vpcstring",
			wantedError:   nil,
			wantedVPC: &VPC{
				ID: "vpc-imagr8vpcstring",
			},
		},
		"returns both the VPC ID and name if both present": {
			displayString: "vpc-imagr8vpcstring (copilot-app-name-env)",
			wantedError:   nil,
			wantedVPC: &VPC{
				ID:   "vpc-imagr8vpcstring",
				Name: "copilot-app-name-env",
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			vpc, err := ExtractVPC(tc.displayString)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedVPC, vpc)
			}
		})
	}
}

func TestEC2_ListVPC(t *testing.T) {
	testCases := map[string]struct {
		mockEC2Client func(m *mocks.Mockapi)

		wantedError error
		wantedVPC   []VPC
	}{
		"fail to describe vpcs": {
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeVpcs(gomock.Any()).Return(nil, errors.New("some error"))
			},
			wantedError: fmt.Errorf("describe VPCs: some error"),
		},
		"success": {
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeVpcs(&ec2.DescribeVpcsInput{}).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []*ec2.Vpc{
						{
							VpcId: aws.String("mockVPCID1"),
						},
					},
					NextToken: aws.String("mockNextToken"),
				}, nil)
				m.EXPECT().DescribeVpcs(&ec2.DescribeVpcsInput{
					NextToken: aws.String("mockNextToken"),
				}).Return(&ec2.DescribeVpcsOutput{
					Vpcs: []*ec2.Vpc{
						{
							VpcId: aws.String("mockVPCID2"),
							Tags: []*ec2.Tag{
								{
									Key:   aws.String("Name"),
									Value: aws.String("mockVPC2Name"),
								},
							},
						},
					},
				}, nil)
			},
			wantedVPC: []VPC{
				{
					ID: "mockVPCID1",
				},
				{
					ID:   "mockVPCID2",
					Name: "mockVPC2Name",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			vpcs, err := ec2Client.ListVPCs()
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedVPC, vpcs)
			}
		})
	}
}

func TestEC2_ListVPCSubnets(t *testing.T) {
	const mockVPCID = "mockVPCID"
	testCases := map[string]struct {
		mockEC2Client func(m *mocks.Mockapi)
		public        bool

		wantedError   error
		wantedSubnets []string
	}{
		"fail to describe subnets": {
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(gomock.Any()).Return(nil, errors.New("error describing subnets"))
			},
			wantedError: fmt.Errorf("describe subnets: error describing subnets"),
		},
		"success": {
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter([]Filter{
						{
							Name:   "vpc-id",
							Values: []string{mockVPCID},
						},
					}),
				}).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []*ec2.Subnet{
						subnet1,
						subnet2,
						subnet3,
					}}, nil)
			},
			wantedSubnets: []string{"subnet-1", "subnet-2", "subnet-3"},
		},
		"success with filtering": {
			public: true,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter([]Filter{
						{
							Name:   "vpc-id",
							Values: []string{mockVPCID},
						},
					}),
				}).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []*ec2.Subnet{
						subnet1,
						subnet2,
						subnet3,
					}}, nil)
			},
			wantedSubnets: []string{"subnet-2", "subnet-3"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			var subnets []string
			var err error
			if tc.public {
				subnets, err = ec2Client.ListVPCSubnets(mockVPCID, FilterForPublicSubnets())
			} else {
				subnets, err = ec2Client.ListVPCSubnets(mockVPCID)
			}
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedSubnets, subnets)
			}
		})
	}
}

func TestEC2_PublicSubnetIDs(t *testing.T) {
	testCases := map[string]struct {
		inFilter []Filter

		mockEC2Client func(m *mocks.Mockapi)

		wantedError error
		wantedARNs  []string
	}{
		"fail to get public subnets": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(nil, errors.New("error describing subnets"))
			},
			wantedError: fmt.Errorf("describe subnets: error describing subnets"),
		},
		"successfully get only public subnets": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []*ec2.Subnet{
						subnet1,
						subnet2,
						subnet3,
					}}, nil)
			},
			wantedARNs: []string{"subnet-2", "subnet-3"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			arns, err := ec2Client.PublicSubnetIDs(tc.inFilter...)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedARNs, arns)
			}
		})
	}
}

func TestEC2_PublicIP(t *testing.T) {
	testCases := map[string]struct {
		inENI string
		mockEC2Client func(m *mocks.Mockapi)

		wantedIP  string
		wantedErr error
	}{
		"failed to describe network interfaces": {
			inENI: "eni-1",
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
					NetworkInterfaceIds: aws.StringSlice([]string{"eni-1"}),
				}).Return(nil, errors.New("some error"))
			},
			wantedErr: errors.New("describe network interface with ENI eni-1: some error"),
		},
		"no association information found": {
			inENI: "eni-1",
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
					NetworkInterfaceIds: aws.StringSlice([]string{"eni-1"}),
				}).Return(&ec2.DescribeNetworkInterfacesOutput{
					NetworkInterfaces: []*ec2.NetworkInterface{
						&ec2.NetworkInterface{},
					},
				}, nil)
			},
			wantedErr: errors.New("no association information found for ENI eni-1"),
		},
		"successfully get public ip": {
			inENI: "eni-1",
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
					NetworkInterfaceIds: aws.StringSlice([]string{"eni-1"}),
				}).Return(&ec2.DescribeNetworkInterfacesOutput{
					NetworkInterfaces: []*ec2.NetworkInterface{
						&ec2.NetworkInterface{
							Association: &ec2.NetworkInterfaceAssociation{
								PublicIp: aws.String("1.2.3"),
							},
						},
					},
				}, nil)
			},
			wantedIP: "1.2.3",
		},
	}

	for name, tc := range testCases{
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			out, err := ec2Client.PublicIP(tc.inENI)
			if tc.wantedErr != nil {
				require.EqualError(t, tc.wantedErr, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedIP, out)
			}
		})
	}
}

func TestEC2_SubnetIDs(t *testing.T) {
	mockNextToken := aws.String("mockNextToken")
	testCases := map[string]struct {
		inFilter []Filter

		mockEC2Client func(m *mocks.Mockapi)

		wantedError error
		wantedARNs  []string
	}{
		"failed to get subnets": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(nil, errors.New("error describing subnets"))
			},
			wantedError: fmt.Errorf("describe subnets: error describing subnets"),
		},
		"successfully get subnets": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []*ec2.Subnet{
						subnet1, subnet2,
					},
				}, nil)
			},
			wantedARNs: []string{"subnet-1", "subnet-2"},
		},
		"successfully get subnets with pagination": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []*ec2.Subnet{
						subnet1, subnet2,
					},
					NextToken: mockNextToken,
				}, nil)
				m.EXPECT().DescribeSubnets(&ec2.DescribeSubnetsInput{
					Filters:   toEC2Filter(inAppEnvFilters),
					NextToken: mockNextToken,
				}).Return(&ec2.DescribeSubnetsOutput{
					Subnets: []*ec2.Subnet{
						subnet3,
					},
				}, nil)
			},
			wantedARNs: []string{"subnet-1", "subnet-2", "subnet-3"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			arns, err := ec2Client.SubnetIDs(tc.inFilter...)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedARNs, arns)
			}
		})
	}
}

func TestEC2_SecurityGroups(t *testing.T) {
	testCases := map[string]struct {
		inFilter []Filter

		mockEC2Client func(m *mocks.Mockapi)

		wantedError error
		wantedARNs  []string
	}{
		"failed to get security groups": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(nil, errors.New("error getting security groups"))
			},

			wantedError: errors.New("describe security groups: error getting security groups"),
		},
		"get security groups success": {
			inFilter: inAppEnvFilters,
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
					Filters: toEC2Filter(inAppEnvFilters),
				}).Return(&ec2.DescribeSecurityGroupsOutput{
					SecurityGroups: []*ec2.SecurityGroup{
						{
							GroupId: aws.String("sg-1"),
						},
						{
							GroupId: aws.String("sg-2"),
						},
					},
				}, nil)
			},

			wantedARNs: []string{"sg-1", "sg-2"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			arns, err := ec2Client.SecurityGroups(inAppEnvFilters...)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedARNs, arns)
			}
		})
	}
}

func TestEC2_HasDNSSupport(t *testing.T) {
	testCases := map[string]struct {
		vpcID string

		mockEC2Client func(m *mocks.Mockapi)

		wantedError   error
		wantedSupport bool
	}{
		"fail to descibe VPC attribute": {
			vpcID: "mockVPCID",
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeVpcAttribute(gomock.Any()).Return(nil, errors.New("some error"))
			},
			wantedError: fmt.Errorf("describe enableDnsSupport attribute for VPC mockVPCID: some error"),
		},
		"success": {
			vpcID: "mockVPCID",
			mockEC2Client: func(m *mocks.Mockapi) {
				m.EXPECT().DescribeVpcAttribute(&ec2.DescribeVpcAttributeInput{
					VpcId:     aws.String("mockVPCID"),
					Attribute: aws.String(ec2.VpcAttributeNameEnableDnsSupport),
				}).Return(&ec2.DescribeVpcAttributeOutput{
					EnableDnsSupport: &ec2.AttributeBooleanValue{
						Value: aws.Bool(true),
					},
				}, nil)
			},
			wantedSupport: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockAPI := mocks.NewMockapi(ctrl)
			tc.mockEC2Client(mockAPI)

			ec2Client := EC2{
				client: mockAPI,
			}

			support, err := ec2Client.HasDNSSupport(tc.vpcID)
			if tc.wantedError != nil {
				require.EqualError(t, tc.wantedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantedSupport, support)
			}
		})
	}
}
