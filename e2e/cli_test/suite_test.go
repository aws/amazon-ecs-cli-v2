// +build e2e isolated

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// This suite of e2e tests do not require additional intrastructure setup
// nor do they interact with AWS services so it's ok to run them as part of
// the normal test target.
package cli_test

import (
	"flag"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestHelpMessages(t *testing.T) {
	RegisterFailHandler(Fail)
	rand.Seed(time.Now().UnixNano())
	RunSpecs(t, "Test help messages displayed when running various archer commands")
}

var cliPath string
var update = flag.Bool("update", false, "update .golden files")

var _ = BeforeSuite(func() {
	// ensure the e2e tests are performed on the latest code changes by
	// compiling CLI from source
	var err error
	cliPath, err = gexec.Build("../../cmd/archer/main.go")
	Expect(err).Should(BeNil())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
