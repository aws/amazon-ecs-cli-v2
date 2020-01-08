// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/session"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/deploy/cloudformation/stack"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/store"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/color"
	"github.com/aws/aws-sdk-go/aws"
	clientSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

const (
	// Display settings.
	minCellWidth           = 20  // minimum number of characters in a table's cell.
	tabWidth               = 4   // number of characters in between columns.
	cellPaddingWidth       = 2   // number of padding characters added by default to a cell.
	paddingChar            = ' ' // character in between columns.
	noAdditionalFormatting = 0
)

// WebAppURI represents the unique identifier to access a web application.
type WebAppURI struct {
	DNSName string // The environment's subdomain if the application is served on HTTPS. Otherwise, the public load balancer's DNS.
	Path    string // Empty if the application is served on HTTPS. Otherwise, the pattern used to match the application.
}

// CfnResource contains application resources created by cloudformation.
type CfnResource struct {
	Type       string
	PhysicalID string
}

// WebAppECSParams contains ECS deploy parameters of a web application.
type WebAppECSParams struct {
	TaskSize
	ContainerPort string
	TaskCount     string
}

type TaskSize struct {
	CPU    string
	Memory string
}

// WebAppConfig contains serialized configuration parameters for a web application.
type WebAppConfig struct {
	Environment string `json:"environment"`
	Port        string `json:"port"`
	Tasks       string `json:"tasks"`
	CPU         string `json:"cpu"`
	Memory      string `json:"memory"`
}

// WebAppRoute contains serialized route parameters for a web application.
type WebAppRoute struct {
	Environment string `json:"environment"`
	URL         string `json:"url"`
	Path        string `json:"path"`
}

// WebApp contains serialized parameters for a web application.
type WebApp struct {
	AppName        string                    `json:"appName"`
	Type           string                    `json:"type"`
	Project        string                    `json:"project"`
	Configurations []WebAppConfig            `json:"configurations"`
	Routes         []WebAppRoute             `json:"routes"`
	Resources      map[string][]*CfnResource `json:"resources,omitempty"`
}

type stackDescriber interface {
	DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error)
}

type sessionFromRoleProvider interface {
	FromRole(roleARN string, region string) (*clientSession.Session, error)
}

type envGetter interface {
	archer.EnvironmentGetter
}

func (uri *WebAppURI) String() string {
	if uri.Path != "" {
		return fmt.Sprintf("%s and path %s", color.HighlightResource("http://"+uri.DNSName), color.HighlightResource(uri.Path))
	}
	return color.HighlightResource("https://" + uri.DNSName)
}

// WebAppDescriber retrieves information about a load balanced web application.
type WebAppDescriber struct {
	app *archer.Application

	store           envGetter
	stackDescribers map[string]stackDescriber
	sessProvider    sessionFromRoleProvider
}

// NewWebAppDescriber instantiates a load balanced application.
func NewWebAppDescriber(project, app string) (*WebAppDescriber, error) {
	svc, err := store.New()
	if err != nil {
		return nil, fmt.Errorf("connect to store: %w", err)
	}
	meta, err := svc.GetApplication(project, app)
	if err != nil {
		return nil, err
	}
	return &WebAppDescriber{
		app:             meta,
		store:           svc,
		stackDescribers: make(map[string]stackDescriber),
		sessProvider:    session.NewProvider(),
	}, nil
}

// ECSParams returns the deploy infomation of a web application given an environment name.
func (d *WebAppDescriber) ECSParams(envName string) (*WebAppECSParams, error) {
	env, err := d.store.GetEnvironment(d.app.Project, envName)
	if err != nil {
		return nil, err
	}

	appParams, err := d.appParams(env)
	if err != nil {
		return nil, err
	}

	return &WebAppECSParams{
		ContainerPort: appParams[stack.LBFargateParamContainerPortKey],
		TaskSize: TaskSize{
			CPU:    appParams[stack.LBFargateTaskCPUKey],
			Memory: appParams[stack.LBFargateTaskMemoryKey],
		},
		TaskCount: appParams[stack.LBFargateTaskCountKey],
	}, nil
}

// StackResources returns the physical ID of stack resources created by cloudformation.
func (d *WebAppDescriber) StackResources(envName string) ([]*CfnResource, error) {
	env, err := d.store.GetEnvironment(d.app.Project, envName)
	if err != nil {
		return nil, err
	}

	appResource, err := d.describeStackResources(env.ManagerRoleARN, env.Region, stack.NameForApp(d.app.Project, env.Name, d.app.Name))
	if err != nil {
		return nil, err
	}
	var resources []*CfnResource
	for _, appResource := range appResource {
		resources = append(resources, &CfnResource{
			PhysicalID: aws.StringValue(appResource.PhysicalResourceId),
			Type:       aws.StringValue(appResource.ResourceType),
		})
	}

	return resources, nil
}

