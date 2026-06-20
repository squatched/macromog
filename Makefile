LUACHECK         := luacheck
BUSTED           := busted
STYLUA           := stylua
LUACOV           := luacov
LUACOV_COBERTURA := luacov-cobertura

LUA_SRC      := macromog.lua lib/
TEST_DIR     := tests/
COVERAGE_MIN := 80

.PHONY: all validate validate-lint validate-format validate-test validate-coverage \
        fix fix-format clean

all: validate

## Run all validation checks (lint + format + coverage)
validate: validate-lint validate-format validate-coverage

## Static analysis with luacheck
validate-lint:
	$(LUACHECK) $(LUA_SRC)

## Formatting check — fails if any file is not formatted
validate-format:
	$(STYLUA) --check $(LUA_SRC)

## Run test suite without coverage instrumentation (fast, local iteration)
validate-test:
	$(BUSTED) $(TEST_DIR)

## Run tests with coverage and enforce the $(COVERAGE_MIN)% threshold
validate-coverage:
	$(BUSTED) --coverage $(TEST_DIR)
	$(LUACOV)
	$(LUACOV_COBERTURA) -o coverage.xml
	@awk '/^Total/{gsub(/%/, "", $$NF); cov = $$NF + 0} \
	  END { \
	    printf "Coverage: %.2f%%\n", cov; \
	    if (cov < $(COVERAGE_MIN)) { \
	      printf "FAIL: %.2f%% is below the %d%% threshold\n", cov, $(COVERAGE_MIN); exit 1 \
	    } else { \
	      printf "PASS: meets %d%% threshold\n", $(COVERAGE_MIN) \
	    } \
	  }' luacov.report.out

## Auto-format all Lua source files in place
fix-format:
	$(STYLUA) $(LUA_SRC)

## Remove generated coverage artifacts
clean:
	rm -f luacov.stats.out luacov.report.out coverage.xml
