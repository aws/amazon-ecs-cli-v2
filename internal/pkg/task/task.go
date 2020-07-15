// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package task provides support for running Amazon ECS tasks.
package task

import (
	"errors"
	"fmt"
	"github.com/aws/copilot-cli/internal/pkg/aws/ec2"
	"github.com/aws/copilot-cli/internal/pkg/aws/ecs"
	"github.com/aws/copilot-cli/internal/pkg/aws/resourcegroups"
)

// VpcGetter gets subnets and security groups.
type VPCGetter interface {
	GetSubnetIDs(filters ...ec2.Filter) ([]string, error)
	GetSecurityGroups(filters ...ec2.Filter) ([]string, error)
}

// ResourceGetter gets resources by tags.
type ResourceGetter interface {
	GetResourcesByTags(resourceType string, tags map[string]string) ([]*resourcegroups.Resource, error)
}

// DefaultClusterGetter gets the default cluster.
type DefaultClusterGetter interface {
	DefaultCluster() (string, error)
}

// TaskRunner runs the tasks and wait for it to start.
type TaskRunner interface {
	RunTask(input ecs.RunTaskInput) ([]string, error)
}

const (
	startedBy           = "copilot-task"
)

var (
	errNoSubnetFound     = errors.New("no subnets found")
	fmtTaskFamilyName    = "copilot-%s"
)

func taskFamilyName(groupName string) string{
	return fmt.Sprintf(fmtTaskFamilyName, groupName)
}
