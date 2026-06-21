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

This runs `validate-lint`, `validate-format`, and `validate-coverage` in
sequence. CI enforces exactly these same targets — if `make validate` passes
locally, CI will pass.

## Validation targets

| Target | What it does |
|--------|-------------|
| `make validate` | Runs all checks below in sequence |
| `make validate-lint` | Static analysis via luacheck |
| `make validate-format` | Format check via StyLua (fails if files need formatting) |
| `make validate-test` | Run test suite without coverage overhead (fast, local iteration) |
| `make validate-coverage` | Run tests with luacov; fails if coverage drops below 80% |

## Fix targets

Fix targets are focused — apply the one you need, not all at once:

| Target | What it fixes |
|--------|--------------|
| `make fix-format` | Auto-formats all Lua source files with StyLua |

There is no blanket `fix` target. Lint errors require manual inspection;
coverage gaps require new tests.

## Validation tools

Install these once before running any `make validate-*` target:

```sh
# luarocks (Arch)
sudo pacman -S luarocks

# luarocks (Debian/Ubuntu)
sudo apt install luarocks

# luarocks (macOS)
brew install luarocks
```

Then install the Lua packages:

```sh
luarocks install luacheck
luarocks install busted
luarocks install luacov
luarocks install luacov-cobertura
```

StyLua is a standalone binary — grab the latest release for your platform
from [github.com/JohnnyMorganz/StyLua/releases](https://github.com/JohnnyMorganz/StyLua/releases)
and put it on your `$PATH`. Or check your package manager.

```sh
# StyLua (Arch)
sudo pacman -S stylua
```

## Releases

Releases are fully automated — never push tags or edit `CHANGELOG.md` by hand.

When releasable commits land on `main`, the Release Please bot opens (or updates) a
**"Release vX.Y.Z" PR** that pre-writes the changelog entry and bumps `VERSION`. A
maintainer merges it; that merge creates the tag and publishes the GitHub Release with
the addon zip attached automatically.

**The semver bump comes from your commit type:**

| Commit type | Version bump |
|---|---|
| `fix:`, `perf:` | patch — 0.0.**x** |
| `feat:` | minor — 0.**x**.0 |
| `feat!:` or `BREAKING CHANGE:` footer | major — **x**.0.0 |
| `chore:`, `docs:`, `style:`, `refactor:`, `test:` | no release |

Your only job is accurate commit messages — the rest is handled for you.

## CI / branch protection

Every PR must pass all three CI jobs before it can merge:

| Status check | Workflow | What it runs |
|---|---|---|
| `Lint / luacheck` | `lint.yml` | `make validate-lint` |
| `Lint / stylua` | `lint.yml` | `make validate-format` |
| `Test / coverage` | `test.yml` | `make validate-coverage` |

Coverage below 80% fails the `Test / coverage` job and blocks the merge.
A comment on the PR will show the per-file breakdown and link to the
uploaded `coverage.xml` artifact so you can see exactly which lines are
missing.

## Commit messages

All commits must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[(<scope>)][!]: <description>

[optional body — wrap at 72 chars]

[optional footer(s)]
```

The commit-msg hook will reject messages that don't comply. Full guidelines —
types, scopes, examples, and the 50/72 line-length rule — are in
[AGENTS.md](AGENTS.md#commit-messages-conventional-commits).

## Pull requests

- Target the `main` branch.
- PRs are squash-merged; the **PR title becomes the commit message**, so it
  must also follow Conventional Commits (enforced automatically by CI).
- One logical change per PR where practical.
