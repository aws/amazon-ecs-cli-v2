// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cli contains the ecs-preview subcommands.
package cli

import (
	"fmt"
	"os"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/archer"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/ecr"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/color"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/command"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/prompt"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/workspace"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GlobalOpts holds fields that are used across multiple commands.
type GlobalOpts struct {
	projectName string
	prompt      prompter
}

// NewGlobalOpts returns a GlobalOpts with the project name retrieved from viper.
func NewGlobalOpts() *GlobalOpts {
	bindProjectName()

	return &GlobalOpts{
		// Leave the projectName as empty in case it's overwritten by a global flag.
		// See https://github.com/aws/amazon-ecs-cli-v2/issues/570#issuecomment-569133741
		prompt: prompt.New(),
	}
}

// ProjectName returns the project name.
// If the name is empty, it caches it after querying viper.
func (o *GlobalOpts) ProjectName() string {
	if o.projectName != "" {
		return o.projectName
	}
	o.projectName = viper.GetString(projectFlag)
	return o.projectName
}

// actionCommand is the interface that every command that creates a resource implements.
type actionCommand interface {
	Ask() error
	Validate() error
	Execute() error
	RecommendedActions() []string
}

// bindProjectName loads the project's name to viper.
// If there is an error, we swallow the error and leave the default value as empty string.
func bindProjectName() {
	name, err := loadProjectName()
	if err != nil {
		return
	}
	viper.SetDefault(projectFlag, name)
}

// loadProjectName retrieves the project's name from the workspace if it exists and returns it.
// If there is an error, it returns an empty string and the error.
func loadProjectName() (string, error) {
	// Load the workspace and set the project flag.
	ws, err := workspace.New()
	if err != nil {
		// If there's an error fetching the workspace, fall back to requiring
		// the project flag be set.
		return "", fmt.Errorf("fetching workspace: %w", err)
	}

	summary, err := ws.Summary()
	if err != nil {
		// If there's an error reading from the workspace, fall back to requiring
		// the project flag be set.
		return "", fmt.Errorf("reading from workspace: %w", err)
	}
	return summary.ProjectName, nil
}

type errReservedArg struct {
	val string
}

func (e *errReservedArg) Error() string {
	return fmt.Sprintf(`argument %s is a reserved keyword, please use a different value`, color.HighlightUserInput(e.val))
}

// reservedArgs returns an error if the arguments contain any reserved keywords.
func reservedArgs(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return nil
	}
	if args[0] == "local" {
		return &errReservedArg{val: "local"}
	}
	return nil
}

// runCmdE wraps one of the run error methods, PreRunE, RunE, of a cobra command so that if a user
// types "help" in the arguments the usage string is printed instead of running the command.
func runCmdE(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "help" {
			_ = cmd.Help() // Help always returns nil.
			os.Exit(0)
		}
		return f(cmd, args)
	}
}

type projectService interface {
	archer.ProjectStore
	archer.EnvironmentStore
	archer.ApplicationStore
}

type ecrService interface {
	GetRepository(name string) (string, error)
	GetECRAuth() (ecr.Auth, error)
}

type dockerService interface {
	Build(uri, tag, path string) error
	Login(uri, username, password string) error
	Push(uri, tag string) error
}

type runner interface {
	Run(name string, args []string, options ...command.Option) error
}

type defaultSessionProvider interface {
	Default() (*session.Session, error)
}

type regionalSessionProvider interface {
	DefaultWithRegion(region string) (*session.Session, error)
}

type sessionFromRoleProvider interface {
	FromRole(roleARN string, region string) (*session.Session, error)
}

type sessionProvider interface {
	defaultSessionProvider
	regionalSessionProvider
	sessionFromRoleProvider
}
