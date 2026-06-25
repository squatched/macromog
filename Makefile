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
DIST_DIR         := dist
PLUGIN_STAGE     := $(DIST_DIR)/Macromog
ifndef VERSION
VERSION := $(shell tr -d '\n\r' < version.txt 2>/dev/null)
ifeq ($(strip $(VERSION)),)
VERSION := $(shell awk -F"'" '/_addon.version/{print $$2; exit}' macromog.lua)
endif
endif
PLUGIN_ZIP       := $(DIST_DIR)/macromog-$(VERSION).zip
BUILD_PLATFORMS  := linux/amd64 windows/386
RELEASE_BINS     := macromog macromog.exe
SPAWN_CC         := i686-w64-mingw32-gcc
SPAWN_SRC        := spawn/macromog_spawn.c
SPAWN_DEF        := spawn/LuaCore.def
SPAWN_BUILD      := spawn/build
SPAWN_DLL        := $(DIST_DIR)/bin/macromog_spawn.dll
SPAWN_C_SRCS     := spawn/macromog_spawn.c spawn/test_spawn.c spawn/helpers.h
SPAWN_TEST_BIN   := $(SPAWN_BUILD)/test_spawn
SPAWN_COV_BIN    := $(SPAWN_BUILD)/test_spawn_cov
SPAWN_GCNO       := $(SPAWN_COV_BIN)-test_spawn.gcno
SPAWN_COV_MIN    := 95
CPPCHECK         := cppcheck

.DEFAULT_GOAL := help

.PHONY: help \
        validate validate-trailing-ws validate-plugin validate-cli \
        validate-plugin-lint validate-plugin-format \
        fix-trailing-ws \
        validate-plugin-test validate-plugin-coverage validate-plugin-package validate-spawn-smoke \
        validate-cli-lint validate-cli-format validate-cli-tidy validate-cli-test validate-cli-coverage \
        fix fix-plugin-format fix-cli-format fix-cli-tidy fix-spawn-format \
        build-cli build-cli-all build-plugin build-release-bins build-spawn-dll package-plugin \
        validate-spawn validate-spawn-lint validate-spawn-format validate-spawn-test \
        validate-spawn-coverage \
        clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Targets:\n"} \
	  /^[a-zA-Z_-]+:.*?##/ {printf "  %-28s %s\n", $$1, $$2}' \
	  $(MAKEFILE_LIST)

validate: validate-trailing-ws validate-plugin validate-cli validate-spawn validate-spawn-smoke ## Run all validation checks across plugin (Lua), CLI (Go), and spawn DLL (C)

# ── Repository-wide ───────────────────────────────────────────────────────────

validate-trailing-ws: ## Fail on trailing whitespace or missing EOF newlines
	@chmod +x scripts/clean-trailing-ws.sh
	scripts/clean-trailing-ws.sh --check --verbose

fix-trailing-ws: ## Strip trailing whitespace and normalize EOF newlines in place
	@chmod +x scripts/clean-trailing-ws.sh
	scripts/clean-trailing-ws.sh

# ── Plugin (Lua / Windower addon) ────────────────────────────────────────────

validate-plugin: validate-plugin-lint validate-plugin-format validate-plugin-coverage validate-plugin-package ## Run all plugin validation checks

validate-plugin-lint: ## Static analysis with luacheck
	$(LUACHECK) $(LUA_SRC)

validate-plugin-format: ## Formatting check — fails if any file is not stylua-clean
	$(STYLUA) --check $(LUA_SRC)

validate-plugin-test: ## Run test suite without coverage instrumentation (fast, local)
	$(BUSTED) $(TEST_DIR)

validate-plugin-package: ## Verify release zip layout (empty data/, bundled Windows CLIs)
	@chmod +x scripts/validate-package.sh
	scripts/validate-package.sh

validate-spawn-smoke: build-release-bins ## Optional: run Windows CLI natively (Windows) or under Wine (Linux)
	@chmod +x scripts/validate-spawn-smoke.sh
	scripts/validate-spawn-smoke.sh

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

# ── Spawn DLL (C) ─────────────────────────────────────────────────────────────

validate-spawn: validate-spawn-lint validate-spawn-format validate-spawn-test validate-spawn-coverage build-spawn-dll ## Run all spawn DLL validation checks

validate-spawn-lint: ## Static analysis of C source with cppcheck
	@which $(CPPCHECK) > /dev/null 2>&1 || \
	    { echo "cppcheck not found — sudo pacman -S cppcheck  (Ubuntu: sudo apt install cppcheck)"; exit 1; }
	$(CPPCHECK) --enable=warning,performance,portability \
	    --error-exitcode=1 \
	    --suppress=missingIncludeSystem \
	    --suppress=missingInclude \
	    --platform=win32A \
	    $(SPAWN_SRC)
	$(CPPCHECK) --enable=warning,performance,portability \
	    --error-exitcode=1 \
	    spawn/test_spawn.c

