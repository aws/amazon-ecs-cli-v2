// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package manifest provides functionality to create Manifest files.
package manifest

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// LoadBalancedWebServiceType is a web service with a load balancer and Fargate as compute.
	LoadBalancedWebServiceType = "Load Balanced Web Service"
	// BackendServiceType is a service that cannot be accessed from the internet but can be reached from other services.
	BackendServiceType = "Backend Service"
)

// ServiceTypes are the supported service manifest types.
var ServiceTypes = []string{
	LoadBalancedWebServiceType,
	BackendServiceType,
}

// RangeOpts contains either a Range or a range configuration for Autoscaling ranges
type RangeOpts struct {
	Range       *Range // Mutually exclusive with RangeConfig
	RangeConfig RangeConfig
}

// Parse extracts the min and max from RangeOpts
func (r RangeOpts) Parse() (min int, max int, err error) {
	if r.Range != nil && !r.RangeConfig.IsEmpty() {
		return 0, 0, errInvalidRangeOpts
	}

	if r.Range != nil {
		return r.Range.Parse()
	}

	return *r.RangeConfig.Min, *r.RangeConfig.Max, nil
}

// UnmarshalYAML overrides the default YAML unmarshaling logic for the RangeOpts
// struct, allowing it to perform more complex unmarshaling behavior.
// This method implements the yaml.Unmarshaler (v2) interface.
func (r *RangeOpts) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&r.RangeConfig); err != nil {
		switch err.(type) {
		case *yaml.TypeError:
			break
		default:
			return err
		}
	}

	if !r.RangeConfig.IsEmpty() {
		// Unmarshaled successfully to r.RangeConfig, unset r.Range, and return.
		r.Range = nil
		return nil
	}

	if err := unmarshal(&r.Range); err != nil {
		return errUnmarshalRangeOpts
	}
	return nil
}

// Range is a number range with maximum and minimum values.
type Range string

// Parse parses Range string and returns the min and max values.
// For example: 1-100 returns 1 and 100.
func (r Range) Parse() (min int, max int, err error) {
	minMax := strings.Split(string(r), "-")
	if len(minMax) != 2 {
		return 0, 0, fmt.Errorf("invalid range value %s. Should be in format of ${min}-${max}", string(r))
	}
	min, err = strconv.Atoi(minMax[0])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert minimum value %s to integer", minMax[0])
	}
	max, err = strconv.Atoi(minMax[1])
	if err != nil {
		return 0, 0, fmt.Errorf("cannot convert maximum value %s to integer", minMax[1])
	}
	return min, max, nil
}

// RangeConfig containers a Min/Max and an optional SpotFrom field which
// specifies the number of services you want to start placing on spot. For
// example, if your range is 1-10 and `spot_from` is 5, up to 4 services will
// be placed on dedicated Fargate capacity, and then after that, any scaling
// event will place additioanl services on spot capacity.
type RangeConfig struct {
	Min      *int `yaml:"min"`
	Max      *int `yaml:"max"`
	SpotFrom *int `yaml:"spot_from"`
}

// IsEmpty returns whether RangeConfig is empty.
func (r *RangeConfig) IsEmpty() bool {
	return r.Min == nil && r.Max == nil && r.SpotFrom == nil
}

// ServiceImageWithPort represents a container image with an exposed port.
type ServiceImageWithPort struct {
	Image `yaml:",inline"`
	Port  *uint16 `yaml:"port"`
}

// Count is a custom type which supports unmarshaling yaml which
// can either be of type int or type Autoscaling.
type Count struct {
	Value       *int        // 0 is a valid value, so we want the default value to be nil.
	Autoscaling Autoscaling // Mutually exclusive with Value.
}

// UnmarshalYAML overrides the default YAML unmarshaling logic for the Count
// struct, allowing it to perform more complex unmarshaling behavior.
// This method implements the yaml.Unmarshaler (v2) interface.
func (c *Count) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&c.Autoscaling); err != nil {
		switch err.(type) {
		case *yaml.TypeError:
			break
		default:
			return err
		}
	}

	if !c.Autoscaling.IsValid() {
		return errUnmarshalSpot
	}

	if !c.Autoscaling.IsEmpty() {
		// Successfully unmarshalled Autoscaling fields, return
		return nil
	}

	if err := unmarshal(&c.Value); err != nil {
		return errUnmarshalCountOpts
	}
	return nil
}

// Autoscaling represents the configurable options for Auto Scaling as well as
// Capacity configuration (spot).
type Autoscaling struct {
	Spot         *int           `yaml:"spot"` // mutually exclusive with Range
	Range        *RangeOpts     `yaml:"range"`
	CPU          *int           `yaml:"cpu_percentage"`
	Memory       *int           `yaml:"memory_percentage"`
	Requests     *int           `yaml:"requests"`
	ResponseTime *time.Duration `yaml:"response_time"`
}

// IsEmpty returns whether Autoscaling is empty.
func (a *Autoscaling) IsEmpty() bool {
	return a.Range == nil && a.CPU == nil && a.Memory == nil &&
		a.Requests == nil && a.ResponseTime == nil && a.Spot == nil
}

// IgnoreRange returns whether desiredCount is specified on spot capacity
func (a *Autoscaling) IgnoreRange() bool {
	return a.Spot != nil
}

// IsValid checks to make sure Spot fields are compatible with other values in Autoscaling
func (a *Autoscaling) IsValid() bool {
	if a.Spot != nil && a.Range != nil {
		return false
	}
	return true
}

// ServiceDockerfileBuildRequired returns if the service container image should be built from local Dockerfile.
func ServiceDockerfileBuildRequired(svc interface{}) (bool, error) {
	return dockerfileBuildRequired("service", svc)
}
