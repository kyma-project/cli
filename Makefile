.DEFAULT_GOAL := local

ifndef KYMA_VERSION
	KYMA_VERSION = 1.11.0-rc2
endif

ifndef VERSION
	VERSION = ${shell git describe --tags --always}
endif

FLAGS = -ldflags '-X github.com/kyma-project/cli/cmd/kyma/version.Version=$(VERSION) -X github.com/kyma-project/cli/cmd/kyma/install.DefaultKymaVersion=$(KYMA_VERSION)'

.PHONY: resolve
resolve: 
	go mod tidy

.PHONY: validate
validate:
	./hack/verify-lint.sh
	./hack/verify-generated-docs.sh

.PHONY: build
build: build-windows build-linux build-darwin

.PHONY: build-windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/kyma.exe $(FLAGS) ./cmd

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/kyma-linux $(FLAGS) ./cmd

.PHONY: build-darwin
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/kyma-darwin $(FLAGS) ./cmd

.PHONY: docs
docs:
	go run ./cmd/gendocs/gendocs.go

.PHONY: test
test:
	go test -coverprofile=cover.out ./...
	@echo "Total test coverage: $$(go tool cover -func=cover.out | grep total | awk '{print $$3}')"
	@rm cover.out

.PHONY: integration-test
integration-test:
	./bin/kyma-linux help

.PHONY: archive
archive:
	cp -r bin/* $(ARTIFACTS)

.PHONY: upload-stable
upload-stable: 
ifdef STABLE
	gsutil cp bin/* $(KYMA_CLI_STABLE_BUCKET)
endif

.PHONY: release
release:
	git remote add origin git@github.com:kyma-project/cli.git
	curl -sL https://git.io/goreleaser | VERSION=v0.118.2 bash

.PHONY: clean
clean:
	rm -rf bin

.PHONY: install
install:
	go install $(FLAGS) ./cmd

.PHONY: local
local: validate test install

.PHONY: ci-pr
ci-pr: resolve validate build test integration-test

.PHONY: ci-master
ci-master: resolve validate build test integration-test upload-stable

.PHONY: ci-release
ci-release: resolve validate build test integration-test archive release