validate-spawn-format: ## Format check for C source and headers
	clang-format --dry-run --Werror $(SPAWN_C_SRCS)

validate-spawn-test: ## Build and run native unit tests for C helper functions
	@mkdir -p $(SPAWN_BUILD)
	gcc -Wall -Wextra -Werror -std=c11 -I spawn/ \
	    -o $(SPAWN_TEST_BIN) spawn/test_spawn.c
	$(SPAWN_TEST_BIN)

validate-spawn-coverage: ## Run C helper tests with gcov; enforce $(SPAWN_COV_MIN)% threshold on helpers.h
	@mkdir -p $(SPAWN_BUILD)
	gcc -Wall -Wextra -std=c11 -I spawn/ -fprofile-arcs -ftest-coverage \
	    -o $(SPAWN_COV_BIN) spawn/test_spawn.c
	$(SPAWN_COV_BIN)
	gcov $(SPAWN_GCNO) > $(SPAWN_BUILD)/spawn-gcov-raw.txt
	@grep -A1 "helpers\.h" $(SPAWN_BUILD)/spawn-gcov-raw.txt | \
	    grep "Lines executed:" | \
	    awk -F"[:%]" '{pct=$$2+0; \
	        printf "Spawn Coverage: %.2f%%\n", pct; \
	        printf "Spawn Coverage: %.2f%%\n", pct > "$(SPAWN_BUILD)/spawn-coverage.txt"; \
	        if (pct < $(SPAWN_COV_MIN)) { \
	            printf "FAIL: %.2f%% is below the %d%% threshold\n", pct, $(SPAWN_COV_MIN); exit 1 \
	        } \
	        printf "PASS: meets %d%% threshold\n", $(SPAWN_COV_MIN)}'
	@rm -f test_spawn.c.gcov helpers.h.gcov

fix-spawn-format: ## Auto-format all C source files and headers in place
	clang-format -i $(SPAWN_C_SRCS)

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

build-release-bins: ## Cross-compile release CLI binaries into dist/bin/
	@mkdir -p $(DIST_DIR)/bin
	GOOS=linux GOARCH=amd64 $(GO) build -o $(DIST_DIR)/bin/macromog $(CLI_MAIN)
	GOOS=windows GOARCH=386 $(GO) build -ldflags="-H windowsgui -s -w" -o $(DIST_DIR)/bin/macromog.exe $(CLI_MAIN)

build-spawn-dll: ## Build macromog_spawn.dll (32-bit, hidden process spawn + file mtime)
	@mkdir -p $(SPAWN_BUILD) $(DIST_DIR)/bin
	i686-w64-mingw32-dlltool -d $(SPAWN_DEF) -D LuaCore.dll -l $(SPAWN_BUILD)/libLuaCore.a
	$(SPAWN_CC) -shared -o $(SPAWN_DLL) $(SPAWN_SRC) \
	    -L $(SPAWN_BUILD) -lLuaCore \
	    -lkernel32 \
	    -Wl,--kill-at \
	    -Wall -O2 -s

build-plugin: build-release-bins build-spawn-dll ## Stage the Windower addon tree under dist/Macromog/
	@rm -rf $(PLUGIN_STAGE)
	@mkdir -p $(PLUGIN_STAGE)/lib $(PLUGIN_STAGE)/data $(PLUGIN_STAGE)/bin
	cp macromog.lua $(PLUGIN_STAGE)/
	cp -r lib/. $(PLUGIN_STAGE)/lib/
	cp $(DIST_DIR)/bin/macromog.exe $(PLUGIN_STAGE)/bin/
	cp $(SPAWN_DLL) $(PLUGIN_STAGE)/bin/

package-plugin: build-plugin ## Create dist/macromog-<version>.zip from the staged addon
	@mkdir -p $(DIST_DIR)
	rm -f $(PLUGIN_ZIP)
	cd $(DIST_DIR) && zip -r "macromog-$(VERSION).zip" Macromog

# ── Umbrella fix targets ──────────────────────────────────────────────────────

fix: fix-trailing-ws fix-plugin-format fix-cli-format fix-cli-tidy fix-spawn-format ## Auto-fix all issues that can be fixed automatically

# ── Housekeeping ─────────────────────────────────────────────────────────────

clean: ## Remove generated coverage artifacts and local build output
	rm -f luacov.stats.out luacov.report.out coverage-plugin.xml coverage-cli.out $(BINARY)
	rm -rf $(DIST_DIR) $(SPAWN_BUILD)
