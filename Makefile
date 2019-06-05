.DEFAULT_GOAL := local

ifndef VERSION
	VERSION = ${shell git describe --tags --always}
endif

FLAGS = -ldflags '-X github.com/kyma-project/cli/pkg/kyma/cmd.Version=$(VERSION)'

.PHONY: resolve
resolve: 
	dep ensure -vendor-only -v

.PHONY: validate
validate:
	go build -o golint-vendored ./vendor/github.com/golang/lint/golint
	./golint-vendored
	rm golint-vendored

.PHONY: build
build:
	go generate ./...
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/kyma.exe $(FLAGS) ./cmd/kyma
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/kyma-linux $(FLAGS) ./cmd/kyma
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/kyma-darwin $(FLAGS) ./cmd/kyma

.PHONY: test
test:
	go test ./...

.PHONY: integration-test
integration-test:
	./bin/kyma-linux help

.PHONY: archive
archive:
	cp -r bin/* $(ARTIFACTS)

.PHONY: release
release:
	export GITHUB_TOKEN="$(BOT_GITHUB_TOKEN)"
	curl -sL https://git.io/goreleaser | bash

.PHONY: clean
clean:
	rm -rf bin

.PHONY: install
install:
	go generate ./...
	go install $(FLAGS) ./cmd/kyma

.PHONY: local
local: validate test install

.PHONY: ci-pr
ci-pr: resolve validate build test integration-test

.PHONY: ci-master
ci-master: resolve validate build test integration-test

.PHONY: ci-release
ci-release: resolve validate build test integration-test archive release

