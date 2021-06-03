SHELL := /bin/bash
BASE_PATH := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

BIN = $(BASE_PATH)/bin
PATH := $(BIN):$(PATH)
export PATH

# printing
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

GO = GOGC=off go
# go module
MODULE = $(shell env GO111MODULE=on $(GO) list -m)

DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell git describe --tags --always --match=v* 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)
VERSION_HASH = 	$(shell git rev-parse --short HEAD)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS += -X "$(MODULE)/version.Version=$(VERSION)" -X "$(MODULE)/version.CommitSHA=$(VERSION_HASH)"

# tools
$(BIN):
	@mkdir -p $@
$(BIN)/%: | $(BIN) ; $(info $(M) installing $(PACKAGE)…)
	$Q env GOBIN=$(BIN) $(GO) install $(PACKAGE)

GOLANGCI_LINT = $(BIN)/golangci-lint
$(BIN)/golangci-lint: PACKAGE=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.37.1

GOIMPORTS = $(BIN)/goimports
$(BIN)/goimports: PACKAGE=golang.org/x/tools/cmd/goimports@v0.1.0

## build: Build
.PHONY: build
build: | build-frontend build-backend ; $(info $(M) building…)

## build-frontend: Build frontend
.PHONY: build-frontend
build-frontend: | ; $(info $(M) building frontend…)
	$Q cd frontend && npm ci && npm run build

## build-backend: Build backend
.PHONY: build-backend
build-backend: | ; $(info $(M) building backend…)
	$Q $(GO) build -ldflags '$(LDFLAGS)' -o .

## test: Run all tests
.PHONY: test
test: | test-frontend test-backend ; $(info $(M) running tests…)

## test-frontend: Run frontend tests
.PHONY: test-frontend
test-frontend: | ; $(info $(M) running frontend tests…)

## test-backend: Run backend tests
.PHONY: test-backend
test-backend: | ; $(info $(M) running backend tests…)
	$Q $(GO) test -v ./...

## lint: Lint
.PHONY: lint
lint: lint-frontend lint-backend lint-commits | ; $(info $(M) running all linters…)

## lint-frontend: Lint frontend
.PHONY: lint-frontend
lint-frontend: | ; $(info $(M) running frontend linters…)
	$Q cd frontend && npm ci && npm run lint

## lint-backend: Lint backend
.PHONY: lint-backend
lint-backend: | $(GOLANGCI_LINT) ; $(info $(M) running backend linters…)
	$Q $(GOLANGCI_LINT) run

## lint-commits: Lint commits
.PHONY: lint-commits
lint-commits: | ; $(info $(M) running commitlint…)
	$Q ./scripts/commitlint.sh

## bump-version: Bump app version
.PHONY: bump-version
bump-version: | ; $(info $(M) creating a new release…)
	$Q ./scripts/bump_version.sh

## help: Show this help
.PHONY: help
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /' | sort

.PHONY: build-release-bin
build-release-bin: build-frontend
	GO111MODULE=on GOOS=linux GOARCH=amd64 $(GO) build -ldflags '$(LDFLAGS)' -o bin/filebrowser-$(VERSION)
	tar -C bin -czf "dist/filebrowser-$(VERSION).tar.gz" "filebrowser-$(VERSION)"
