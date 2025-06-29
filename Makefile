include common.mk
include tools.mk

LDFLAGS += -X "$(MODULE)/version.Version=$(VERSION)" -X "$(MODULE)/version.CommitSHA=$(VERSION_HASH)"

SITE_DOCKER_FLAGS = \
	-v $(CURDIR)/www:/docs \
	-v $(CURDIR)/LICENSE:/docs/docs/LICENSE \
	-v $(CURDIR)/SECURITY.md:/docs/docs/security.md \
	-v $(CURDIR)/CHANGELOG.md:/docs/docs/changelog.md \
	-v $(CURDIR)/CODE-OF-CONDUCT.md:/docs/docs/code-of-conduct.md \
	-v $(CURDIR)/CONTRIBUTING.md:/docs/docs/contributing.md

## Build:

.PHONY: build
build: | build-frontend build-backend ## Build binary

.PHONY: build-frontend
build-frontend: ## Build frontend
	$Q cd frontend && pnpm install --frozen-lockfile && pnpm run build

.PHONY: build-backend
build-backend: ## Build backend
	$Q $(go) build -ldflags '$(LDFLAGS)' -o .

.PHONY: test
test: | test-frontend test-backend ## Run all tests

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	$Q cd frontend && pnpm install --frozen-lockfile && pnpm run typecheck

.PHONY: test-backend
test-backend: ## Run backend tests
	$Q $(go) test -v ./...

.PHONY: lint
lint: lint-frontend lint-backend ## Run all linters

.PHONY: lint-frontend
lint-frontend: ## Run frontend linters
	$Q cd frontend && pnpm install --frozen-lockfile && pnpm run lint

.PHONY: lint-backend
lint-backend: | $(golangci-lint) ## Run backend linters
	$Q $(golangci-lint) run -v

.PHONY: lint-commits
lint-commits: $(commitlint) ## Run commit linters
	$Q ./scripts/commitlint.sh

fmt: $(goimports) ## Format source files
	$Q $(goimports) -local $(MODULE) -w $$(find . -type f -name '*.go' -not -path "./vendor/*")

clean: clean-tools ## Clean

## Release:

.PHONY: bump-version
bump-version: $(standard-version) ## Bump app version
	$Q ./scripts/bump_version.sh

.PHONY: site
site: ## Build site
	@rm -rf www/public
	docker build -f www/Dockerfile --progress=plain -t filebrowser.site www
	docker run --rm $(SITE_DOCKER_FLAGS) filebrowser.site build -d "public"

.PHONY: site-serve
site-serve: ## Serve site for development
	docker build -f www/Dockerfile --progress=plain -t filebrowser.site www
	docker run --rm -it -p 8000:8000 $(SITE_DOCKER_FLAGS) filebrowser.site

## Help:
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target> [options]${RESET}'
	@echo ''
	@echo 'Options:'
	@$(call global_option, "V [0|1]", "enable verbose mode (default:0)")
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
