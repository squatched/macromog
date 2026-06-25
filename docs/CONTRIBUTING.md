# Contributing to Macromog

Thanks for helping out, adventurer! Kupomog appreciates every contribution, kupo!

## First-time setup

```sh
git clone git@github.com:squatched/macromog.git
cd macromog

# Install the commit-msg hook (enforces commit message rules locally)
ln -sf ../../.githooks/commit-msg .git/hooks/commit-msg
```

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

This runs `validate-plugin` and `validate-cli` in sequence. CI enforces the
same targets — if `make validate` passes locally, CI will pass.

Optional checks (see below) are not required for PRs.

## Validation tiers

| Tier | Target | Required? |
|------|--------|-----------|
| Gate | `make validate` | Yes — every PR |
| Fast iteration | `make validate-plugin-test`, `make validate-cli-test` | No — while developing |
| Optional | `make validate-wine-smoke` | No — Linux + Wine only |

### Required targets (`make validate`)

| Target | What it does |
|--------|-------------|
| `make validate` | Runs all checks below (trailing WS, plugin, CLI) |
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

### Optional targets

| Target | What it does |
|--------|-------------|
| `make validate-wine-smoke` | Runs the cross-compiled Windows CLI under Wine; prints `SKIP` if Wine is not installed |

`validate-wine-smoke` checks that the bundled `macromog.exe` launches under
Wine (`--help`). Shared-config read/write under Wine is covered by Go integration
tests on the Linux host; **Windows contributors do not need Wine** — run
`dist\bin\macromog.exe --help` natively instead.

## Fix targets

| Target | What it fixes |
|--------|--------------|
| `make fix` | Auto-fix plugin and CLI formatting |
| `make fix-plugin-format` | StyLua on `macromog.lua` and `lib/` |
| `make fix-cli-format` | `gofmt` on `cmd/` |
| `make fix-cli-tidy` | `go mod tidy` |

Lint errors and coverage gaps still require manual fixes.

## Validation tools

### Required for `make validate`

**Go** — see `go.mod` for the minimum version. Install from
[golang.org](https://go.dev/dl/) or your package manager.

**Lua tooling** via luarocks:

```sh
# luarocks (Arch)
sudo pacman -S luarocks

# luarocks (Debian/Ubuntu)
sudo apt install luarocks

# luarocks (macOS)
brew install luarocks
```

```sh
luarocks install luacheck
luarocks install busted
luarocks install luacov
luarocks install luacov-cobertura
```

**StyLua** — standalone binary or package manager:

```sh
# Arch
sudo pacman -S stylua
```

Other platforms: [github.com/JohnnyMorganz/StyLua/releases](https://github.com/JohnnyMorganz/StyLua/releases)

### Optional for `make validate-wine-smoke`

| Tool | Who needs it | Install |
|------|----------------|---------|
| Wine | Linux contributors testing the Windows `.exe` under Wine | Arch: `sudo pacman -S wine` · Debian/Ubuntu: `sudo apt install wine` |

Not required on native Windows — use the `.exe` directly.

The macromog CLI is a Go binary, not .NET. If Wine prompts to install a .NET
runtime when you run the `.exe` manually, cancel it — the CLI does not need it.
`make validate-wine-smoke` disables that stub automatically.

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

Coverage below 80% (plugin or CLI) fails CI. PR comments include a coverage
summary when available.

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
