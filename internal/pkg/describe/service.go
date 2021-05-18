// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"fmt"
	"io"
	"net/url"
	"sort"

	awsecs "github.com/aws/copilot-cli/internal/pkg/aws/ecs"

	"github.com/aws/copilot-cli/internal/pkg/ecs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/copilot-cli/internal/pkg/aws/apprunner"
	"github.com/aws/copilot-cli/internal/pkg/aws/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/deploy/cloudformation/stack"
)

const (
	// Ignored resources
	rulePriorityFunction = "Custom::RulePriorityFunction"
	waitCondition        = "AWS::CloudFormation::WaitCondition"
	waitConditionHandle  = "AWS::CloudFormation::WaitConditionHandle"
)

const apprunnerServiceType = "AWS::AppRunner::Service"

// envVar contains serialized environment variables for a service.
type envVar struct {
	Environment string `json:"environment"`
	Name        string `json:"name"`
	Value       string `json:"value"`
}

type envVars []*envVar

func (e envVars) humanString(w io.Writer) {
	headers := []string{"Name", "Environment", "Value"}
	var rows [][]string
	sort.SliceStable(e, func(i, j int) bool { return e[i].Environment < e[j].Environment })
	sort.SliceStable(e, func(i, j int) bool { return e[i].Name < e[j].Name })

	for _, v := range e {
		rows = append(rows, []string{v.Name, v.Environment, v.Value})
	}

	printTable(w, headers, rows)
}

type containerEnvVar struct {
	*envVar

	Container string `json:"container"`
}

type containerEnvVars []*containerEnvVar

func (e containerEnvVars) humanString(w io.Writer) {
	headers := []string{"Name", "Container", "Environment", "Value"}
	var rows [][]string
	sort.SliceStable(e, func(i, j int) bool { return e[i].Environment < e[j].Environment })
	sort.SliceStable(e, func(i, j int) bool { return e[i].Container < e[j].Container })
	sort.SliceStable(e, func(i, j int) bool { return e[i].Name < e[j].Name })

	for _, v := range e {
		rows = append(rows, []string{v.Name, v.Container, v.Environment, v.Value})
	}

	printTable(w, headers, rows)
}

type stackAndResourcesDescriber interface {
	Stack(stackName string) (*cloudformation.Stack, error)
	StackResources(stackName string) ([]*cloudformation.StackResource, error)
	Metadata(stackName string) (string, error)
}

type ecsClient interface {
	TaskDefinition(app, env, svc string) (*awsecs.TaskDefinition, error)
}

type ecsSvcDescriber interface {
	Params() (map[string]string, error)
	EnvOutputs() (map[string]string, error)
	EnvVars() ([]*awsecs.ContainerEnvVar, error)
	Secrets() ([]*awsecs.ContainerSecret, error)
	ServiceStackResources() ([]*cloudformation.StackResource, error)
}

type apprunnerClient interface {
	DescribeService(svcArn string) (*apprunner.Service, error)
}

type apprunnerSvcDescriber interface {
	Params() (map[string]string, error)
	EnvOutputs() (map[string]string, error)
	ServiceStackResources() ([]*cloudformation.StackResource, error)
	Service() (*apprunner.Service, error)
	ServiceARN() (string, error)
	ServiceURL() (string, error)
}

// ConfigStoreSvc wraps methods of config store.
type ConfigStoreSvc interface {
	GetEnvironment(appName string, environmentName string) (*config.Environment, error)
	ListEnvironments(appName string) ([]*config.Environment, error)
	ListServices(appName string) ([]*config.Workload, error)
	GetWorkload(appName string, name string) (*config.Workload, error)
}

// DeployedEnvServicesLister wraps methods of deploy store.
type DeployedEnvServicesLister interface {
	ListEnvironmentsDeployedTo(appName string, svcName string) ([]string, error)
	ListDeployedServices(appName string, envName string) ([]string, error)
}

