# Makefile for sadp - Hikvision Device Discovery Tool

ROOT_FOLDER         := $(shell git rev-parse --show-toplevel)
include             $(ROOT_FOLDER)/.scripts/golang.mk
include             $(ROOT_FOLDER)/.scripts/markdown.mk

SHA1                := $(shell git rev-parse --verify HEAD)
SHA1_SHORT          := $(shell git rev-parse --verify --short HEAD)
PWD                 := $(shell pwd)

# Project settings
PROJECT             ?= sadp
BINARY              := sadp
BINARY_GUI          := sadp-gui
MAIN_PATH           := ./cmd/sadp
MAIN_PATH_GUI       := ./cmd/sadp-gui
MODULE              := github.com/cameronnewman/hikvision-tooling

#
# Default Goals
#
.DEFAULT_GOAL       := help

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Returns a list of all the make goals
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: version
version: ## Returns version for build
	@echo "Build Version: v$(VERSION_HASH)"

#
# Alias targets (map to go-* targets from golang.mk)
#

.PHONY: build
build: go-build ## Build the binary

.PHONY: install
install: go-install ## Install binary to GOBIN

.PHONY: release
release: go-release ## Build for all platforms

.PHONY: test
test: go-test ## Run tests

.PHONY: test-cover
test-cover: go-test-cover ## Run tests with coverage

.PHONY: test-cover-html
test-cover-html: go-test-cover-html ## Generate HTML coverage report

.PHONY: lint
lint: go-lint ## Run linters

.PHONY: lint-fix
lint-fix: go-lint-fix ## Run linters with auto-fix

.PHONY: fmt
fmt: go-fmt ## Format Go code

.PHONY: fmt-check
fmt-check: go-fmt-check ## Check if code is formatted

.PHONY: mod
mod: go-mod ## Run go mod tidy

.PHONY: vet
vet: go-vet ## Run go vet

.PHONY: clean
clean: go-clean ## Remove build artifacts

.PHONY: check
check: fmt-check vet lint test ## Run all checks

.PHONY: run
run: build ## Build and run
	./$(BUILD_DIR)/$(BINARY)

.PHONY: build-gui
build-gui: ## Build the GUI binary for current platform
	@echo "+++ $(shell date) - Running 'go build' for GUI"

ifeq ($(filter $(ENVIRONMENT),local docker),$(ENVIRONMENT))
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_GUI) $(MAIN_PATH_GUI)
else
	@mkdir -p $(BUILD_DIR)
	DOCKER_BUILDKIT=1 \
	$(DOCKER) run --rm \
	-v $(PWD):/usr/src/app \
	-w /usr/src/app \
	--entrypoint=bash \
	$(GOLANG_BUILD_IMAGE) \
	-c "apt-get update && apt-get install -y libgl1-mesa-dev xorg-dev && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildvcs=false $(GOFLAGS) -ldflags '-s -w -X main.Version=$(VERSION) -X main.Commit=$(VERSION_HASH)' -o $(BUILD_DIR)/$(BINARY_GUI) $(MAIN_PATH_GUI)"
endif

	@echo "$(shell date) - Completed 'go build' for GUI: $(BUILD_DIR)/$(BINARY_GUI)"

.PHONY: release-gui
release-gui: ## Build GUI for Windows and Linux
	@echo "+++ $(shell date) - Building GUI release binaries for Windows and Linux..."

ifeq ($(filter $(ENVIRONMENT),local docker),$(ENVIRONMENT))
	@mkdir -p $(BUILD_DIR)
	@echo "Building Linux amd64..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_GUI)-linux-amd64 $(MAIN_PATH_GUI)
	@echo "Building Windows amd64..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_GUI)-windows-amd64.exe $(MAIN_PATH_GUI)
else
	@mkdir -p $(BUILD_DIR)
	@echo "Building Linux amd64..."
	DOCKER_BUILDKIT=1 \
	$(DOCKER) run --rm \
	-v $(PWD):/usr/src/app \
	-w /usr/src/app \
	--entrypoint=bash \
	$(GOLANG_BUILD_IMAGE) \
	-c "apt-get update && apt-get install -y libgl1-mesa-dev xorg-dev && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -buildvcs=false -ldflags '-s -w -X main.Version=$(VERSION) -X main.Commit=$(VERSION_HASH)' -o $(BUILD_DIR)/$(BINARY_GUI)-linux-amd64 $(MAIN_PATH_GUI)"
	@echo "Building Windows amd64..."
	DOCKER_BUILDKIT=1 \
	$(DOCKER) run --rm \
	-v $(PWD):/usr/src/app \
	-w /usr/src/app \
	--entrypoint=bash \
	$(GOLANG_BUILD_IMAGE) \
	-c "apt-get update && apt-get install -y gcc-mingw-w64-x86-64 libgl1-mesa-dev xorg-dev && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -buildvcs=false -ldflags '-s -w -X main.Version=$(VERSION) -X main.Commit=$(VERSION_HASH) -H windowsgui' -o $(BUILD_DIR)/$(BINARY_GUI)-windows-amd64.exe $(MAIN_PATH_GUI)"
endif

	@echo "$(shell date) - Completed GUI release builds in $(BUILD_DIR)/"

.PHONY: run-gui
run-gui: build-gui ## Build and run GUI
	./$(BUILD_DIR)/$(BINARY_GUI)
