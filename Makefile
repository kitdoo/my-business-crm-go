SHELL = bash
UNAME := $(shell uname)

PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
APP_NAME = my-business-crm
APP_PROJECT = my-business-crm
APP_ENV_PREFIX ?= CRM_
APP_VERSION ?= v$(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
APP_VERSION_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

GOOS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH ?= $(shell uname -m)
GO_RACE_DETECTOR_ENABLE ?= 0
COMPILER_FLAGS ?=

CGO_ENABLE = 0

BIN_OUTPUT_DIR ?= ${PROJECT_ROOT}/build/bin/${APP_VERSION}/${GOOS}-${GOARCH}
LDFLAGS ?= -w -s

ifeq (Darwin,$(UNAME))
LDFLAGS := $(LDFLAGS) -extldflags=-Wl,-ld_classic
endif

DOCKER_IMAGE ?= "ghcr.io/kitdoo/my-business-crm-go"
DOCKER_IMAGE_LATEST_TAG = "latest"

ifeq ($(findstring alpha,$(APP_VERSION)), alpha)
	DOCKER_IMAGE_LATEST_TAG = "dev"
else ifeq ($(findstring beta,$(APP_VERSION)), beta)
	DOCKER_IMAGE_LATEST_TAG = "beta"
else ifeq ($(findstring rc,$(APP_VERSION)), rc)
	DOCKER_IMAGE_LATEST_TAG = "rc"
endif

ifeq ($(findstring -race,$(COMPILER_FLAGS)), -race)
	CGO_ENABLE = 1
else ifeq ($(GO_RACE_DETECTOR_ENABLE), 1)
	CGO_ENABLE = 1
	COMPILER_FLAGS := $(COMPILER_FLAGS) -race
endif

.PHONY: all
all: help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n\033[36m\033[0m"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Remove temporary files
	@rm -rf ${BIN_OUTPUT_DIR}

.PHONY: devtools
devtools: ## Install dev tools (gci, golangci-lint, dupl, gosec, govulncheck, protoc-gen-go, protoc-gen-go-grpc)
	@go install github.com/daixiang0/gci@latest
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	@go install github.com/mibk/dupl@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: fmt
fmt: tidy  ## Run go fmt + gci on all go files
	@gci write \
	  -s standard \
	  -s default \
	  -s "prefix(google.golang.org)" \
	  -s "prefix(golang.org)" \
	  -s "prefix(github.com/altessa-s)" \
	  -s "prefix(github.com/altessa-s/go-atlas)" \
	  -s "prefix(github.com/kitdoo/my-business-crm-go)" \
	  -s blank -s alias \
	 $$(go list -f {{.Dir}} ./...) \

.PHONY: lint
lint: tidy fmt ## Run linter
	golangci-lint run ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	@go mod tidy

.PHONY: generate
generate: ## Generate code by go generate
	@go generate ./...

.PHONY: test
test: ## Run tests with the race detector
	@go test ./... -race -count=1

.PHONY: bench
bench: ## Run benchmarks
	@go test ./... -bench=. -run=^$$ -benchmem

.PHONY: coverage
coverage: ## Generate test coverage report
	@go test ./... -race -count=1 -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: security-scan
security-scan: ## Run gosec + govulncheck
	@gosec ./...
	@govulncheck ./...

.PHONY: build
build: clean tidy ## Build project
	@echo "-------------------------------------------"
	@echo "Build options:"
	@echo PROJECT_ROOT="${PROJECT_ROOT}"
	@echo GOOS="${GOOS}"
	@echo GOARCH="${GOARCH}"
	@echo CGO_ENABLED="${CGO_ENABLE}"
	@echo VERSION="${APP_VERSION}"
	@echo VERSION_COMMIT="${APP_VERSION_COMMIT}"
	@echo ENV_PREFIX="${APP_ENV_PREFIX}"
	@echo LDFLAGS="${LDFLAGS}"
	@echo COMPILER_FLAGS="${COMPILER_FLAGS}"
	@echo BIN_OUTPUT_DIR="${BIN_OUTPUT_DIR}"
	@echo "-------------------------------------------"

	CGO_ENABLED=${CGO_ENABLE} GOOS=${GOOS} GOARCH=${GOARCH} go build -trimpath \
		${COMPILER_FLAGS} \
		-ldflags "${LDFLAGS} \
			-X github.com/altessa-s/go-atlas/core/runtime/appinfo.Name=${APP_NAME} \
			-X github.com/altessa-s/go-atlas/core/runtime/appinfo.Project=${APP_PROJECT} \
			-X github.com/altessa-s/go-atlas/core/runtime/appinfo.EnvPrefix=${APP_ENV_PREFIX} \
			-X github.com/altessa-s/go-atlas/core/runtime/appinfo.Version=${APP_VERSION} \
			-X github.com/altessa-s/go-atlas/core/runtime/appinfo.Commit=${APP_VERSION_COMMIT}" \
		-o ${BIN_OUTPUT_DIR}/${APP_NAME} ${PROJECT_ROOT}/cmd/${APP_NAME}

##@ Proto

.PHONY: proto-lint
proto-lint: ## buf lint (proto/)
	@cd proto && buf lint

.PHONY: proto-build
proto-build: ## Validate .proto files compile (buf build, proto/)
	@cd proto && buf build

.PHONY: proto-format
proto-format: ## buf format -w (proto/)
	@cd proto && buf format -w

.PHONY: proto-breaking
proto-breaking: ## buf breaking against origin/main (proto/)
	@cd proto && buf breaking --against '.git#branch=main,subdir=proto'

.PHONY: proto-clean
proto-clean: ## Remove generated proto output (proto/gen/)
	@rm -rf proto/gen/

.PHONY: proto proto-go proto-js
proto: proto-go proto-js ## Generate all protobuf bindings into proto/gen/

proto-go: ## Generate Go protobuf bindings into proto/gen/go/
	@cd proto && buf generate --template buf.gen.go.yaml

proto-js: ## Generate JavaScript protobuf bindings into proto/gen/js/ (requires protoc-gen-es on PATH: npm i -g @bufbuild/protoc-gen-es)
	@cd proto && NODE_OPTIONS="--no-experimental-webstorage" buf generate --template buf.gen.js.yaml

.PHONY: build-docker-image
build-docker-image: ## Build and push the docker image
ifndef DOCKER_IMAGE
	$(error DOCKER_IMAGE is not set)
endif

	@docker buildx build --push \
		--platform=linux/amd64 \
		--output=type=image,push=true \
		--build-arg APP_NAME="${APP_NAME}" \
		--build-arg APP_VERSION="${APP_VERSION}" \
		--build-arg APP_VERSION_COMMIT="${APP_VERSION_COMMIT}" \
		--build-arg APP_ENV_PREFIX="${APP_ENV_PREFIX}" \
		--build-arg LDFLAGS="${LDFLAGS}" \
		--build-arg COMPILER_FLAGS="${COMPILER_FLAGS}" \
		--build-arg CGO_ENABLE="${CGO_ENABLE}" \
		--build-arg GO_RACE_DETECTOR_ENABLE="${GO_RACE_DETECTOR_ENABLE}" \
		--build-arg BIN_OUTPUT_DIR="${BIN_OUTPUT_DIR}" \
		--tag "${DOCKER_IMAGE}:${APP_VERSION}" \
		--tag "${DOCKER_IMAGE}:${DOCKER_IMAGE_LATEST_TAG}" \
		--label=org.opencontainers.image.title="${APP_NAME}" \
		--label=org.opencontainers.image.revision="${APP_VERSION_COMMIT}" \
		--label=org.opencontainers.image.version="${APP_VERSION}" \
		--label=org.opencontainers.image.created=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		-f ./Dockerfile .
