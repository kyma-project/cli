## Location to install dependencies to
ifndef PROJECT_ROOT
$(error PROJECT_ROOT is undefined)
endif
LOCALBIN ?= $(realpath $(PROJECT_ROOT))/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

GOLANGCI_LINT_VERSION ?= latest
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint

.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint if necessary.
	command -v $(LOCALBIN)/golangci-lint || GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)