// ServiceConfig contains serialized configuration parameters for a service.
type ServiceConfig struct {
	Environment string `json:"environment"`
	Port        string `json:"port"`
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
}

type configurations []*ServiceConfig

func (c configurations) humanString(w io.Writer) {
	headers := []string{"Environment", "CPU (vCPU)", "Memory (MiB)", "Port"}
	var rows [][]string
	for _, config := range c {
		rows = append(rows, []string{config.Environment, cpuToString(config.CPU), config.Memory, config.Port})
	}

	printTable(w, headers, rows)
}

type ECSServiceConfig struct {
	*ServiceConfig

	Tasks string `json:"tasks"`
}

type ecsConfigurations []*ECSServiceConfig

func (c ecsConfigurations) humanString(w io.Writer) {
	headers := []string{"Environment", "Tasks", "CPU (vCPU)", "Memory (MiB)", "Port"}
	var rows [][]string
	for _, config := range c {
		rows = append(rows, []string{config.Environment, config.Tasks, cpuToString(config.CPU), config.Memory, config.Port})
	}

	printTable(w, headers, rows)
}

// ServiceDescriber provides base functionality for retrieving info about a service.
type ServiceDescriber struct {
	app     string
	service string
	env     string

	cfn       cfn
	ecsClient ecsClient
	sess      *session.Session
}

// NewServiceConfig contains fields that initiates ServiceDescriber struct.
type NewServiceConfig struct {
	App         string
	Env         string
	Svc         string
	ConfigStore ConfigStoreSvc
}

func NewServiceDescriber(opt NewServiceConfig) (*ServiceDescriber, error) {
	environment, err := opt.ConfigStore.GetEnvironment(opt.App, opt.Env)
	if err != nil {
		return nil, fmt.Errorf("get environment %s: %w", opt.Env, err)
	}
	sess, err := sessions.NewProvider().FromRole(environment.ManagerRoleARN, environment.Region)
	if err != nil {
		return nil, err
	}
	return &ServiceDescriber{
		app:     opt.App,
		service: opt.Svc,
		env:     opt.Env,

		cfn:       cloudformation.New(sess),
		ecsClient: ecs.New(sess),
		sess:      sess,
	}, nil
}

// EnvVars returns the environment variables of the task definition.
func (d *ServiceDescriber) EnvVars() ([]*awsecs.ContainerEnvVar, error) {
	taskDefinition, err := d.ecsClient.TaskDefinition(d.app, d.env, d.service)
	if err != nil {
		return nil, fmt.Errorf("describe task definition for service %s: %w", d.service, err)
	}
	return taskDefinition.EnvironmentVariables(), nil
}

// Secrets returns the secrets of the task definition.
func (d *ServiceDescriber) Secrets() ([]*awsecs.ContainerSecret, error) {
	taskDefinition, err := d.ecsClient.TaskDefinition(d.app, d.env, d.service)
	if err != nil {
		return nil, fmt.Errorf("describe task definition for service %s: %w", d.service, err)
	}
	return taskDefinition.Secrets(), nil
}

// ServiceStackResources returns the filtered service stack resources created by CloudFormation.
func (d *ServiceDescriber) ServiceStackResources() ([]*cloudformation.StackResource, error) {
	svcResources, err := d.cfn.StackResources(stack.NameForService(d.app, d.env, d.service))
	if err != nil {
		return nil, err
	}
	var resources []*cloudformation.StackResource
	ignoredResources := map[string]bool{
		rulePriorityFunction: true,
		waitCondition:        true,
		waitConditionHandle:  true,
	}
	for _, svcResource := range svcResources {
		if ignoredResources[aws.StringValue(svcResource.ResourceType)] {
			continue
		}
		resources = append(resources, svcResource)
	}

	return resources, nil
}

