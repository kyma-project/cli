.DEFAULT_GOAL := local

APP_NAME = kymactl
IMG = $(DOCKER_PUSH_REPOSITORY)$(DOCKER_PUSH_DIRECTORY)/$(APP_NAME)
FLAGS = -ldflags '-X github.com/kyma-incubator/kymactl/pkg/kymactl/cmd.Version=$(DOCKER_TAG)'


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
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./bin/kymactl.exe $(FLAGS) ./cmd/kymactl.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/kymactl-linux $(FLAGS) ./cmd/kymactl.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/kymactl-darwin $(FLAGS) ./cmd/kymactl.go

.PHONY: test
test:
	go test ./...

.PHONY: integration-test
integration-test:
	./bin/kymactl-linux help

.PHONY: archive
archive:
	cp -r bin/* $(ARTIFACTS)

.PHONY: clean
clean:
	rm -rf bin



.PHONY: local
local: validate build test

.PHONY: ci-pr
ci-pr: resolve validate build test integration-test archive

.PHONY: ci-master
ci-master: resolve validate build test integration-test archive

.PHONY: ci-release
ci-release: resolve validate build test integration-test archive


