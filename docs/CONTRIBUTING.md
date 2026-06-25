# Contributing to Macromog

Thanks for helping out, adventurer! Kupomog appreciates every contribution, kupo!

## Supported environments

| Environment | Status | Notes |
|-------------|--------|-------|
| Linux (Arch) | ✅ Fully tested — primary dev environment | |
| Linux (Debian/Ubuntu) | ✅ Tested in CI | |
| Windows via WSL2 | ✅ Recommended Windows path | Follow whichever Linux distro instructions match your WSL2 install |
| Windows native (MSYS2) | ⚠️ Untested | For advanced users; PRs with corrections welcome |

macOS is not a supported platform — FFXI has no native macOS client.

### Windows contributors: use WSL2

Install WSL2 with your preferred Linux distro (Ubuntu is the default):

```
wsl --install
```

Then follow the **Debian/Ubuntu** instructions throughout this document. The
smoke test (`validate-spawn-smoke`) detects WSL2 automatically via
`$WSL_DISTRO_NAME` and runs `macromog.exe` through WSL2's Windows interop —
which executes the binary as a real native Windows process — so no Wine is
needed.

If you know what you're doing and prefer a fully native Windows shell, MSYS2
works but is untested. You'll need to adapt the Linux tool install commands to
MSYS2's `pacman`.

## First-time setup

```sh
git clone git@github.com:squatched/macromog.git
cd macromog

# Install the commit-msg hook (enforces commit message format locally)
ln -sf ../../.githooks/commit-msg .git/hooks/commit-msg
```

`ln -sf` works on Linux and WSL2. On a native Windows shell (MSYS2 / Git Bash),
enable **Developer Mode** first (*Settings → System → Developer Mode*).

