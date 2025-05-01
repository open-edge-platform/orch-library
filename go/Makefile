# SPDX-FileCopyrightText: (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

GOCMD             := go
PKG               := github.com/open-edge-platform/orch-library/go
GOLANG_COVER_VERSION = v0.2.0
GOLANG_GOCOVER_COBERTURA_VERSION = v1.2.0
GOPATH := $(shell go env GOPATH)

include ../common.mk

## Virtual environment name
VENV_NAME = venv-env

all: generate build lint test
	@# Help: Runs build, lint, test stages

# GO variables
GOARCH	:= $(shell go env GOARCH)
GOCMD   := go


go-tidy: ## Run go mod tidy
	$(GOCMD) mod tidy

go-lint-fix: ## Apply automated lint/formatting fixes to go files
	golangci-lint run --fix --config .golangci.yml

go-lint: $(OUT_DIR) ## Run go lint
	golangci-lint --version
	golangci-lint run $(LINT_DIRS) --config .golangci.yml

build: go-tidy go-build

go-fmt: ## Run go fmt against code.
	go fmt ./...

go-build: go-fmt
	@# Help: Runs build stage
	@echo "---MAKEFILE BUILD---"
	$(GOCMD) build $(PKG)/pkg/...
	@echo "---END MAKEFILE Build---"

go-generate:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.3.0
	go install go.uber.org/mock/mockgen@v0.5.0
	go generate pkg/openidconnect/generate.go
	go generate pkg/openpolicyagent/generate.go

go-generate-mocks:
	go install github.com/vektra/mockery/v2@v2.40.2
	go generate pkg/auth/m2m.go


go-test:
	@# Help: Runs test stage
	@echo "---MAKEFILE TEST---"
	$(GOCMD) test -race -gcflags=-l `go list  $(PKG)/pkg/... | grep -v "/mocks" | grep -v "/test/"`
	@echo "---END MAKEFILE TEST---"

go-cover-dependency:
	go tool cover -V || go install golang.org/x/tools/cmd/cover@${GOLANG_COVER_VERSION}
	go install github.com/boumenot/gocover-cobertura@${GOLANG_GOCOVER_COBERTURA_VERSION}

lint: go-lint mdlint license
	@# Help: Runs lint stage

test: go-tidy go-test coverage
	@# Help: Runs test stage


coverage: go-cover-dependency
	@# Help: Runs coverage stage
	@echo "---MAKEFILE COVERAGE---"
	$(GOCMD) test -gcflags=-l `go list  $(PKG)/pkg/... | grep -v "/mocks" | grep -v "/openpolicyagent" | grep -v "/openidconnect" | grep -v "/northbound" | grep -v "grpc/retry" | grep -v "/test/"`  -v -coverprofile=coverage.txt -covermode count
	${GOPATH}/bin/gocover-cobertura < coverage.txt > coverage.xml
	#$(GOCMD) tool cover -html=cover.out -o cover.html
	#$(GOCMD) tool cover -func cover.out -o cover.function-coverage.log
	@echo "---END MAKEFILE COVERAGE---"



clean: clean-all
	@# Help: Runs clean-all stage

list: 
	@# Help: displays make targets
	help
