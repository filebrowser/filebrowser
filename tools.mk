include common.mk

# tools
TOOLS_DIR := $(BASE_PATH)/tools
TOOLS_GO_DEPS := $(TOOLS_DIR)/go.mod $(TOOLS_DIR)/go.sum
TOOLS_BIN := $(TOOLS_DIR)/bin
$(eval $(shell mkdir -p $(TOOLS_BIN)))
PATH := $(TOOLS_BIN):$(PATH)
export PATH

.PHONY: clean-tools
clean-tools:
	$Q rm -rf $(TOOLS_BIN)

goimports=$(TOOLS_BIN)/goimports
$(goimports): $(TOOLS_GO_DEPS)
	$Q cd $(TOOLS_DIR) && $(go) build -o $@ golang.org/x/tools/cmd/goimports

golangci-lint=$(TOOLS_BIN)/golangci-lint
$(golangci-lint): $(TOOLS_GO_DEPS)
	$Q cd $(TOOLS_DIR) && $(go) build -o $@ github.com/golangci/golangci-lint/v2/cmd/golangci-lint

# js tools
TOOLS_JS_DEPS=$(TOOLS_DIR)/node_modules/.modified
$(TOOLS_JS_DEPS): $(TOOLS_DIR)/package.json $(TOOLS_DIR)/yarn.lock
	$Q cd ${TOOLS_DIR} && yarn install
	$Q touch -am $@

standard-version=$(TOOLS_BIN)/standard-version
$(standard-version): $(TOOLS_JS_DEPS)
	$Q ln -sf $(TOOLS_DIR)/node_modules/.bin/standard-version $@
	$Q touch -am $@

commitlint=$(TOOLS_BIN)/commitlint
$(commitlint): $(TOOLS_JS_DEPS)
	$Q ln -sf $(TOOLS_DIR)/node_modules/.bin/commitlint $@
	$Q touch -am $@