// +build integration

// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/template"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestTemplate_ParseScheduledJob(t *testing.T) {
	testCases := map[string]struct {
		opts template.WorkloadOpts
	}{
		"renders a valid template by default": {
			opts: template.WorkloadOpts{},
		},
		"renders with timeout and no retries": {
			opts: template.WorkloadOpts{
				StateMachine: &template.StateMachineOpts{
					Timeout: aws.Int(3600),
				},
			},
		},
		"renders with options": {
			opts: template.WorkloadOpts{
				StateMachine: &template.StateMachineOpts{
					Retries: aws.Int(5),
					Timeout: aws.Int(3600),
				},
			},
		},
		"renders with options and addons": {
			opts: template.WorkloadOpts{
				StateMachine: &template.StateMachineOpts{
					Retries: aws.Int(3),
				},
				NestedStack: &template.WorkloadNestedStackOpts{
					StackName:       "AddonsStack",
					VariableOutputs: []string{"TableName"},
					SecretOutputs:   []string{"TablePassword"},
					PolicyOutputs:   []string{"TablePolicy"},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			sess, err := sessions.NewProvider().Default()
			require.NoError(t, err)
			cfn := cloudformation.New(sess)
			tpl := template.New()

			// WHEN
			content, err := tpl.ParseScheduledJob(tc.opts)
			require.NoError(t, err)

			// THEN
			_, err = cfn.ValidateTemplate(&cloudformation.ValidateTemplateInput{
				TemplateBody: aws.String(content.String()),
			})
			require.NoError(t, err, content.String())
		})
	}
}

func TestTemplate_ParseLoadBalancedWebService(t *testing.T) {
	defaultHttpHealthCheck := template.HTTPHealthCheckOpts{
		HealthCheckPath: "/",
	}
	testCases := map[string]struct {
		opts template.WorkloadOpts
	}{
		"renders a valid template by default": {
			opts: template.WorkloadOpts{
				HTTPHealthCheck: defaultHttpHealthCheck,
			},
		},
		"renders a valid template with addons with no outputs": {
			opts: template.WorkloadOpts{
				HTTPHealthCheck: defaultHttpHealthCheck,
				NestedStack: &template.WorkloadNestedStackOpts{
					StackName: "AddonsStack",
				},
			},
		},
		"renders a valid template with addons with outputs": {
			opts: template.WorkloadOpts{
				HTTPHealthCheck: defaultHttpHealthCheck,
				NestedStack: &template.WorkloadNestedStackOpts{
					StackName:       "AddonsStack",
					VariableOutputs: []string{"TableName"},
					SecretOutputs:   []string{"TablePassword"},
					PolicyOutputs:   []string{"TablePolicy"},
				},
			},
		},
		"renders a valid template with all storage options": {
			opts: template.WorkloadOpts{
				HTTPHealthCheck: defaultHttpHealthCheck,
				Storage: &template.StorageOpts{
					EFSPerms: []*template.EFSPermission{
						{
							AccessPointID: aws.String("ap-1234"),
							FilesystemID:  aws.String("fs-5678"),
							Write:         true,
						},
					},
					MountPoints: []*template.MountPoint{
						{
							SourceVolume:  aws.String("efs"),
							ContainerPath: aws.String("/var/www"),
							ReadOnly:      aws.Bool(false),
						},
					},
					Volumes: []*template.Volume{
						{
							AccessPointID: aws.String("ap-1234"),
							Filesystem:    aws.String("fs-5678"),
							IAM:           aws.String("ENABLED"),
							Name:          aws.String("efs"),
							RootDirectory: aws.String("/"),
						},
					},
				},
			},
		},
		"renders a valid template with minimal storage options": {
			opts: template.WorkloadOpts{
				HTTPHealthCheck: defaultHttpHealthCheck,
				Storage: &template.StorageOpts{
					EFSPerms: []*template.EFSPermission{
						{
							FilesystemID: aws.String("fs-5678"),
						},
					},
					MountPoints: []*template.MountPoint{
						{
							SourceVolume:  aws.String("efs"),
							ContainerPath: aws.String("/var/www"),
							ReadOnly:      aws.Bool(true),
						},
					},
					Volumes: []*template.Volume{
						{
							Filesystem:    aws.String("fs-5678"),
							Name:          aws.String("efs"),
							RootDirectory: aws.String("/"),
						},
					},
				},
			},
		},
		"renders a valid template with entrypoint and command overrides": {
			opts: template.WorkloadOpts{
				HTTPHealthCheck: defaultHttpHealthCheck,
				EntryPoint: []string{"/bin/echo", "hello"},
				Command: []string{"world"},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			sess, err := sessions.NewProvider().Default()
			require.NoError(t, err)
			cfn := cloudformation.New(sess)
			tpl := template.New()

			// WHEN
			content, err := tpl.ParseLoadBalancedWebService(tc.opts)
			require.NoError(t, err)

			// THEN
			_, err = cfn.ValidateTemplate(&cloudformation.ValidateTemplateInput{
				TemplateBody: aws.String(content.String()),
			})
			require.NoError(t, err, content.String())
		})
	}
}

