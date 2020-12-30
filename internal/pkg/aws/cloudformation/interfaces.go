// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type changeSetAPI interface {
	CreateChangeSet(*cloudformation.CreateChangeSetInput) (*cloudformation.CreateChangeSetOutput, error)
	WaitUntilChangeSetCreateCompleteWithContext(aws.Context, *cloudformation.DescribeChangeSetInput, ...request.WaiterOption) error
	DescribeChangeSet(*cloudformation.DescribeChangeSetInput) (*cloudformation.DescribeChangeSetOutput, error)
	ExecuteChangeSet(*cloudformation.ExecuteChangeSetInput) (*cloudformation.ExecuteChangeSetOutput, error)
	DeleteChangeSet(*cloudformation.DeleteChangeSetInput) (*cloudformation.DeleteChangeSetOutput, error)
}

type client interface {
	changeSetAPI

	DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	DescribeStackEvents(*cloudformation.DescribeStackEventsInput) (*cloudformation.DescribeStackEventsOutput, error)
	GetTemplate(input *cloudformation.GetTemplateInput) (*cloudformation.GetTemplateOutput, error)
	DeleteStack(*cloudformation.DeleteStackInput) (*cloudformation.DeleteStackOutput, error)

	ListStacks(*cloudformation.ListStacksInput) (*cloudformation.ListStacksOutput, error)

	WaitUntilStackCreateCompleteWithContext(aws.Context, *cloudformation.DescribeStacksInput, ...request.WaiterOption) error
	WaitUntilStackUpdateCompleteWithContext(aws.Context, *cloudformation.DescribeStacksInput, ...request.WaiterOption) error
	WaitUntilStackDeleteCompleteWithContext(aws.Context, *cloudformation.DescribeStacksInput, ...request.WaiterOption) error
}
