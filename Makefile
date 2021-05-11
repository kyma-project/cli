.DEFAULT_GOAL := local

ifndef KYMA_VERSION
	KYMA_VERSION = 1.22.0
endif

ifndef VERSION
	VERSION = ${shell git describe --tags --always}
endif

ifeq ($(VERSION),stable)
	VERSION = stable-${shell git rev-parse --short HEAD}
endif

FLAGS = -ldflags '-s -w -X github.com/kyma-project/cli/cmd/kyma/version.Version=$(VERSION) -X github.com/kyma-project/cli/cmd/kyma/install.DefaultKymaVersion=$(KYMA_VERSION) -X github.com/kyma-project/cli/cmd/kyma/upgrade.DefaultKymaVersion=$(KYMA_VERSION)'

.PHONY: resolve
resolve:
	go mod tidy

.PHONY: validate
validate:
	./hack/verify-lint.sh
	./hack/verify-generated-docs.sh

.PHONY: build
build: build-windows build-linux build-darwin build-windows-arm build-linux-arm

.PHONY: build-windows
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/kyma.exe $(FLAGS) ./cmd

.PHONY: build-windows-arm
build-windows-arm:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm go build -o ./bin/kyma.exe $(FLAGS) ./cmd

.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/kyma-linux $(FLAGS) ./cmd

.PHONY: build-linux-arm
build-linux-arm:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/kyma-linux-arm $(FLAGS) ./cmd

.PHONY: build-darwin
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/kyma-darwin $(FLAGS) ./cmd

# .PHONY: build-darwin-arm
# build-darwin-arm:
# 	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/kyma-darwin-arm $(FLAGS) ./cmd

.PHONY: docs
docs:
	rm -f ./docs/gen-docs/*
	go run ./cmd/gendocs/gendocs.go

.PHONY: test
test:
	go test -race -coverprofile=cover.out ./...
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
	./hack/release.sh

.PHONY: clean
clean:
	rm -rf bin

.PHONY: install
install:
	CGO_ENABLED=0 go build -o ./bin/kyma $(FLAGS) ./cmd
	mv ./bin/kyma ${GOPATH}/bin

.PHONY: local
local: validate test install

.PHONY: ci-pr
ci-pr: resolve validate build test integration-test

.PHONY: ci-main
ci-main: resolve validate build test integration-test upload-stable

.PHONY: ci-release
ci-release: resolve validate build test integration-test archive release

