SHELL = /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# Variables
VERSION?=latest

##@ General

# help target is based on https://github.com/operator-framework/operator-sdk/blob/master/release/Makefile.
.DEFAULT_GOAL := help
help: ## Show this help screen.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help

.PHONY: lint-shell

##@ Building

images: image-helm ## Build images.
.PHONY: images

image-helm: ## Build helm image.
	docker build \
        -f build/images/Dockerfile.helm \
		-t localhost:5000/ods-pipeline/helm \
		.
.PHONY: image-helm

tasks: ## Render tasks. Use VERSION=1.0.0 make tasks to render specific version.
	go run github.com/opendevstack/ods-pipeline/cmd/taskmanifest \
		-data ImageRepository=ghcr.io/opendevstack/ods-pipeline-go \
		-data Version=$(VERSION) \
		-template build/tasks/deploy.yaml \
		-destination tasks/deploy.yaml
.PHONY: tasks

docs: tasks ## Render documentation for tasks.
	go run github.com/opendevstack/ods-pipeline/cmd/taskdoc \
		-task tasks/deploy.yaml \
		-description build/docs/deploy.adoc \
		-destination docs/deploy.adoc
.PHONY: docs

##@ Testing

test: test-cmd test-e2e ## Run complete testsuite.
.PHONY: test

test-cmd: ## Run cmd tests.
	go test -v -count=1 -timeout 10m ./cmd/...
.PHONY: test-cmd

test-e2e: ## Run testsuite of end-to-end task runs.
	go test -v -count=1 -timeout 10m ./test/e2e/...
.PHONY: test-e2e

##@ CI

check-docs: docs ## Check docs are up-to-date
	@printf "Docs / tasks are " && git diff --quiet docs tasks && echo "up-to-date." || (echo "not up-to-date! Run 'make docs' to update."; false)
.PHONY: check-docs

ci: check-docs test ## Run CI tasks
.PHONY: ci