Then install the validation tools (see [Validation tools](#validation-tools) below).

## Before you start

Before beginning any work, make sure you're starting from a clean baseline:

```sh
git status          # working tree must be clean
make validate       # all checks must pass before you touch anything
```

If `make validate` fails on a clean checkout, fix that first before making
your own changes — otherwise you can't tell whether a failure is yours or
pre-existing.

## Before pushing / calling work complete

Run the full suite before pushing or declaring a PR ready:

```sh
make validate
```

This runs all checks in sequence — if `make validate` passes locally, CI will pass.

## Validation tiers

| Tier | Target | Required? |
|------|--------|-----------|
| Gate | `make validate` | Yes — every PR |
| Fast iteration | `make validate-plugin-test`, `make validate-cli-test`, `make validate-spawn` | No — while developing |

### Required targets (`make validate`)

| Target | What it does |
|--------|-------------|
| `make validate` | Runs all checks below (trailing WS, plugin, CLI, spawn DLL) |
| `make validate-trailing-ws` | Fails on trailing whitespace or bad EOF newlines in tracked and untracked text files (respecting `.gitignore`) |
| `make fix-trailing-ws` | Auto-fixes trailing whitespace and EOF newlines |
| `make validate-plugin` | Lint, format, coverage, and package layout for the Lua addon |
| `make validate-plugin-lint` | Static analysis via luacheck |
| `make validate-plugin-format` | Format check via StyLua |
| `make validate-plugin-test` | Busted tests without coverage (fast) |
| `make validate-plugin-coverage` | Busted + luacov; fails below 80% plugin coverage |
| `make validate-plugin-package` | Builds release zip and verifies `Macromog/` layout |
| `make validate-cli` | Lint, format, tidy, test, and coverage for the Go CLI |
| `make validate-cli-lint` | `go vet` |
| `make validate-cli-format` | `gofmt` check |
| `make validate-cli-tidy` | `go mod tidy` check |
| `make validate-cli-test` | `go test` without coverage (fast) |
| `make validate-cli-coverage` | `go test -coverprofile`; fails below 80% CLI coverage |
| `make validate-spawn` | Lint, format, unit tests, and coverage for the C DLL |
| `make validate-spawn-lint` | `cppcheck` static analysis |
| `make validate-spawn-format` | `clang-format` check |
| `make validate-spawn-test` | Compile and run native C unit tests for helper functions |
| `make validate-spawn-coverage` | `gcov` on `helpers.h`; fails below 95% spawn coverage |
| `make validate-spawn-smoke` | Run `macromog.exe --help` as a native Windows process; self-skips on Linux without Wine |

`validate-spawn-smoke` is part of `make validate` but behaves differently per
environment:

- **WSL2**: runs `macromog.exe` via Windows interop as a real Windows process — no Wine needed
- **Native Windows** (MSYS2 / Git Bash): runs `macromog.exe` directly
- **Linux with Wine**: runs `macromog.exe` under Wine
- **Linux without Wine / CI**: prints `SKIP` and exits 0

Shared-config read/write is covered by Go integration tests on the Linux host.

## Fix targets

| Target | What it fixes |
|--------|--------------|
| `make fix` | Auto-fix plugin, CLI, and spawn DLL formatting |
| `make fix-plugin-format` | StyLua on `macromog.lua` and `lib/` |
| `make fix-cli-format` | `gofmt` on `cmd/` |
| `make fix-cli-tidy` | `go mod tidy` |
| `make fix-spawn-format` | `clang-format` on `spawn/*.c` and `spawn/*.h` |

Lint errors and coverage gaps still require manual fixes.

## Validation tools

Install instructions are for **Arch** and **Debian/Ubuntu**. WSL2 users: pick
whichever column matches your installed distro.

### Go

See `go.mod` for the minimum version.

```sh
# Arch
sudo pacman -S go

# Debian/Ubuntu
sudo apt install golang-go
```

Or download the latest installer from [go.dev/dl](https://go.dev/dl/).

### mingw-w64 (spawn DLL)

Required to build `macromog_spawn.dll`. This is a Windows cross-compiler — it
builds PE binaries from a Linux host (including WSL2).

```sh
# Arch
sudo pacman -S mingw-w64-gcc

# Debian/Ubuntu
sudo apt install gcc-mingw-w64-i686
```

### cppcheck

```sh
# Arch
sudo pacman -S cppcheck

# Debian/Ubuntu
sudo apt install cppcheck
```

### clang-format

```sh
# Arch
sudo pacman -S clang

# Debian/Ubuntu
sudo apt install clang-format
```

### Lua tooling

```sh
# luarocks (Arch)
sudo pacman -S luarocks

# luarocks (Debian/Ubuntu)
sudo apt install luarocks
```

Then install the Lua packages:

```sh
luarocks install luacheck
luarocks install busted
luarocks install luacov
luarocks install luacov-cobertura
```

### StyLua

```sh
# Arch
sudo pacman -S stylua
```

Other platforms (including Debian/Ubuntu): download a pre-built binary from
[github.com/JohnnyMorganz/StyLua/releases](https://github.com/JohnnyMorganz/StyLua/releases)

### Wine (optional — Linux only)

`validate-spawn-smoke` self-skips on Linux when Wine is not installed. WSL2
contributors don't need Wine — interop handles it. If you're on bare Linux and
want the smoke check to run:

```sh
# Arch
sudo pacman -S wine

# Debian/Ubuntu
sudo apt install wine
```

The macromog CLI is a Go binary, not .NET. If Wine prompts to install a .NET
runtime when you run the `.exe` manually, cancel it — the CLI does not need it.
`make validate-spawn-smoke` disables that prompt automatically.

## Releases

Releases are fully automated — never push tags or edit `CHANGELOG.md` by hand.

When releasable commits land on `main`, the Release Please bot opens (or updates) a
**"Release vX.Y.Z" PR** that pre-writes the changelog entry and bumps `version.txt`
(and `_addon.version` in `macromog.lua`). A maintainer merges it; that merge creates
the tag and publishes the GitHub Release with `dist/macromog-<version>.zip` plus the
`macromog`, `macromog.exe`, and the plugin zip attached automatically.

**The semver bump comes from your commit type:**

| Commit type | Version bump |
|---|---|
| `fix:`, `perf:` | patch — 0.0.**x** |
| `feat:` | minor — 0.**x**.0 |
| `feat!:` or `BREAKING CHANGE:` footer | major — **x**.0.0 |
| `chore:`, `docs:`, `style:`, `refactor:`, `test:` | no release |

Your only job is accurate commit messages — the rest is handled for you.

## CI

PRs must pass the workflows under `.github/workflows/`. All workflow steps call
`make` targets directly — local `make validate` is the source of truth.

Coverage below threshold fails CI: 80% for plugin and CLI, 95% for spawn DLL helpers.
PR comments include a coverage summary for all three components when available.

## Commit messages

All commits must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[(<scope>)][!]: <description>

[optional body — wrap at 72 chars]

[optional footer(s)]
```

The commit-msg hook will reject messages that don't comply. Full guidelines —
types, scopes, examples, and the 50/72 line-length rule — are in
[AGENTS.md](../AGENTS.md#commit-messages-conventional-commits).

## Pull requests

- Target the `main` branch.
- PRs are squash-merged; the **PR title becomes the commit message**, so it
  must also follow Conventional Commits (enforced automatically by CI).
- One logical change per PR where practical.
