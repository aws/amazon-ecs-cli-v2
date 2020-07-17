// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package docker provides an interface to the system's Docker daemon.
package docker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/copilot-cli/internal/pkg/term/command"
)

// Runner represents a command that can be run.
type Runner struct {
	runner
}

type runner interface {
	Run(name string, args []string, options ...command.Option) error
}

// New returns a Runner.
func New() Runner {
	return Runner{
		runner: command.New(),
	}
}

type BuildArguments struct {
	Dockerfile string
	Context    string
	Args       map[string]string
}

// Build will run a `docker build` command with the input uri, tag, and Dockerfile path.
func (r Runner) Build(uri, imageTag, path, context string, additionalTags ...string) error {
	imageName := imageName(uri, imageTag)
	var dfDir string
	if context == "" {
		dfDir = filepath.Dir(path)
	} else {
		dfDir = context
	}

	args := []string{"build"}
	for _, tag := range append(additionalTags, imageTag) {
		args = append(args, "-t", imageName(uri, tag))
	}
	args = append(args, dfDir, "-f", path)

	err := r.Run("docker", args)
	if err != nil {
		return fmt.Errorf("building image: %w", err)
	}

	return nil
}

// Login will run a `docker login` command against the Service repository URI with the input uri and auth data.
func (r Runner) Login(uri, username, password string) error {
	err := r.Run("docker",
		[]string{"login", "-u", username, "--password-stdin", uri},
		command.Stdin(strings.NewReader(password)))

	if err != nil {
		return fmt.Errorf("authenticate to ECR: %w", err)
	}

	return nil
}

// Push will run `docker push` command against the repository URI with the input uri and image tags.
func (r Runner) Push(uri, imageTag string, additionalTags ...string) error {
	for _, imageTag := range append(additionalTags, imageTag) {
		path := imageName(uri, imageTag)

		err := r.Run("docker", []string{"push", path})
		if err != nil {
			return fmt.Errorf("docker push %s: %w", path, err)
		}
	}

	return nil
}

func imageName(uri, tag string) string {
	return fmt.Sprintf("%s:%s", uri, tag)
}