// EnvOutputs returns the output of the environment stack.
func (d *ServiceDescriber) EnvOutputs() (map[string]string, error) {
	envStack, err := d.cfn.Describe(stack.NameForEnv(d.app, d.env))
	if err != nil {
		return nil, err
	}
	outputs := make(map[string]string)
	for _, out := range envStack.Outputs {
		outputs[*out.OutputKey] = *out.OutputValue
	}
	return outputs, nil
}

// Params returns the parameters of the service stack.
func (d *ServiceDescriber) Params() (map[string]string, error) {
	svcStack, err := d.cfn.Describe(stack.NameForService(d.app, d.env, d.service))
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	for _, param := range svcStack.Parameters {
		params[*param.ParameterKey] = *param.ParameterValue
	}
	return params, nil
}

// ECSServiceDescriber retrieves information about a service.
type ECSServiceDescriber struct {
	*ServiceDescriber

	ecsClient ecsClient
}

// NewServiceDescriber instantiates a new service.
func NewECSServiceDescriber(opt NewServiceConfig) (*ECSServiceDescriber, error) {
	serviceDescriber, err := NewServiceDescriber(opt)
	if err != nil {
		return nil, err
	}

	return &ECSServiceDescriber{
		ServiceDescriber: serviceDescriber,

		ecsClient: ecs.New(serviceDescriber.sess),
	}, nil
}

// EnvVars returns the environment variables of the task definition.
func (d *ECSServiceDescriber) EnvVars() ([]*awsecs.ContainerEnvVar, error) {
	taskDefinition, err := d.ecsClient.TaskDefinition(d.app, d.env, d.service)
	if err != nil {
		return nil, fmt.Errorf("describe task definition for service %s: %w", d.service, err)
	}
	return taskDefinition.EnvironmentVariables(), nil
}

// Secrets returns the secrets of the task definition.
func (d *ECSServiceDescriber) Secrets() ([]*awsecs.ContainerSecret, error) {
	taskDefinition, err := d.ecsClient.TaskDefinition(d.app, d.env, d.service)
	if err != nil {
		return nil, fmt.Errorf("describe task definition for service %s: %w", d.service, err)
	}
	return taskDefinition.Secrets(), nil
}

// AppRunnerServiceDescriber retrieves information about a service.
type AppRunnerServiceDescriber struct {
	*ServiceDescriber

	apprunnerClient apprunnerClient
}

func NewAppRunnerServiceDescriber(opt NewServiceConfig) (*AppRunnerServiceDescriber, error) {
	serviceDescriber, err := NewServiceDescriber(opt)
	if err != nil {
		return nil, err
	}

	return &AppRunnerServiceDescriber{
		ServiceDescriber: serviceDescriber,

		apprunnerClient: apprunner.New(serviceDescriber.sess),
	}, nil
}

func (d *AppRunnerServiceDescriber) ServiceARN() (string, error) {
	serviceStackResources, err := d.ServiceStackResources()
	if err != nil {
		return "", err
	}

	for _, resource := range serviceStackResources {
		arn := aws.StringValue(resource.PhysicalResourceId)
		if aws.StringValue(resource.ResourceType) == apprunnerServiceType && arn != "" {
			return arn, nil
		}
	}

	return "", fmt.Errorf("no App Runner Service in service stack")
}

func (d *AppRunnerServiceDescriber) Service() (*apprunner.Service, error) {
	serviceARN, err := d.ServiceARN()
	if err != nil {
		return nil, err
	}

	return d.apprunnerClient.DescribeService(serviceARN)
}

func (d *AppRunnerServiceDescriber) ServiceURL() (string, error) {
	service, err := d.Service()
	if err != nil {
		return "", fmt.Errorf("retrieve service URI: %w", err)
	}

	return formatAppRunnerUrl(service.ServiceURL), nil
}

func formatAppRunnerUrl(serviceURL string) string {
	svcUrl := &url.URL{
		Host: serviceURL,
		// App Runner defaults to https
		Scheme: "https",
	}

	return svcUrl.String()
}