func TestTemplate_ParseNetwork(t *testing.T) {
	type cfn struct {
		Resources struct {
			Service struct {
				Properties struct {
					NetworkConfiguration map[interface{}]interface{} `yaml:"NetworkConfiguration"`
				} `yaml:"Properties"`
			} `yaml:"Service"`
		} `yaml:"Resources"`
	}

	testCases := map[string]struct {
		input *template.NetworkOpts

		wantedNetworkConfig string
	}{
		"should render AWS VPC configuration for public subnets by default": {
			input: nil,
			wantedNetworkConfig: `
  AwsvpcConfiguration:
    AssignPublicIp: ENABLED
    Subnets:
    - Fn::Select:
      - 0
      - Fn::Split:
        - ','
        - Fn::ImportValue: !Sub '${AppName}-${EnvName}-PublicSubnets'
    - Fn::Select:
      - 1
      - Fn::Split:
        - ','
        - Fn::ImportValue: !Sub '${AppName}-${EnvName}-PublicSubnets'
    SecurityGroups:
      - Fn::ImportValue: !Sub '${AppName}-${EnvName}-EnvironmentSecurityGroup'
`,
		},
		"should render AWS VPC configuration for private subnets": {
			input: &template.NetworkOpts{
				AssignPublicIP: "DISABLED",
				SubnetsType:    "PrivateSubnets",
			},
			wantedNetworkConfig: `
  AwsvpcConfiguration:
    AssignPublicIp: DISABLED
    Subnets:
    - Fn::Select:
      - 0
      - Fn::Split:
        - ','
        - Fn::ImportValue: !Sub '${AppName}-${EnvName}-PrivateSubnets'
    - Fn::Select:
      - 1
      - Fn::Split:
        - ','
        - Fn::ImportValue: !Sub '${AppName}-${EnvName}-PrivateSubnets'
    SecurityGroups:
      - Fn::ImportValue: !Sub '${AppName}-${EnvName}-EnvironmentSecurityGroup'
`,
		},
		"should render AWS VPC configuration for private subnets with security groups": {
			input: &template.NetworkOpts{
				AssignPublicIP: "DISABLED",
				SubnetsType:    "PrivateSubnets",
				SecurityGroups: []string{
					"sg-1bcf1d5b",
					"sg-asdasdas",
				},
			},
			wantedNetworkConfig: `
  AwsvpcConfiguration:
    AssignPublicIp: DISABLED
    Subnets:
    - Fn::Select:
      - 0
      - Fn::Split:
        - ','
        - Fn::ImportValue: !Sub '${AppName}-${EnvName}-PrivateSubnets'
    - Fn::Select:
      - 1
      - Fn::Split:
        - ','
        - Fn::ImportValue: !Sub '${AppName}-${EnvName}-PrivateSubnets'
    SecurityGroups:
      - Fn::ImportValue: !Sub '${AppName}-${EnvName}-EnvironmentSecurityGroup'
      - "sg-1bcf1d5b"
      - "sg-asdasdas"
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			tpl := template.New()
			wanted := make(map[interface{}]interface{})
			err := yaml.Unmarshal([]byte(tc.wantedNetworkConfig), &wanted)
			require.NoError(t, err, "unmarshal wanted config")

			// WHEN
			content, err := tpl.ParseLoadBalancedWebService(template.WorkloadOpts{
				Network: tc.input,
			})

			// THEN
			require.NoError(t, err, "parse load balanced web service")
			var actual cfn
			err = yaml.Unmarshal(content.Bytes(), &actual)
			require.NoError(t, err, "unmarshal actual config")
			require.Equal(t, wanted, actual.Resources.Service.Properties.NetworkConfiguration)
		})
	}
}
