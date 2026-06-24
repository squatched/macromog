LUACHECK         := luacheck
BUSTED           := busted
STYLUA           := stylua
LUACOV           := luacov
LUACOV_COBERTURA := luacov-cobertura
GO               := go

LUA_SRC          := macromog.lua lib/
TEST_DIR         := tests/
PLUGIN_COV_MIN   := 80
CLI_COV_MIN      := 80
GO_SRC           := ./...
CLI_DIRS         := ./cmd
CLI_MAIN         := ./cmd/macromog
BINARY           := bin/macromog
BUILD_PLATFORMS  := \
    linux/amd64  linux/386   \
    windows/amd64 windows/386

.DEFAULT_GOAL := help

.PHONY: help \
        validate validate-plugin validate-cli \
        validate-plugin-lint validate-plugin-format \
        validate-plugin-test validate-plugin-coverage \
        validate-cli-lint validate-cli-format validate-cli-tidy validate-cli-test validate-cli-coverage \
        fix fix-plugin-format fix-cli-format fix-cli-tidy \
        build-cli build-cli-all \
        clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} \
	  /^[a-zA-Z_-]+:.*?##/ {printf "  %-28s %s\n", $$1, $$2}' \
	  $(MAKEFILE_LIST)

validate: validate-plugin validate-cli ## Run all validation checks across plugin (Lua) and CLI (Go)

# ── Plugin (Lua / Windower addon) ────────────────────────────────────────────

validate-plugin: validate-plugin-lint validate-plugin-format validate-plugin-coverage ## Run all plugin validation checks

validate-plugin-lint: ## Static analysis with luacheck
	$(LUACHECK) $(LUA_SRC)

validate-plugin-format: ## Formatting check — fails if any file is not stylua-clean
	$(STYLUA) --check $(LUA_SRC)

validate-plugin-test: ## Run test suite without coverage instrumentation (fast, local)
	$(BUSTED) $(TEST_DIR)

validate-plugin-coverage: ## Run tests with coverage and enforce $(PLUGIN_COV_MIN)% threshold
	$(BUSTED) --coverage $(TEST_DIR)
	$(LUACOV)
	$(LUACOV_COBERTURA) -o coverage-plugin.xml
	@awk '/^Total/{gsub(/%/, "", $$NF); cov = $$NF + 0} \
	  END { \
	    printf "Plugin Coverage: %.2f%%\n", cov; \
	    if (cov < $(PLUGIN_COV_MIN)) { \
	      printf "FAIL: %.2f%% is below the %d%% threshold\n", cov, $(PLUGIN_COV_MIN); exit 1 \
	    } else { \
	      printf "PASS: meets %d%% threshold\n", $(PLUGIN_COV_MIN) \
	    } \
	  }' luacov.report.out

fix-plugin-format: ## Auto-format all Lua source files in place
	$(STYLUA) $(LUA_SRC)

# ── CLI (Go) ─────────────────────────────────────────────────────────────────

validate-cli: validate-cli-lint validate-cli-format validate-cli-tidy validate-cli-test validate-cli-coverage ## Run all CLI validation checks

validate-cli-lint: ## Static analysis with go vet
	$(GO) vet $(GO_SRC)

validate-cli-format: ## Formatting check — fails if any file is not gofmt-clean
	@out=$$(gofmt -l $(CLI_DIRS)); \
	if [ -n "$$out" ]; then \
		echo "unformatted files (run 'make fix-cli-format'):"; \
		echo "$$out"; \
		exit 1; \
	fi

validate-cli-tidy: ## Check that go.mod and go.sum are tidy
	@cp go.mod go.mod.tidy-bak && cp go.sum go.sum.tidy-bak; \
	$(GO) mod tidy; \
	if ! diff -q go.mod go.mod.tidy-bak > /dev/null 2>&1 || \
	   ! diff -q go.sum go.sum.tidy-bak > /dev/null 2>&1; then \
	    cp go.mod.tidy-bak go.mod; cp go.sum.tidy-bak go.sum; \
	    rm -f go.mod.tidy-bak go.sum.tidy-bak; \
	    echo "go.mod/go.sum not tidy — run 'make fix-cli-tidy'"; exit 1; \
	fi; \
	rm -f go.mod.tidy-bak go.sum.tidy-bak

validate-cli-test: ## Run Go test suite without coverage instrumentation (fast, local)
	$(GO) test $(GO_SRC)

validate-cli-coverage: ## Run Go tests with coverage and enforce $(CLI_COV_MIN)% threshold
	$(GO) test -coverprofile=coverage-cli.out $(GO_SRC)
	@go tool cover -func=coverage-cli.out | awk '/^total:/{gsub(/%/, "", $$NF); cov = $$NF + 0} \
	  END { \
	    printf "CLI Coverage: %.2f%%\n", cov; \
	    if (cov < $(CLI_COV_MIN)) { \
	      printf "FAIL: %.2f%% is below the %d%% threshold\n", cov, $(CLI_COV_MIN); exit 1 \
	    } else { \
	      printf "PASS: meets %d%% threshold\n", $(CLI_COV_MIN) \
	    } \
	  }'

fix-cli-tidy: ## Run go mod tidy in place
	$(GO) mod tidy

fix-cli-format: ## Auto-format all Go source files in place
	gofmt -w $(CLI_DIRS)

# ── Build ─────────────────────────────────────────────────────────────────────

build-cli: ## Build the CLI binary for the current platform (output: ./bin/macromog)
	$(GO) build -o $(BINARY) $(CLI_MAIN)

build-cli-all: ## Cross-compile the CLI for all release platforms (compilation check)
	@for target in $(BUILD_PLATFORMS); do \
	    os=$$(printf '%s' "$$target" | cut -d/ -f1); \
	    arch=$$(printf '%s' "$$target" | cut -d/ -f2); \
	    printf "  %-24s" "$$os/$$arch"; \
	    GOOS=$$os GOARCH=$$arch $(GO) build -o /dev/null ./cmd/... && printf "OK\n" || exit 1; \
	done
	@printf "All platforms: OK\n"

# ── Umbrella fix targets ──────────────────────────────────────────────────────

fix: fix-plugin-format fix-cli-format fix-cli-tidy ## Auto-fix all issues that can be fixed automatically

# ── Housekeeping ─────────────────────────────────────────────────────────────

clean: ## Remove generated coverage artifacts and local build output
	rm -f luacov.stats.out luacov.report.out coverage-plugin.xml coverage-cli.out $(BINARY)
