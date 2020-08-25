// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package addon contains the service to manage addons.
package addon

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/copilot-cli/internal/pkg/template"
)

const (
	dynamoDbAddonPath = "addons/ddb/cf.yml"
	s3AddonPath       = "addons/s3/cf.yml"
)

var regexpMatchAttribute = regexp.MustCompile(`^(\S+):([sbnSBN])`)

var storageTemplateFunctions = map[string]interface{}{
	"logicalIDSafe": template.StripNonAlphaNumFunc,
	"envVarName":    template.EnvVarNameFunc,
}

// DynamoDB contains configuration options which fully descibe a DynamoDB table.
// Implements the encoding.BinaryMarshaler interface.
type DynamoDB struct {
	DynamoDBProps

	parser template.Parser
}

// S3 contains configuration options which fully describe an S3 bucket.
// Implements the encoding.BinaryMarshaler interface.
type S3 struct {
	S3Props

	parser template.Parser
}

// StorageProps holds basic input properties for addon.NewDynamoDB() or addon.NewS3().
type StorageProps struct {
	Name string
}

// S3Props contains S3-specific properties for addon.NewS3().
type S3Props struct {
	*StorageProps
}

// DynamoDBProps contains DynamoDB-specific properties for addon.NewDynamoDB().
type DynamoDBProps struct {
	*StorageProps
	Attributes   []DDBAttribute
	LSIs         []DDBLocalSecondaryIndex
	SortKey      *string
	PartitionKey *string
	HasLSI       bool
}

// DDBAttribute holds the attribute definition of a DynamoDB attribute (keys, local secondary indices).
type DDBAttribute struct {
	Name     *string
	DataType *string // Must be one of "N", "S", "B"
}

// DDBLocalSecondaryIndex holds a representation of an LSI.
type DDBLocalSecondaryIndex struct {
	PartitionKey *string
	SortKey      *string
	Name         *string
}

// MarshalBinary serializes the DynamoDB object into a binary YAML CF template.
// Implements the encoding.BinaryMarshaler interface.
func (d *DynamoDB) MarshalBinary() ([]byte, error) {
	content, err := d.parser.Parse(dynamoDbAddonPath, *d, template.WithFuncs(storageTemplateFunctions))
	if err != nil {
		return nil, err
	}
	return content.Bytes(), nil
}

// NewDynamoDB creates a DynamoDB cloudformation template specifying attributes,
// primary key schema, and local secondary index configuration.
func NewDynamoDB(input *DynamoDBProps) *DynamoDB {
	return &DynamoDB{
		DynamoDBProps: *input,

		parser: template.New(),
	}
}

// MarshalBinary serializes the S3 object into a binary YAML CF template.
// Implements the encoding.BinaryMarshaler interface.
func (s *S3) MarshalBinary() ([]byte, error) {
	content, err := s.parser.Parse(s3AddonPath, *s, template.WithFuncs(storageTemplateFunctions))
	if err != nil {
		return nil, err
	}
	return content.Bytes(), nil
}

// NewS3 creates a new S3 marshaler which can be used to write CF via addonWriter.
func NewS3(input *S3Props) *S3 {
	return &S3{
		S3Props: *input,

		parser: template.New(),
	}
}

// BuildPartitionKey generates the properties required to specify the partition key
// based on customer inputs.
func (p *DynamoDBProps) BuildPartitionKey(partitionKey string) error {
	partitionKeyAttribute, err := DDBAttributeFromKey(partitionKey)
	if err != nil {
		return err
	}
	p.Attributes = append(p.Attributes, partitionKeyAttribute)
	p.PartitionKey = partitionKeyAttribute.Name
	return nil
}

// BuildSortKey generates the correct property configuration based on customer inputs.
func (p *DynamoDBProps) BuildSortKey(noSort bool, sortKey string) (bool, error) {
	if noSort || sortKey == "" {
		return false, nil
	}
	sortKeyAttribute, err := DDBAttributeFromKey(sortKey)
	if err != nil {
		return false, err
	}
	p.Attributes = append(p.Attributes, sortKeyAttribute)
	p.SortKey = sortKeyAttribute.Name
	return true, nil
}

// BuildLocalSecondaryIndex generates the correct LocalSecondaryIndex property configuration
// based on customer input to ensure that the CF template is valid. BuildLocalSecondaryIndex
// should be called last, after BuildPartitionKey && BuildSortKey
func (p *DynamoDBProps) BuildLocalSecondaryIndex(noLSI bool, lsiSorts []string) (bool, error) {
	// If there isn't yet a partition key on the struct, we can't do anything with this call.
	if p.PartitionKey == nil {
		return false, fmt.Errorf("partition key not specified")
	}
	// If a sort key hasn't been specified, or the customer has specified that there is no LSI,
	// or there is implicitly no LSI based on the input lsiSorts value, do nothing.
	if p.SortKey == nil || noLSI || len(lsiSorts) == 0 {
		p.HasLSI = false
		return false, nil
	}
	for _, att := range lsiSorts {
		currAtt, err := DDBAttributeFromKey(att)
		if err != nil {
			return false, err
		}
		p.Attributes = append(p.Attributes, currAtt)
	}
	p.HasLSI = true
	lsiConfig, err := newLSI(*p.PartitionKey, lsiSorts)
	if err != nil {
		return false, err
	}
	p.LSIs = lsiConfig
	return true, nil
}

// DDBAttributeFromKey parses the DDB type and name out of keys specified in the form "Email:S"
func DDBAttributeFromKey(input string) (DDBAttribute, error) {
	attrs := regexpMatchAttribute.FindStringSubmatch(input)
	if len(attrs) == 0 {
		return DDBAttribute{}, fmt.Errorf("parse attribute from key: %s", input)
	}
	upperString := strings.ToUpper(attrs[2])
	return DDBAttribute{
		Name:     &attrs[1],
		DataType: &upperString,
	}, nil
}

func newLSI(partitionKey string, lsis []string) ([]DDBLocalSecondaryIndex, error) {
	var output []DDBLocalSecondaryIndex
	for _, lsi := range lsis {
		lsiAttr, err := DDBAttributeFromKey(lsi)
		if err != nil {
			return nil, err
		}
		output = append(output, DDBLocalSecondaryIndex{
			PartitionKey: &partitionKey,
			SortKey:      lsiAttr.Name,
			Name:         lsiAttr.Name,
		})
	}
	return output, nil
}
