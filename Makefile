GO    := GO15VENDOREXPERIMENT=1 go
GOLANGCILINT := golangci-lint
PROMU := $(GOPATH)/bin/promu
pkgs   = $(shell $(GO) list ./... | grep -v /vendor/)
SRC_DIR=github.com/fxinnovation/poweradmin_exporter

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)
DOCKER_REPO             ?= fxinnovation
DOCKER_IMAGE_NAME       ?= poweradmin_exporter
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
LDFLAGS=-ldflags "\
          -X $(SRC_DIR)/information.Version=$(BUILD_VERSION) \
          -X $(SRC_DIR)/information.BuildTime=$(BUILD_TIME) \
          -X $(SRC_DIR)/information.GitCommit=$(GIT_COMMIT) \
          -X $(SRC_DIR)/information.GitDirty=$(GIT_DIRTY) \
          -X $(SRC_DIR)/information.GitDescribe=$(GIT_DESCRIBE)"

all: format build test

clean: ## clean target for cover
	@rm -rf ./target || true
	@mkdir ./target || true

test: build ## running test after build
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

test-cover: style vet ## go test with coverage
	@$(GO) test  $(pkgs) -cover -race -v $(LDFLAGS)

test-coverage: clean style vet ## go test coverage for jenkins
	gocov test $(pkgs) --short -cpu=2 -p=2 -v $(LDFLAGS) | gocov-xml > ./coverage-test.xml


style: ## check code style
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

format: ## Format code
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet: ## vet code
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

dependencies: ## download the dependencies
	rm -rf Gopkg.lock vendor/
	dep ensure

build: promu ## build code with promu
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

tarball: promu ## creates a release tarball
	@echo ">> building release tarball"
	@$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

docker: ## creates docker image
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

dockerlint: ## lints dockerfile
	@echo ">> linting Dockerfile"
	@docker run --rm -i hadolint/hadolint < Dockerfile

promu: ## gets promu for building
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) get -u github.com/prometheus/promu

lint: ## lint code
	@echo ">> linting code"
	@$(GOLANGCILINT) run


help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

setup: ## downloads makefile dependencies
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.DEFAULT_GOAL := help

.PHONY: all style format dependencies build test vet tarball promu