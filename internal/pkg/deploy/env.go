// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package deploy holds the structures to deploy infrastructure resources.
// This file defines environment deployment resources.
package deploy

import (
	"github.com/aws/copilot-cli/internal/pkg/config"
	"github.com/aws/copilot-cli/internal/pkg/template"
)

const (
	// LegacyEnvTemplateVersion is the version associated with the environment template before we started versioning.
	LegacyEnvTemplateVersion = "v0.0.0"
	// LatestEnvTemplateVersion is the latest version number available for environment templates.
	LatestEnvTemplateVersion = "v1.0.0"
)

// CreateEnvironmentInput holds the fields required to deploy an environment.
type CreateEnvironmentInput struct {
	AppName                  string                  // Name of the application this environment belongs to.
	Name                     string                  // Name of the environment, must be unique within an application.
	Prod                     bool                    // Whether or not this environment is a production environment.
	ToolsAccountPrincipalARN string                  // The Principal ARN of the tools account.
	AppDNSName               string                  // The DNS name of this application, if it exists
	AdditionalTags           map[string]string       // AdditionalTags are labels applied to resources under the application.
	ImportVPCConfig          *template.ImportVPCOpts // Optional configuration if users have an existing VPC.
	AdjustVPCConfig          *template.AdjustVPCOpts // Optional configuration if users want to override default VPC configuration.

	// The version of the environment template to creat the stack. If empty, creates the legacy stack.
	Version string
}

// CreateEnvironmentResponse holds the created environment on successful deployment.
// Otherwise, the environment is set to nil and a descriptive error is returned.
type CreateEnvironmentResponse struct {
	Env *config.Environment
	Err error
}
