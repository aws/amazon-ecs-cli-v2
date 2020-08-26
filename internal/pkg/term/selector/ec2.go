// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package selector provides functionality for users to select an application, environment, or service name.
package selector

import (
	"errors"
	"fmt"

	"github.com/aws/copilot-cli/internal/pkg/aws/ec2"
)

var (
	// ErrVPCNotFound is returned when no existing VPCs are found.
	ErrVPCNotFound = errors.New("no existing VPCs found")
	// ErrSubnetsNotFound is returned when no existing subnets are found.
	ErrSubnetsNotFound = errors.New("no existing subnets found")
)

// VPCSubnetLister list VPCs and subnets.
type VPCSubnetLister interface {
	ListVPCs() ([]ec2.VPC, error)
	ListVPCSubnets(vpcID string, opts ...ec2.ListVPCSubnetsOpts) ([]string, error)
}

// EC2Select is a selector for Ec2 resources.
type EC2Select struct {
	prompt Prompter
	ec2Svc VPCSubnetLister
}

// NewEC2Select returns a new selector that chooses Ec2 resources.
func NewEC2Select(prompt Prompter, ec2Client VPCSubnetLister) *EC2Select {
	return &EC2Select{
		prompt: prompt,
		ec2Svc: ec2Client,
	}
}

// VPC has the user select an available VPC.
func (s *EC2Select) VPC(prompt, help string) (string, error) {
	vpcs, err := s.ec2Svc.ListVPCs()
	if err != nil {
		return "", fmt.Errorf("list VPC ID: %w", err)
	}
	if len(vpcs) == 0 {
		return "", ErrVPCNotFound
	}
	var options []string
	for _, vpc := range vpcs {
		stringifiedVPC := vpc.String()
		options = append(options, stringifiedVPC)
	}
	vpc, err := s.prompt.SelectOne(
		prompt, help,
		options)
	if err != nil {
		return "", fmt.Errorf("select VPC: %w", err)
	}
	extractedVPC, err := ec2.ExtractVPC(vpc)
	if err != nil {
		return "", fmt.Errorf("extract VPC ID: %w", err)
	}
	return extractedVPC.ID, nil
}

// PublicSubnets has the user multiselect public subnets given the VPC ID.
func (s *EC2Select) PublicSubnets(prompt, help, vpcID string) ([]string, error) {
	return s.subnet(prompt, help, vpcID, ec2.FilterForPublicSubnets())
}

// PrivateSubnets has the user multiselect private subnets given the VPC ID.
func (s *EC2Select) PrivateSubnets(prompt, help, vpcID string) ([]string, error) {
	return s.subnet(prompt, help, vpcID, ec2.FilterForPrivateSubnets())
}

func (s *EC2Select) subnet(prompt, help string, vpcID string, filter ec2.ListVPCSubnetsOpts) ([]string, error) {
	subnets, err := s.ec2Svc.ListVPCSubnets(vpcID, filter)
	if err != nil {
		return nil, fmt.Errorf("list subnets for VPC %s: %w", vpcID, err)
	}
	if len(subnets) == 0 {
		return nil, ErrSubnetsNotFound
	}
	ans, err := s.prompt.MultiSelect(
		prompt, help,
		subnets)
	if err != nil {
		return nil, err
	}
	return ans, nil
}
