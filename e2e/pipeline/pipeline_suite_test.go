// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pipeline_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/copilot-cli/e2e/internal/client"
)

// Command-line tools.
var (
	copilot *client.CLI
	aws     *client.AWS
)

// Application identifiers.
var (
	appName  = fmt.Sprintf("e2e-pipeline-%d", time.Now().Unix())
	envNames = []string{"test", "prod"}
	svcName  = "frontend"
)

// CodeCommit credentials.
var (
	repoName          = appName
	codeCommitIAMUser = fmt.Sprintf("%s-cc", appName)
	codeCommitCreds   *client.IAMServiceCreds
)

func TestPipeline(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Suite")
}

var _ = BeforeSuite(func() {
	cli, err := client.NewCLIWithDir(repoName)
	Expect(err).NotTo(HaveOccurred())
	copilot = cli
	aws = client.NewAWS()

	creds, err := aws.CreateCodeCommitIAMUser(codeCommitIAMUser)
	Expect(err).NotTo(HaveOccurred())
	codeCommitCreds = creds
})

var _ = AfterSuite(func() {
	_ = aws.DeleteCodeCommitRepo(appName)
	_ = aws.DeleteCodeCommitIAMUser(codeCommitIAMUser, codeCommitCreds.CredentialID)
})
