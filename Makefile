# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
PROJECT_ROOT=.

include ${PROJECT_ROOT}/hack/tools.mk


##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##@ Development
.PHONY: test
test: ## Run unit tests.
	go test `go list ./... | grep -v /tests/e2e` -race -coverprofile=cover.out
	@echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')"
	@rm cover.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run -v

.PHONY: build
build:
	go build -o bin/kyma@v3 main.go   



