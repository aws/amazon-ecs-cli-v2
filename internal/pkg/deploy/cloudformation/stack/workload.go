// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stack

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/copilot-cli/internal/pkg/addon"
	"github.com/aws/copilot-cli/internal/pkg/deploy"
	"github.com/aws/copilot-cli/internal/pkg/manifest"
	"github.com/aws/copilot-cli/internal/pkg/template"
)

// Template rendering configuration common across workloads.
const (
	wkldParamsTemplatePath = "workloads/params.json.tmpl"
)

// Parameter logical IDs common across workloads.
const (
	WorkloadAppNameParamKey           = "AppName"
	WorkloadEnvNameParamKey           = "EnvName"
	WorkloadNameParamKey              = "WorkloadName"
	WorkloadContainerImageParamKey    = "ContainerImage"
	WorkloadTaskCPUParamKey           = "TaskCPU"
	WorkloadTaskMemoryParamKey        = "TaskMemory"
	WorkloadTaskCountParamKey         = "TaskCount"
	WorkloadLogRetentionParamKey      = "LogRetention"
	WorkloadAddonsTemplateURLParamKey = "AddonsTemplateURL"
)

// RuntimeConfig represents configuration that's defined outside of the manifest file
// that is needed to create a CloudFormation stack.
type RuntimeConfig struct {
	Image             *ECRImage         // Optional. Image location in an ECR repository.
	AddonsTemplateURL string            // Optional. S3 object URL for the addons template.
	AdditionalTags    map[string]string // AdditionalTags are labels applied to resources in the workload stack.
}

// ECRImage represents configuration about the pushed ECR image that is needed to
// create a CloudFormation stack.
type ECRImage struct {
	RepoURL  string // RepoURL is the ECR repository URL the container image should be pushed to.
	ImageTag string // Tag is the container image's unique tag.
}

// GetLocation returns the location of the ECR image.
func (i ECRImage) GetLocation() string {
	return fmt.Sprintf("%s:%s", i.RepoURL, i.ImageTag)
}

type templater interface {
	Template() (string, error)
}

type location interface {
	GetLocation() string
}

// wkld represents a containerized workload running on Amazon ECS.
// A workload can be a long-running service, an ephemeral task, or a periodic task.
type wkld struct {
	name  string
	env   string
	app   string
	tc    manifest.TaskConfig
	rc    RuntimeConfig
	image location

	parser template.Parser
	addons templater
}

// StackName returns the name of the stack.
func (w *wkld) StackName() string {
	return NameForService(w.app, w.env, w.name)
}

// Parameters returns the list of CloudFormation parameters used by the template.
func (w *wkld) Parameters() ([]*cloudformation.Parameter, error) {
	desiredCount := w.tc.Count.Value
	// If auto scaling is configured, override the desired count value.
	if !w.tc.Count.Autoscaling.IsEmpty() {
		min, _, err := w.tc.Count.Autoscaling.Range.Parse()
		if err != nil {
			return nil, fmt.Errorf("parse task count value %s: %w", string(w.tc.Count.Autoscaling.Range), err)
		}
		desiredCount = aws.Int(min)
	}
	var img string
	if w.image != nil {
		img = w.image.GetLocation()
	}
	if w.rc.Image != nil {
		img = w.rc.Image.GetLocation()
	}
	return []*cloudformation.Parameter{
		{
			ParameterKey:   aws.String(WorkloadAppNameParamKey),
			ParameterValue: aws.String(w.app),
		},
		{
			ParameterKey:   aws.String(WorkloadEnvNameParamKey),
			ParameterValue: aws.String(w.env),
		},
		{
			ParameterKey:   aws.String(WorkloadNameParamKey),
			ParameterValue: aws.String(w.name),
		},
		{
			ParameterKey:   aws.String(WorkloadContainerImageParamKey),
			ParameterValue: aws.String(img),
		},
		{
			ParameterKey:   aws.String(WorkloadTaskCPUParamKey),
			ParameterValue: aws.String(strconv.Itoa(aws.IntValue(w.tc.CPU))),
		},
		{
			ParameterKey:   aws.String(WorkloadTaskMemoryParamKey),
			ParameterValue: aws.String(strconv.Itoa(aws.IntValue(w.tc.Memory))),
		},
		{
			ParameterKey:   aws.String(WorkloadTaskCountParamKey),
			ParameterValue: aws.String(strconv.Itoa(*desiredCount)),
		},
		{
			ParameterKey:   aws.String(WorkloadLogRetentionParamKey),
			ParameterValue: aws.String("30"),
		},
		{
			ParameterKey:   aws.String(WorkloadAddonsTemplateURLParamKey),
			ParameterValue: aws.String(w.rc.AddonsTemplateURL),
		},
	}, nil
}

// Tags returns the list of tags to apply to the CloudFormation stack.
func (w *wkld) Tags() []*cloudformation.Tag {
	return mergeAndFlattenTags(w.rc.AdditionalTags, map[string]string{
		deploy.AppTagKey:     w.app,
		deploy.EnvTagKey:     w.env,
		deploy.ServiceTagKey: w.name,
	})
}

type templateConfigurer interface {
	Parameters() ([]*cloudformation.Parameter, error)
	Tags() []*cloudformation.Tag
}

func (w *wkld) templateConfiguration(tc templateConfigurer) (string, error) {
	params, err := tc.Parameters()
	if err != nil {
		return "", err
	}
	doc, err := w.parser.Parse(wkldParamsTemplatePath, struct {
		Parameters []*cloudformation.Parameter
		Tags       []*cloudformation.Tag
	}{
		Parameters: params,
		Tags:       tc.Tags(),
	}, template.WithFuncs(map[string]interface{}{
		"inc": template.IncFunc,
	}))
	if err != nil {
		return "", err
	}
	return doc.String(), nil
}

func (w *wkld) addonsOutputs() (*template.WorkloadNestedStackOpts, error) {
	stack, err := w.addons.Template()
	if err != nil {
		var noAddonsErr *addon.ErrDirNotExist
		if !errors.As(err, &noAddonsErr) {
			return nil, fmt.Errorf("generate addons template for %s: %w", w.name, err)
		}
		return nil, nil // Addons directory does not exist, so there are no outputs and error.
	}

	out, err := addon.Outputs(stack)
	if err != nil {
		return nil, fmt.Errorf("get addons outputs for %s: %w", w.name, err)
	}
	return &template.WorkloadNestedStackOpts{
		StackName:       addon.StackName,
		VariableOutputs: envVarOutputNames(out),
		SecretOutputs:   secretOutputNames(out),
		PolicyOutputs:   managedPolicyOutputNames(out),
	}, nil
}

func secretOutputNames(outputs []addon.Output) []string {
	var secrets []string
	for _, out := range outputs {
		if out.IsSecret {
			secrets = append(secrets, out.Name)
		}
	}
	return secrets
}

func managedPolicyOutputNames(outputs []addon.Output) []string {
	var policies []string
	for _, out := range outputs {
		if out.IsManagedPolicy {
			policies = append(policies, out.Name)
		}
	}
	return policies
}

func envVarOutputNames(outputs []addon.Output) []string {
	var envVars []string
	for _, out := range outputs {
		if !out.IsSecret && !out.IsManagedPolicy {
			envVars = append(envVars, out.Name)
		}
	}
	return envVars
}
