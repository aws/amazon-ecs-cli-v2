# Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

PACKAGES=./pkg... ./internal...

all: build

.PHONY: build
build:
	CGO_ENABLED=0 go build -o ./bin/local/archer ./cmd/archer

.PHONY: test
test:
	go test -v -race -cover -count=1 ${PACKAGES}

.PHONY: integ-test
integ-test:
	go test -v -run Integration -tags integration ${PACKAGES}

.PHONY: gen-mocks
gen-mocks:
	mockgen -source=./pkg/archer/env.go -package=mocks -destination=./mocks/mock_env.go
	mockgen -source=./pkg/archer/project.go -package=mocks -destination=./mocks/mock_project.go
