// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package codestar provides a client to make API requests to AWS CodeStar Connections.
package codestar

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
)

type api interface {
	GetConnection(input *codestarconnections.GetConnectionInput) (*codestarconnections.GetConnectionOutput, error)
}

// Connection represents a client to make requests to AWS CodeStarConnections.
type CodeStar struct {
	client api
}

// New creates a new CloudFormation client.
func New(s *session.Session) *CodeStar {
	return &CodeStar{
		codestarconnections.New(s),
	}
}

// WaitUntilStatusAvailable blocks until the connection status has been updated from `PENDING` to `AVAILABLE` or until the max attempt window expires.
func (c *CodeStar) WaitUntilStatusAvailable(ctx context.Context, connectionARN string) error {
	var interval time.Duration // Default to 0.
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for connection %s status to change from PENDING to AVAILABLE", connectionARN)
		case <-time.After(interval):
			output, err := c.client.GetConnection(&codestarconnections.GetConnectionInput{ConnectionArn: aws.String(connectionARN)})
			if err != nil {
				return fmt.Errorf("get connection details: %w", err)
			}
			if aws.StringValue(output.Connection.ConnectionStatus) == codestarconnections.ConnectionStatusAvailable {
				return nil
			}
			interval = 5 * time.Second
		}
	}
}
