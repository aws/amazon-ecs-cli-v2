// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package repository provides support for building and pushing images to a repository.
package repository

import (
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/docker"
)

// ContainerLoginBuildPusher provides support for logging in to repositories, building images and pushing images to repositories.
type ContainerLoginBuildPusher interface {
	Build(args *docker.BuildArguments) error
	Login(uri, username, password string) error
	Push(uri, imageTag string, additionalTags ...string) error
}

// Registry gets information of repositories.
type Registry interface {
	RepositoryURI(name string) (string, error)
	Auth() (string, string, error)
}

// Repository builds and pushes images to a repository.
type Repository struct {
	name     string
	registry Registry

	uri string
}

// New instantiates a new Repository.
func New(name string, registry Registry) (*Repository, error) {
	uri, err := registry.RepositoryURI(name)
	if err != nil {
		return nil, fmt.Errorf("get repository URI: %w", err)
	}

	return &Repository{
		name:     name,
		uri:      uri,
		registry: registry,
	}, nil
}

// BuildAndPush builds the image from Dockerfile and pushes it to the repository with tags.
func (r *Repository) BuildAndPush(docker ContainerLoginBuildPusher, args *docker.BuildArguments) error {
	if args.URI == "" {
		args.URI = r.uri
	}
	if err := docker.Build(args); err != nil {
		return fmt.Errorf("build Dockerfile at %s: %w", args.Dockerfile, err)
	}

	username, password, err := r.registry.Auth()
	if err != nil {
		return fmt.Errorf("get auth: %w", err)
	}

	if err := docker.Login(args.URI, username, password); err != nil {
		return fmt.Errorf("login to repo %s: %w", r.name, err)
	}

	if err := docker.Push(args.URI, args.ImageTag, args.AdditionalTags...); err != nil {
		return fmt.Errorf("push to repo %s: %w", r.name, err)
	}
	return nil
}

// URI returns the uri of the repository.
func (r *Repository) URI() string {
	return r.uri
}
