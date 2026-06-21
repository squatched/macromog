LUACHECK         := luacheck
BUSTED           := busted
STYLUA           := stylua
LUACOV           := luacov
LUACOV_COBERTURA := luacov-cobertura
GO               := go

LUA_SRC          := macromog.lua lib/
TEST_DIR         := tests/
PLUGIN_COV_MIN   := 80
# CLI_COV_MIN is 0 for the stub; raise it as implementation fills in
CLI_COV_MIN      := 0
GO_SRC           := ./...
CLI_DIRS         := ./cmd
CLI_MAIN         := ./cmd/macromog
BINARY           := macromog
BUILD_PLATFORMS  := \
    darwin/amd64 darwin/arm64 \
    linux/amd64  linux/386   \
    windows/amd64 windows/386

.PHONY: all \
        validate validate-plugin validate-cli \
        validate-plugin-lint validate-plugin-format \
        validate-plugin-test validate-plugin-coverage \
        validate-cli-lint validate-cli-format validate-cli-test validate-cli-coverage \
        fix fix-plugin-format fix-cli-format \
        build-cli build-cli-all \
        clean

all: validate

## Run all validation checks across the plugin (Lua) and CLI (Go)
validate: validate-plugin validate-cli

# ── Plugin (Lua / Windower addon) ────────────────────────────────────────────

## Run all plugin validation checks (lint + format + coverage)
validate-plugin: validate-plugin-lint validate-plugin-format validate-plugin-coverage

## Static analysis with luacheck
validate-plugin-lint:
	$(LUACHECK) $(LUA_SRC)

## Formatting check — fails if any file is not formatted
validate-plugin-format:
	$(STYLUA) --check $(LUA_SRC)

## Run test suite without coverage instrumentation (fast, local iteration)
validate-plugin-test:
	$(BUSTED) $(TEST_DIR)

## Run tests with coverage and enforce the $(PLUGIN_COV_MIN)% threshold
validate-plugin-coverage:
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

## Auto-format all Lua source files in place
fix-plugin-format:
	$(STYLUA) $(LUA_SRC)

# ── CLI (Go) ─────────────────────────────────────────────────────────────────

## Run all CLI validation checks (lint + format + test + coverage)
validate-cli: validate-cli-lint validate-cli-format validate-cli-test validate-cli-coverage

## Static analysis with go vet
validate-cli-lint:
	$(GO) vet $(GO_SRC)

## Formatting check — fails if any file is not gofmt-clean
validate-cli-format:
	@out=$$(gofmt -l $(CLI_DIRS)); \
	if [ -n "$$out" ]; then \
		echo "unformatted files (run 'make fix-cli-format'):"; \
		echo "$$out"; \
		exit 1; \
	fi

## Run Go test suite without coverage instrumentation (fast, local iteration)
validate-cli-test:
	$(GO) test $(GO_SRC)

## Run Go tests with coverage and enforce the $(CLI_COV_MIN)% threshold
validate-cli-coverage:
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

## Auto-format all Go source files in place
fix-cli-format:
	gofmt -w $(CLI_DIRS)

# ── Build ─────────────────────────────────────────────────────────────────────

## Build the CLI binary for the current platform (output: ./macromog)
build-cli:
	$(GO) build -o $(BINARY) $(CLI_MAIN)

## Cross-compile the CLI for all release platforms (compilation check, no output)
build-cli-all:
	@for target in $(BUILD_PLATFORMS); do \
	    os=$$(printf '%s' "$$target" | cut -d/ -f1); \
	    arch=$$(printf '%s' "$$target" | cut -d/ -f2); \
	    printf "  %-24s" "$$os/$$arch"; \
	    GOOS=$$os GOARCH=$$arch $(GO) build ./cmd/... && printf "OK\n" || exit 1; \
	done
	@printf "All platforms: OK\n"

# ── Umbrella fix targets ──────────────────────────────────────────────────────

## Auto-fix all issues that can be fixed automatically
fix: fix-plugin-format fix-cli-format

# ── Housekeeping ─────────────────────────────────────────────────────────────

## Remove generated coverage artifacts and local build output
clean:
	rm -f luacov.stats.out luacov.report.out coverage-plugin.xml coverage-cli.out $(BINARY)
