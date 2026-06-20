# Macromog

**Kupo!** Your trusty Moogle macro librarian for Final Fantasy XI and Windower 4.

Macromog allows you to manage your entire collection of FFXI macros using simple, readable YAML files. Export your current macros, make changes in a text editor, validate them, and import them back with automatic backups.

Perfect for players with 40 macro books full of job-specific setups, gear swaps, and complex commands.

## Features (Planned / In Progress)

- Export current in-game macros to a per-character YAML file
- Import from YAML with full validation and automatic backup
- Sparse storage — only defined books, sets, and macros are saved
- Strict validation against FFXI limits (40 books, 10 sets/book, 20 macros/set, 8-char names, 6 lines, etc.)
- Support for all FFXI client locales
- Easy to read/edit YAML structure

**Brought to you by Kupomog**, the helpful Moogle archivist!

## Quick Start

(See `INSTRUCTIONS.md` for full setup and commands once implemented, kupo!)

## Repository Structure

- `macromog.lua` - Main addon
- `SPEC.md` - Detailed behavior specification
- `AGENTS.md` - Documentation tone guide
- `data/` - Your YAML files live here

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for setup steps, commit message guidelines, and PR workflow.

---

*Macromog is a community-driven Windower 4 addon. Contributions welcome!*
