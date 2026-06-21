# Macromog

**Kupo!** Your trusty Moogle macro librarian for Final Fantasy XI and Windower 4.

Macromog allows you to manage your entire collection of FFXI macros using simple, readable YAML files. Export your current macros, make changes in a text editor, validate them, and import them back with automatic backups.

Perfect for players with 40 macro books full of job-specific setups, gear swaps, and complex commands.

## Status

Early Development. Core YAML handling & validation in progress. Full memory/DAT integration coming soon, kupo!

## Features

- Export your current in-game macros (including custom book names) to a structured `<character>_macros.yml`
- Import from YAML with full validation and automatic backup
- Sparse format — only defined books, sets, and macros are stored
- Strict validation against FFXI limits (40 books, 10 sets/book, 20 macros/set, 8-char macro titles, 6 lines max, etc.)
- Supports all FFXI client locales

**Brought to you by Kupomog**, the helpful Moogle archivist!

## Quick Start

Full setup and command walkthroughs will live in `docs/` once implemented, kupo!

## Documentation

- [Behavior specification](docs/SPEC.md) — YAML schema, CLI/plugin commands, constraints
- [DAT file format](docs/DAT-FORMAT.md) — FFXI macro `.dat` and `.ttl` binary layout
- [Contributing](docs/CONTRIBUTING.md) — setup, validation, PR workflow

## Repository Structure

- `macromog.lua` — Main addon
- `docs/` — Technical documentation
- `AGENTS.md` — Development and agent guide (repo root for tooling)
- `data/` — YAML macro files

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for setup steps, commit message guidelines, and PR workflow.

---

*Macromog is a community-driven Windower 4 addon. Contributions welcome!*
