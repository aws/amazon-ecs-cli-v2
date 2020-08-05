// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package deploy holds the structures to deploy infrastructure resources.
package deploy

import (
	"fmt"

	rg "github.com/aws/copilot-cli/internal/pkg/aws/resourcegroups"
	"github.com/aws/copilot-cli/internal/pkg/aws/sessions"
	"github.com/aws/copilot-cli/internal/pkg/config"
)

const (
	// AppTagKey is tag key for Copilot app.
	AppTagKey = "copilot-application"
	// EnvTagKey is tag key for Copilot env.
	EnvTagKey = "copilot-environment"
	// ServiceTagKey is tag key for Copilot svc.
	ServiceTagKey = "copilot-service"
	// TaskTagKey is tag key for Copilot task.
	TaskTagKey = "copilot-task"
)

const (
	ecsServiceResourceType = "ecs:service"
)

// Resource represents an AWS resource.
type Resource struct {
	LogicalName string
	Type        string
}

// ResourceEvent represents a status update for an AWS resource during a deployment.
type ResourceEvent struct {
	Resource
	Status       string
	StatusReason string
}

type resourceGetter interface {
	GetResourcesByTags(resourceType string, tags map[string]string) ([]*rg.Resource, error)
}

// ConfigStoreClient wraps config store methods utilized by deploy store.
type ConfigStoreClient interface {
	GetEnvironment(appName string, environmentName string) (*config.Environment, error)
	ListEnvironments(appName string) ([]*config.Environment, error)
	GetService(appName, svcName string) (*config.Service, error)
}

// Store fetches information on deployed services.
type Store struct {
	configStore         ConfigStoreClient
	newRgClientFromIDs  func(string, string) (resourceGetter, error)
	newRgClientFromRole func(string, string) (resourceGetter, error)
}

// NewStore returns a new store.
func NewStore(store ConfigStoreClient) (*Store, error) {
	s := &Store{
		configStore: store,
	}
	s.newRgClientFromIDs = func(appName, envName string) (resourceGetter, error) {
		env, err := s.configStore.GetEnvironment(appName, envName)
		if err != nil {
			return nil, fmt.Errorf("get environment config %s: %w", envName, err)
		}
		sess, err := sessions.NewProvider().FromRole(env.ManagerRoleARN, env.Region)
		if err != nil {
			return nil, fmt.Errorf("create new session from env role: %w", err)
		}
		return rg.New(sess), nil
	}
	s.newRgClientFromRole = func(roleARN, region string) (resourceGetter, error) {
		sess, err := sessions.NewProvider().FromRole(roleARN, region)
		if err != nil {
			return nil, fmt.Errorf("create new session from env role: %w", err)
		}
		return rg.New(sess), nil
	}
	return s, nil
}

// ListDeployedServices returns the names of deployed services in an environment part of an application.
func (s *Store) ListDeployedServices(appName string, envName string) ([]string, error) {
	rgClient, err := s.newRgClientFromIDs(appName, envName)
	if err != nil {
		return nil, err
	}
	resources, err := rgClient.GetResourcesByTags(ecsServiceResourceType, map[string]string{
		AppTagKey: appName,
		EnvTagKey: envName,
	})
	if err != nil {
		return nil, fmt.Errorf("get resources by Copilot tags: %w", err)
	}
	svcs := make([]string, len(resources))
	for ind, resource := range resources {
		svcName := resource.Tags[ServiceTagKey]
		if svcName == "" {
			return nil, fmt.Errorf("service with ARN %s is not tagged with %s", resource.ARN, ServiceTagKey)
		}
		svc, err := s.configStore.GetService(appName, svcName)
		if err != nil {
			return nil, fmt.Errorf("get service %s: %w", svcName, err)
		}
		svcs[ind] = svc.Name
	}
	return svcs, nil
}

type result struct {
	name string
	err  error
}

func (s *Store) deployedServices(rgClient resourceGetter, app, env, svc string) result {
	resources, err := rgClient.GetResourcesByTags(ecsServiceResourceType, map[string]string{
		AppTagKey:     app,
		EnvTagKey:     env,
		ServiceTagKey: svc,
	})
	if err != nil {
		return result{err: fmt.Errorf("get resources by Copilot tags: %w", err)}
	}
	// If no resources found, the resp length is 0.
	var res result
	if len(resources) != 0 {
		res.name = env
	}
	return res
}

// ListEnvironmentsDeployedTo returns all the environment that a service is deployed in.
func (s *Store) ListEnvironmentsDeployedTo(appName string, svcName string) ([]string, error) {
	envs, err := s.configStore.ListEnvironments(appName)
	if err != nil {
		return nil, fmt.Errorf("list environment for app %s: %w", appName, err)
	}
	deployedEnv := make(chan result, len(envs))
	defer close(deployedEnv)
	for _, env := range envs {
		go func(env *config.Environment) {
			rgClient, err := s.newRgClientFromRole(env.ManagerRoleARN, env.Region)
			if err != nil {
				deployedEnv <- result{err: err}
				return
			}
			deployedEnv <- s.deployedServices(rgClient, appName, env.Name, svcName)
		}(env)
	}
	var envsWithDeployment []string
	for i := 0; i < len(envs); i++ {
		env := <-deployedEnv
		if env.err != nil {
			return nil, env.err
		}
		if env.name != "" {
			envsWithDeployment = append(envsWithDeployment, env.name)
		}
	}
	return envsWithDeployment, nil
}

// IsServiceDeployed returns whether a service is deployed in an environment or not.
func (s *Store) IsServiceDeployed(appName string, envName string, svcName string) (bool, error) {
	rgClient, err := s.newRgClientFromIDs(appName, envName)
	if err != nil {
		return false, err
	}
	svcARNs, err := rgClient.GetResourcesByTags(ecsServiceResourceType, map[string]string{
		AppTagKey:     appName,
		EnvTagKey:     envName,
		ServiceTagKey: svcName,
	})
	if err != nil {
		return false, fmt.Errorf("get resources by Copilot tags: %w", err)
	}
	if len(svcARNs) != 0 {
		return true, nil
	}
	return false, nil
}
