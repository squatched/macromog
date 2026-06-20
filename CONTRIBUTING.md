# Contributing to Macromog

Thanks for helping out, adventurer! Kupomog appreciates every contribution, kupo!

## First-time setup

```sh
git clone git@github.com:squatched/macromog.git
cd macromog

# Install the commit-msg hook (enforces commit message rules locally)
ln -sf ../../.githooks/commit-msg .git/hooks/commit-msg
```

That's it. No build step, no dependencies to install.

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

## Linting

```sh
# Requires luarocks: sudo apt install luarocks / brew install luarocks
luarocks install luacheck
luacheck macromog.lua lib/
```

CI runs this on every push and PR.
