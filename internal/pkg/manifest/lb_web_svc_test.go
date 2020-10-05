// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/template"
	"github.com/aws/copilot-cli/internal/pkg/template/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestLoadBalancedWebService_MarshalBinary(t *testing.T) {
	testCases := map[string]struct {
		mockDependencies func(ctrl *gomock.Controller, manifest *LoadBalancedWebService)

		wantedBinary []byte
		wantedError  error
	}{
		"error parsing template": {
			mockDependencies: func(ctrl *gomock.Controller, manifest *LoadBalancedWebService) {
				m := mocks.NewMockParser(ctrl)
				manifest.parser = m
				m.EXPECT().Parse(lbWebSvcManifestPath, *manifest, gomock.Any()).Return(nil, errors.New("some error"))
			},

			wantedError: errors.New("some error"),
		},
		"returns rendered content": {
			mockDependencies: func(ctrl *gomock.Controller, manifest *LoadBalancedWebService) {
				m := mocks.NewMockParser(ctrl)
				manifest.parser = m
				m.EXPECT().Parse(lbWebSvcManifestPath, *manifest, gomock.Any()).Return(&template.Content{Buffer: bytes.NewBufferString("hello")}, nil)

			},

			wantedBinary: []byte("hello"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			manifest := &LoadBalancedWebService{}
			tc.mockDependencies(ctrl, manifest)

			// WHEN
			b, err := manifest.MarshalBinary()

			// THEN
			require.Equal(t, tc.wantedError, err)
			require.Equal(t, tc.wantedBinary, b)
		})
	}
}

