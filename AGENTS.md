# AGENTS.md - Documentation & Voice Guide

## Project Voice: Kupomog
Kupomog is a cheerful, knowledgeable Moogle archivist who helps adventurers organize their macro books.

**Tone Guidelines:**
- Friendly and encouraging, never condescending
- Use light Moogle flavor ("kupo!", "adventurer", etc.) sparingly for charm
- Clear, concise technical writing
- Assume users are experienced FFXI players but new to the tool
- Positive and community-oriented
- Consistent across all documentation

## Usage
- README.md: Welcoming overview
- SPEC.md: Precise technical reference
- INSTRUCTIONS.md (future): Step-by-step with examples
- In-addon messages: Match this tone

Approved by Kupomog himself, kupo!

Kupomog especially loves clear documentation around book names, macro titles (8 chars max), and the 6-line limit, kupo!

---

## Commit Messages: Conventional Commits

All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) spec.
This drives automated versioning via `release-please` — wrong format = wrong version bump.

### Format

```
<type>(<scope>): <short summary>

[optional body]

[optional footer(s)]
```

### Types

| Type | Semver bump | When to use |
|------|-------------|-------------|
| `feat` | minor | New user-facing feature |
| `fix` | patch | Bug fix |
| `perf` | patch | Performance improvement |
| `refactor` | — | Code restructure, no behavior change |
| `chore` | — | Tooling, deps, config, CI |
| `docs` | — | Documentation only |
| `test` | — | Tests only |
| `style` | — | Formatting, whitespace |

Append `!` after the type for a **breaking change** (major bump): `feat!:`, `fix!:`

### Scopes (optional but encouraged)

`core`, `yaml`, `macros`, `validate`, `ci`, `docs`

### Examples

```
feat(yaml): implement pure-Lua YAML parser for macro structure
fix(validate): correct book index upper bound check
chore(ci): add release-please workflow
feat!: change YAML schema — book indices now 1-based
docs: add INSTRUCTIONS.md with quick-start guide
```

### Rules
- Summary line: imperative mood, lowercase, no trailing period, ≤72 chars
- Body: explain *why*, not *what* (the diff shows what)
- Breaking changes: add `BREAKING CHANGE: <description>` in the footer (in addition to `!`)

---

## Starting Fresh Work

Before writing a single line of code:

1. **Clean working tree** — `git status` must show no uncommitted changes. If
   files are dirty, stop and resolve that first.
2. **Passing baseline** — execute `make validate` in a shell tool and verify
   from the **actual output** that all checks pass. Do not simulate this step,
   assume it passes, or describe what it would do — run it and show the result.
   If anything is already failing, fix it before touching anything else. This
   ensures you can attribute any new failure to your own changes.

## Before Pushing / Calling Work Complete

Execute the full suite in a shell tool before declaring a task done or pushing
to a branch:

```sh
make validate
```

This runs `validate-lint`, `validate-format`, and `validate-coverage` in
sequence — exactly what CI enforces. If it passes locally, CI will pass.

**You must actually run this command.** Do not simulate the result, describe
what it would do, or carry forward an earlier passing result. Run it fresh and
share the real output. A task is not complete until this command has been
executed and its output confirms success.

## Commit Workflow

When working in this repo, agents must follow this workflow for every commit:

1. **Single-purpose commits** — each commit must contain exactly one logical change. Do not bundle unrelated edits.
2. **Stop before committing** — after completing a logical unit of work, pause and do not commit yet.
3. **Present for approval** — provide:
   - A brief summary of what changed and why
   - The proposed commit message (following Conventional Commits above)
4. **Wait for explicit approval** — only commit after the user approves the message. Do not proceed to the next task until the commit is made and acknowledged.
