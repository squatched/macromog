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

(See `INSTRUCTIONS.md` for full setup and commands once implemented, kupo!)

## Repository Structure

- `macromog.lua` - Main addon
- `SPEC.md` - Detailed behavior specification
- `AGENTS.md` - Development guide
- `data/` - YAML files live here

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup steps, commit message guidelines, and PR workflow.

---

*Macromog is a community-driven Windower 4 addon. Contributions welcome!*
