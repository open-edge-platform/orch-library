# SPDX-FileCopyrightText: (C) 2024 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

GOCMD             := go
PKG               := github.com/open-edge-platform/orch-library/go/dazl
GOLANG_COVER_VERSION = v0.2.0
GOLANG_GOCOVER_COBERTURA_VERSION = v1.2.0
GOPATH := $(shell go env GOPATH)

include ../../../common.mk

## Virtual environment name
VENV_NAME = venv-env

all: generate build lint test
	@# Help: Runs build, lint, test stages

# GO variables
GOARCH	:= $(shell go env GOARCH)
GOCMD   := GOPRIVATE="github.com/open-edge-platform/*" go


go-tidy: ## Run go mod tidy
	$(GOCMD) mod tidy

go-lint-fix: ## Apply automated lint/formatting fixes to go files
	golangci-lint run --fix --config .golangci.yml

go-lint: $(OUT_DIR) ## Run go lint
	golangci-lint --version
	golangci-lint run $(LINT_DIRS) --config .golangci.yml

build: go-tidy go-build


go-build:
	@# Help: Runs build stage
	@echo "---MAKEFILE BUILD---"
	$(GOCMD) build $(PKG)/pkg/...
	@echo "---END MAKEFILE Build---"

go-test:
	@# Help: Runs test stage
	@echo "---MAKEFILE TEST---"
	$(GOCMD) test -race -gcflags=-l `go list  $(PKG)/...`
	@echo "---END MAKEFILE TEST---"

go-cover-dependency:
	go tool cover -V || go install golang.org/x/tools/cmd/cover@${GOLANG_COVER_VERSION}
	go install github.com/boumenot/gocover-cobertura@${GOLANG_GOCOVER_COBERTURA_VERSION}

lint: go-lint mdlint license
	@# Help: Runs lint stage

test: go-tidy go-test coverage
	@# Help: Runs test stage

mdlint: ## Link MD files
	markdownlint --version ;\
	markdownlint "**/*.md" -c ../../../.markdownlint.yml

coverage: go-cover-dependency
	@# Help: Runs coverage stage
	@echo "---MAKEFILE COVERAGE---"
	$(GOCMD) test -gcflags=-l `go list  $(PKG)/...`  -v -coverprofile=coverage.txt -covermode count
	${GOPATH}/bin/gocover-cobertura < coverage.txt > coverage.xml
	#$(GOCMD) tool cover -html=cover.out -o cover.html
	#$(GOCMD) tool cover -func cover.out -o cover.function-coverage.log
	@echo "---END MAKEFILE COVERAGE---"



clean: clean-all
	@# Help: Runs clean-all stage

list: 
	@# Help: displays make targets
	help