func TestLoadBalancedWebService_ApplyEnv(t *testing.T) {
	testCases := map[string]struct {
		in         *LoadBalancedWebService
		envToApply string

		wanted *LoadBalancedWebService
	}{
		"with no existing environments": {
			in: &LoadBalancedWebService{
				Workload: Workload{
					Name: aws.String("phonetool"),
					Type: aws.String(LoadBalancedWebServiceType),
				},
				LoadBalancedWebServiceConfig: LoadBalancedWebServiceConfig{
					ImageConfig: ServiceImageWithPort{
						Image: Image{
							Build: BuildArgsOrString{
								BuildArgs: DockerBuildArgs{
									Dockerfile: aws.String("./Dockerfile"),
								},
							},
						},
						Port: aws.Uint16(80),
					},
					RoutingRule: RoutingRule{
						Path:            aws.String("/awards/*"),
						HealthCheckPath: aws.String("/"),
					},
					TaskConfig: TaskConfig{
						CPU:    aws.Int(1024),
						Memory: aws.Int(1024),
						Count: Count{
							Value: aws.Int(1),
						},
					},
				},
			},
			envToApply: "prod-iad",

			wanted: &LoadBalancedWebService{
				Workload: Workload{
					Name: aws.String("phonetool"),
					Type: aws.String(LoadBalancedWebServiceType),
				},
				LoadBalancedWebServiceConfig: LoadBalancedWebServiceConfig{
					ImageConfig: ServiceImageWithPort{
						Image: Image{
							Build: BuildArgsOrString{
								BuildArgs: DockerBuildArgs{
									Dockerfile: aws.String("./Dockerfile"),
								},
							},
						},
						Port: aws.Uint16(80),
					},
					RoutingRule: RoutingRule{
						Path:            aws.String("/awards/*"),
						HealthCheckPath: aws.String("/"),
					},
					TaskConfig: TaskConfig{
						CPU:    aws.Int(1024),
						Memory: aws.Int(1024),
						Count: Count{
							Value: aws.Int(1),
						},
					},
				},
			},
		},
		"with overrides": {
			in: &LoadBalancedWebService{
				Workload: Workload{
					Name: aws.String("phonetool"),
					Type: aws.String(LoadBalancedWebServiceType),
				},
				LoadBalancedWebServiceConfig: LoadBalancedWebServiceConfig{
					ImageConfig: ServiceImageWithPort{
						Image: Image{
							Build: BuildArgsOrString{
								BuildArgs: DockerBuildArgs{
									Dockerfile: aws.String("./Dockerfile"),
								},
							},
						},
						Port: aws.Uint16(80),
					},
					RoutingRule: RoutingRule{
						Path:            aws.String("/awards/*"),
						HealthCheckPath: aws.String("/"),
					},
					TaskConfig: TaskConfig{
						CPU:    aws.Int(1024),
						Memory: aws.Int(1024),
						Count: Count{
							Value: aws.Int(1),
						},
						Variables: map[string]string{
							"LOG_LEVEL":      "DEBUG",
							"DDB_TABLE_NAME": "awards",
						},
						Secrets: map[string]string{
							"GITHUB_TOKEN": "1111",
							"TWILIO_TOKEN": "1111",
						},
					},
					Sidecar: Sidecar{
						Sidecars: map[string]*SidecarConfig{
							"xray": {
								Port:       aws.String("2000"),
								Image:      aws.String("123456789012.dkr.ecr.us-east-2.amazonaws.com/xray-daemon"),
								CredsParam: aws.String("some arn"),
							},
						},
					},
					Logging: &Logging{
						ConfigFile: aws.String("mockConfigFile"),
					},
				},
				Environments: map[string]*LoadBalancedWebServiceConfig{
					"prod-iad": {
						ImageConfig: ServiceImageWithPort{
							Image: Image{
								Build: BuildArgsOrString{
									BuildArgs: DockerBuildArgs{
										Dockerfile: aws.String("./RealDockerfile"),
									},
								},
							},
							Port: aws.Uint16(5000),
						},
						RoutingRule: RoutingRule{
							TargetContainer: aws.String("xray"),
						},
						TaskConfig: TaskConfig{
							CPU: aws.Int(2046),
							Count: Count{
								Value: aws.Int(0),
							},
							Variables: map[string]string{
								"DDB_TABLE_NAME": "awards-prod",
							},
						},
						Sidecar: Sidecar{
							Sidecars: map[string]*SidecarConfig{
								"xray": {
									Port: aws.String("2000/udp"),
								},
							},
						},
						Logging: &Logging{
							SecretOptions: map[string]string{
								"FOO": "BAR",
							},
						},
					},
				},
			},
			envToApply: "prod-iad",

			wanted: &LoadBalancedWebService{
				Workload: Workload{
					Name: aws.String("phonetool"),
					Type: aws.String(LoadBalancedWebServiceType),
				},
				LoadBalancedWebServiceConfig: LoadBalancedWebServiceConfig{
					ImageConfig: ServiceImageWithPort{
						Image: Image{
							Build: BuildArgsOrString{
								BuildArgs: DockerBuildArgs{
									Dockerfile: aws.String("./RealDockerfile"),
								},
							},
						},
						Port: aws.Uint16(5000),
					},
					RoutingRule: RoutingRule{
						Path:            aws.String("/awards/*"),
						HealthCheckPath: aws.String("/"),
						TargetContainer: aws.String("xray"),
					},
					TaskConfig: TaskConfig{
						CPU:    aws.Int(2046),
						Memory: aws.Int(1024),
						Count: Count{
							Value: aws.Int(0),
						},
						Variables: map[string]string{
							"LOG_LEVEL":      "DEBUG",
							"DDB_TABLE_NAME": "awards-prod",
						},
						Secrets: map[string]string{
							"GITHUB_TOKEN": "1111",
							"TWILIO_TOKEN": "1111",
						},
					},
					Sidecar: Sidecar{
						Sidecars: map[string]*SidecarConfig{
							"xray": {
								Port:       aws.String("2000/udp"),
								Image:      aws.String("123456789012.dkr.ecr.us-east-2.amazonaws.com/xray-daemon"),
								CredsParam: aws.String("some arn"),
							},
						},
					},
					Logging: &Logging{
						ConfigFile: aws.String("mockConfigFile"),
						SecretOptions: map[string]string{
							"FOO": "BAR",
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN

			// WHEN
			conf, _ := tc.in.ApplyEnv(tc.envToApply)

			// THEN
			require.Equal(t, tc.wanted, conf, "returned configuration should have overrides from the environment")
		})
	}
}

func TestLoadBalancedWebService_BuildRequired(t *testing.T) {
	testCases := map[string]struct {
		image   Image
		want    bool
		wantErr error
	}{
		"error if both build and location are set or not set": {
			image: Image{
				Build: BuildArgsOrString{
					BuildString: aws.String("mockBuildString"),
				},
				Location: aws.String("mockLocation"),
			},
			wantErr: fmt.Errorf(`either "image.build" or "image.location" needs to be specified in the manifest`),
		},
		"return true if location is not set": {
			image: Image{
				Build: BuildArgsOrString{
					BuildString: aws.String("mockBuildString"),
				},
			},
			want: true,
		},
		"return false if location is set": {
			image: Image{
				Build:    BuildArgsOrString{},
				Location: aws.String("mockLocation"),
			},
			want: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			manifest := &LoadBalancedWebService{
				LoadBalancedWebServiceConfig: LoadBalancedWebServiceConfig{
					ImageConfig: ServiceImageWithPort{
						Image: tc.image,
					},
				},
			}

			// WHEN
			got, gotErr := manifest.BuildRequired()

			// THEN
			if gotErr != nil {
				require.EqualError(t, gotErr, tc.wantErr.Error())
			} else {
				require.Equal(t, tc.want, got)
			}
		})
	}
}