// URI returns the WebAppURI to identify this application uniquely given an environment name.
func (d *WebAppDescriber) URI(envName string) (*WebAppURI, error) {
	env, err := d.store.GetEnvironment(d.app.Project, envName)
	if err != nil {
		return nil, err
	}

	envOutputs, err := d.envOutputs(env)
	if err != nil {
		return nil, err
	}
	appParams, err := d.appParams(env)
	if err != nil {
		return nil, err
	}

	uri := &WebAppURI{
		DNSName: envOutputs[stack.EnvOutputPublicLoadBalancerDNSName],
		Path:    appParams[stack.LBFargateRulePathKey],
	}
	_, isHTTPS := envOutputs[stack.EnvOutputSubdomain]
	if isHTTPS {
		dnsName := fmt.Sprintf("%s.%s", d.app.Name, envOutputs[stack.EnvOutputSubdomain])
		uri = &WebAppURI{
			DNSName: dnsName,
		}
	}
	return uri, nil
}

func (d *WebAppDescriber) envOutputs(env *archer.Environment) (map[string]string, error) {
	envStack, err := d.stack(env.ManagerRoleARN, env.Region, stack.NameForEnv(d.app.Project, env.Name))
	if err != nil {
		return nil, err
	}
	outputs := make(map[string]string)
	for _, out := range envStack.Outputs {
		outputs[*out.OutputKey] = *out.OutputValue
	}
	return outputs, nil
}

func (d *WebAppDescriber) appParams(env *archer.Environment) (map[string]string, error) {
	appStack, err := d.stack(env.ManagerRoleARN, env.Region, stack.NameForApp(d.app.Project, env.Name, d.app.Name))
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	for _, param := range appStack.Parameters {
		params[*param.ParameterKey] = *param.ParameterValue
	}
	return params, nil
}

func (d *WebAppDescriber) describeStackResources(roleARN, region, stackName string) ([]*cloudformation.StackResource, error) {
	svc, err := d.stackDescriber(roleARN, region)
	if err != nil {
		return nil, err
	}
	out, err := svc.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, fmt.Errorf("describe resources for stack %s: %w", stackName, err)
	}
	return out.StackResources, nil
}

func (d *WebAppDescriber) stack(roleARN, region, stackName string) (*cloudformation.Stack, error) {
	svc, err := d.stackDescriber(roleARN, region)
	if err != nil {
		return nil, err
	}
	out, err := svc.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, fmt.Errorf("describe stack %s: %w", stackName, err)
	}
	if len(out.Stacks) == 0 {
		return nil, fmt.Errorf("stack %s not found", stackName)
	}
	return out.Stacks[0], nil
}

func (d *WebAppDescriber) stackDescriber(roleARN, region string) (stackDescriber, error) {
	if _, ok := d.stackDescribers[roleARN]; !ok {
		sess, err := d.sessProvider.FromRole(roleARN, region)
		if err != nil {
			return nil, fmt.Errorf("session for role %s and region %s: %w", roleARN, region, err)
		}
		d.stackDescribers[roleARN] = cloudformation.New(sess)
	}
	return d.stackDescribers[roleARN], nil
}

// JSONString returns the stringified WebApp struct with json format.
func (w *WebApp) JSONString() (string, error) {
	b, err := json.Marshal(w)
	if err != nil {
		return "", fmt.Errorf("marshal applications: %w", err)
	}
	return fmt.Sprintf("%s\n", b), nil
}

// HumanString returns the stringified WebApp struct with human readable format.
func (w *WebApp) HumanString() string {
	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, minCellWidth, tabWidth, cellPaddingWidth, paddingChar, noAdditionalFormatting)
	fmt.Fprintf(writer, color.Bold.Sprint("About\n\n"))
	writer.Flush()
	fmt.Fprintf(writer, "  %s\t%s\n", "Project", w.Project)
	fmt.Fprintf(writer, "  %s\t%s\n", "Name", w.AppName)
	fmt.Fprintf(writer, "  %s\t%s\n", "Type", w.Type)
	fmt.Fprintf(writer, color.Bold.Sprint("\nConfigurations\n\n"))
	writer.Flush()
	fmt.Fprintf(writer, "  %s\t%s\t%s\t%s\t%s\n", "Environment", "Tasks", "CPU (vCPU)", "Memory (MiB)", "Port")
	for _, config := range w.Configurations {
		fmt.Fprintf(writer, "  %s\t%s\t%s\t%s\t%s\n", config.Environment, config.Tasks, cpuToString(config.CPU), config.Memory, config.Port)
	}
	fmt.Fprintf(writer, color.Bold.Sprint("\nRoutes\n\n"))
	writer.Flush()
	fmt.Fprintf(writer, "  %s\t%s\t%s\n", "Environment", "URL", "Path")
	for _, route := range w.Routes {
		fmt.Fprintf(writer, "  %s\t%s\t%s\n", route.Environment, route.URL, route.Path)
	}
	if len(w.Resources) != 0 {
		fmt.Fprintf(writer, color.Bold.Sprint("\nResources\n"))
		writer.Flush()

		// Go maps don't have a guaranteed order.
		// Show the resources by the order of environments displayed under Routes for a consistent view.
		for _, route := range w.Routes {
			env := route.Environment
			resources := w.Resources[env]
			fmt.Fprintf(writer, "\n  %s\n", env)
			for _, resource := range resources {
				fmt.Fprintf(writer, "    %s\t%s\n", resource.Type, resource.PhysicalID)
			}
		}
	}
	writer.Flush()
	return b.String()
}

func cpuToString(s string) string {
	cpuInt, _ := strconv.Atoi(s)
	cpuFloat := float64(cpuInt) / 1024
	return fmt.Sprintf("%g", cpuFloat)
}
