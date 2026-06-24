# macromog CLI — User Guide

Welcome, adventurer! This guide covers everything you need to manage your FFXI macro books with the `macromog` command-line tool, kupo!

---

## Installation

Download the binary for your platform from the [Releases](https://github.com/squatched/macromog/releases) page and place it somewhere on your `PATH`.

| Platform | Binary |
|----------|--------|
| Windows 64-bit | `macromog-windows-amd64.exe` |
| Windows 32-bit | `macromog-windows-386.exe` |
| Linux 64-bit | `macromog-linux-amd64` |
| Linux 32-bit | `macromog-linux-386` |

---

## Quick Start

The typical workflow is three steps:

```sh
# 1. Export your current macros to YAML
macromog export

# 2. Edit the resulting YAML file to your liking
#    (your editor here)

# 3. Import the edited YAML back into FFXI
macromog import squatched_macros_20260620_033000.yml
```

macromog auto-detects your FFXI install on Windows and Linux. If detection fails, supply `--ffxi-path` pointing at your FFXI root (the folder that contains `USER`).

---

## Commands

| Command | What it does |
|---------|-------------|
| [`alias`](#alias--set-a-character-nickname) | Assign a friendly name to a character folder |
| [`export`](#export--export-macros-to-yaml) | Read macro `.dat` files and write them to YAML |
| [`import`](#import--import-macros-from-yaml) | Write macros from a YAML file back to `.dat` files |
| [`template`](#template--generate-a-blank-yaml-template) | Generate a blank YAML template for a given scope |
| [`validate`](#validate--validate-a-yaml-file) | Check a YAML file for schema and FFXI constraint errors |
| [`backup`](#backup--back-up-macro-files) | Create a timestamped copy of all macro `.dat` files |
| [`list`](#list--list-characters-and-books) | Show detected characters and their populated books |

---

## Global Flags

These flags work with all commands:

| Flag | Description |
|------|-------------|
| `--output text\|json` | Output format. Default: `text`. Use `json` for scripting. |
| `--ffxi-path <path>` | Path to your FFXI install root. Auto-detected if omitted (see [Detection](#environment-and-detection)). |
| `--char-dir <path>` | Path to a character's folder inside USER (e.g. `/path/to/USER/a1b2c3d4`). Bypasses character selection entirely. |
| `--char-name <name>` | Character alias (set with `macromog alias`). Bypasses selection; USER folder is auto-detected or use `--ffxi-path`. |

---

## Scope Selectors

The `--scope` flag narrows which macros are touched during `export`, `import`, or `template` generation.

### Selector components

| Component | Meaning | Range |
|-----------|---------|-------|
| `B<n>` | Book n | 1–40 |
| `S<n>` | Set n (within current book) | 1–10 |
| `C<n>` | Ctrl key n (within current set) | 0–9 |
| `A<n>` | Alt key n (within current set) | 0–9 |

### Ranges and wildcards

```
B1-5        books 1 through 5
B1S2-4      book 1, sets 2 through 4
B*          all books (full scope)
B1S*        all sets in book 1 (book-level scope)
B1S3C*      all ctrl keys in book 1, set 3
```

### Comma siblings

A comma within a selector creates a sibling at the current level:

```
B1,3,5      books 1, 3, and 5
B1S2,4      book 1, sets 2 and 4
B1S3A1,3    book 1, set 3, alt keys 1 and 3
B1S3A1,C2   book 1, set 3 — alt key 1 and ctrl key 2
```

### Multiple `--scope` flags

Use multiple flags for disjoint selectors at the same level:

```sh
macromog export --scope B1S3A1 --scope B5S2C4
```

All flags must resolve to the same scope level. Mixing levels across flags is an error.

### Scope levels

| Deepest component | Scope level |
|-------------------|-------------|
| `C` or `A` | macro |
| `S` | set |
| `B` | book |
| `*` or `B*` | full |

### Effect on import

The scope embedded in a YAML file controls what import is allowed to clear:

| YAML scope level | What import may clear |
|------------------|-----------------------|
| `full` | All 40 books (absent books are deleted) |
| `book` | Only the specified books |
| `set` | Only the specified (book, set) pairs |
| `macro` | Nothing — only the named slots are updated |

If you pass `--scope` to `import` and it exceeds the YAML's declared scope, macromog prompts for confirmation before proceeding.

---

## Character Selection

macromog needs to know which character's macros to operate on. It resolves this in order:

1. **`--char-dir <path>`** — use this folder directly, no questions asked.
2. **`--char-name <name>`** — look up the alias in `USER/characters.yml` and use its folder.
3. **`--all`** — operate on every character found in the USER folder.
4. **Auto-detect** — scan the USER folder:
   - One character found: use it automatically (prints a notice).
   - Multiple characters found on an interactive terminal: show a selection prompt.
   - Multiple characters found on non-interactive input (scripts, CI): error — use `--char-dir` or `--all`.

### Interactive Selection Prompt

```
Multiple characters found. Select characters:
  [1] Squatched (a1b2c3d4) (12 books)
  [2] e5f6a7b8 (3 books)
Enter numbers (e.g. 1, 1,3, 1-2, all):
```

Accepted input forms:

| Input | Meaning |
|-------|---------|
| `1` | Character 1 only |
| `1,3` | Characters 1 and 3 |
| `1-3` | Characters 1 through 3 |
| `all` | All listed characters |

---

### `alias` — Set a character nickname

```sh
macromog alias <char-id> <name>
macromog alias --remove <char-id>
```

Assign a friendly name to a character folder. Once set, use `--char-name <name>` instead of `--char-dir` in all other commands.

**Arguments:**
- `<char-id>` — the hex folder ID (e.g. `a1b2c3d4`)
- `<name>` — the friendly name to assign

**Flags:**
- `--ffxi-path <path>` — FFXI install root (auto-detected if omitted)
- `--remove` — remove the alias instead of setting it

**Examples:**
```sh
macromog alias a1b2c3d4 Squatched
macromog alias --remove a1b2c3d4
macromog alias --ffxi-path "/mnt/games/FINAL FANTASY XI" a1b2c3d4 Squatched
```

Aliases are stored in `USER/characters.yml`.

---

### `export` — Export macros to YAML

```sh
macromog export [flags] [<char-dir>] [output]
```

Reads macro `.dat` files and writes a sparse YAML file containing only non-empty macros.

**Arguments:**
- `[<char-dir>]` — character folder inside USER (positional alternative to `--char-dir`)
- `[output]` — output file path (positional alternative to `--output`/`-o`)

**Flags:**
- `--ffxi-path`, `--char-dir`, `--char-name`, `--all` — see [Global Flags](#global-flags) and [Character Selection](#character-selection)
- `--output <file>` / `-o <file>` — output YAML path; requires exactly one character
- `--name <name>` — embed this character name in the YAML `character:` field; requires one character
- `--scope <selector>` — limit export to specific books/sets/macros (repeatable; see [Scope Selectors](#scope-selectors))

**Output filename** (when `--output` is omitted): `<character>_macros_<YYYYMMDD_HHMMSS>.yml`

**Examples:**
```sh
# Auto-detect character, write timestamped file to current directory
macromog export

# Explicit character directory, explicit output file
macromog export /path/to/USER/a1b2c3d4 macros.yml

# By alias
macromog export --char-name Squatched -o macros.yml

# Export all characters
macromog export --all

# Export only books 1 and 5
macromog export --scope B1,5 -o books1and5.yml
```

---

### `import` — Import macros from YAML

```sh
macromog import [flags] <file> [<char-dir>]
```

Reads a YAML file, validates it, and writes macros to `.dat` files. A timestamped backup is created automatically before any writes. Validation runs automatically — you do not need to run `validate` first, though doing so gives you a clear error report before committing to an import.

**Arguments:**
- `<file>` — the YAML file to import (required)
- `[<char-dir>]` — character folder inside USER (positional alternative to `--char-dir`)

**Flags:**
- `--ffxi-path`, `--char-dir`, `--char-name`, `--all` — see [Global Flags](#global-flags) and [Character Selection](#character-selection)
- `--no-backup` — skip the automatic backup (use with care)
- `--dry-run` — validate the YAML and show what would be written, without writing anything
- `--scope <selector>` — override the scope embedded in the YAML (repeatable; see [Scope Selectors](#scope-selectors))

**Examples:**
```sh
# Standard import (backup is automatic, validation is automatic)
macromog import squatched_macros_20260620_033000.yml

# Dry run: show what would happen without writing
macromog import --dry-run macros.yml

# Import to a specific character
macromog import --char-name Squatched macros.yml

# Import only book 1, even from a full-scope YAML
macromog import --scope B1 macros.yml

# Import into all characters (useful for shared macro sets)
macromog import --all macros.yml
```

**Scope override confirmation**: if `--scope` expands the scope beyond what the YAML declared, macromog asks for confirmation before proceeding. This prevents accidentally clearing books or sets that the YAML has no data for.

---

### `template` — Generate a blank YAML template

```sh
macromog template [flags] <output>
```

Generates a blank YAML file pre-filled with every macro slot for the given scope. All slots have empty names and six empty content lines — fill in what you need and delete the rest.

**Arguments:**
- `<output>` — output YAML file (required)

**Flags:**
- `--scope <selector>` — scope for the template (default: full scope, all 40 books)
- `--char-name <name>` — embed a character name in the template's `character:` field

**Examples:**
```sh
# Full template (all 40 books, 10 sets each, 20 macros per set)
macromog template full.yml

# Only book 1, set 3
macromog template b1s3.yml --scope B1S3

# Specific macro slots
macromog template slots.yml --scope B1S3A1,C2

# With character name
macromog template squatched.yml --char-name Squatched
```

Templates include `version:` and `scope:` but not `exported_at:` — they are not exports.

---

### `validate` — Validate a YAML file

```sh
macromog validate <file>
```

Checks a YAML file against the macromog schema and FFXI constraints. Exits 0 if valid, 1 if errors are found.

**Examples:**
```sh
macromog validate macros.yml
macromog validate --output json macros.yml   # machine-readable
```

Validation checks include:
- Schema version is 1
- Book indices 1–40, set indices 1–10, key indices 0–9
- Book names: max 15 alphanumeric characters
- Macro names: max 8 characters
- Macro content lines: max 60 characters each (lines with auto-translate markers are exempt)

---

### `backup` — Back up macro files

```sh
macromog backup [flags] [<char-dir>]
```

Copies all `mcr*.dat` and `*.ttl` files to a timestamped directory named `<char-id>_YYYYMMDD_HHMMSS`.

**Arguments:**
- `[<char-dir>]` — character folder inside USER (positional alternative to `--char-dir`)

**Flags:**
- `--ffxi-path`, `--char-dir`, `--char-name`, `--all` — see [Global Flags](#global-flags)
- `--out <path>` — directory to write the backup into (default: current directory)
- `--in-place` — write the backup into `<char-dir>/backups/`

`--out` and `--in-place` are mutually exclusive.

**Examples:**
```sh
# Back up to current directory
macromog backup

# Back up to a specific location
macromog backup --char-name Squatched --out ~/macro-backups

# Back up in-place (inside the character folder)
macromog backup --in-place

# Back up all characters
macromog backup --all --out ~/macro-backups
```

---

### `list` — List characters and books

```sh
macromog list [flags]
```

Without a character selector, scans the FFXI USER folder and lists every detected character with its book count. With `--char-dir` or `--char-name`, lists all populated books for that character.

**Flags:**
- `--ffxi-path`, `--char-dir`, `--char-name` — see [Global Flags](#global-flags)

**Examples:**
```sh
# List all characters
macromog list

# List books for a specific character
macromog list --char-name Squatched
macromog list --char-dir /path/to/USER/a1b2c3d4
```

**Example output:**
```
FFXI USER: /mnt/games/FFXI/USER

  Squatched (a1b2c3d4)    12 books
  e5f6a7b8                 3 books
```

```
Character: Squatched (a1b2c3d4)

  Book  1  WHM75NIN    10 sets
  Book  5  RDM75NIN     4 sets
  Book 33  (unnamed)    1 set
```

---

## JSON Output

All commands support `--output json` for machine-readable output — useful for scripting or editor integrations.

```sh
macromog --output json list
macromog --output json export --char-name Squatched -o macros.yml
macromog --output json validate macros.yml
```

The `--output` flag may appear before or after the subcommand name. When placed after the subcommand, macromog distinguishes between `text|json` (the global flag) and a filename argument (e.g. `export --output macros.yml`).

### JSON shapes

**`list` (all characters):**
```json
{
  "user_dir": "/path/to/USER",
  "characters": [
    { "id": "a1b2c3d4", "name": "Squatched", "book_count": 12 }
  ]
}
```

**`list` (single character):**
```json
{
  "character": "a1b2c3d4",
  "name": "Squatched",
  "books": [
    { "index": 1, "name": "WHM75NIN", "set_count": 10 }
  ]
}
```

**`export`:**
```json
{ "character": "a1b2c3d4", "path": "squatched_macros_20260620_033000.yml", "ok": true }
```

**`import`:**
```json
{ "character": "a1b2c3d4", "yaml_file": "macros.yml", "sets": 12, "backup_path": "...", "dry_run": false, "ok": true }
```

**`backup`:**
```json
{ "character": "a1b2c3d4", "path": "a1b2c3d4_20260620_033000", "ok": true }
```

**`validate`:**
```json
{ "file": "macros.yml", "valid": true }
```

**`alias set`:**
```json
{ "char_id": "a1b2c3d4", "name": "Squatched" }
```

**`alias remove`:**
```json
{ "char_id": "a1b2c3d4", "removed": true }
```

When operating on multiple characters (`--all`), `export`, `import`, and `backup` emit a JSON array instead of a single object.

---

## Auto-Translate Markers

FFXI auto-translate phrases stored in macros appear as placeholder tokens on export:

```yaml
contents:
  - /ma ≺[07021203]≻ <me>
  - ≺[02020114]≻ good luck!
```

`≺[XXXXXXXX]≻` is a printable stand-in for the binary token (8-digit hex resource ID). These markers are preserved on import and converted back to the binary format the game expects.

Lines containing auto-translate markers are exempt from the 60-character length validation, since the *expanded* auto-translate text is what the game counts.

---

## Common Workflows

### Full macro overhaul

```sh
macromog export                              # export everything
# edit the YAML
macromog validate macros.yml                 # catch errors before importing
macromog import macros.yml                   # validates + auto-backups before writing
```

### Shared macros across characters

```sh
# Export from one character
macromog export --char-name Squatched -o shared.yml

# Import into all characters
macromog import --all shared.yml
```

### Scoped update (one book)

```sh
# Export only book 3
macromog export --scope B3 -o book3.yml

# Edit book3.yml, then import back (only book 3 is touched)
macromog import book3.yml
```

### Starting from scratch with a template

```sh
# Generate a template for book 1
macromog template b1.yml --scope B1 --char-name Squatched

# Fill in macros, validate, then import
macromog validate b1.yml
macromog import b1.yml
```

### Scripting with JSON output

```sh
# Check if a file is valid; parse the result with jq
macromog --output json validate macros.yml | jq '.valid'

# Get the backup path after a backup
macromog --output json backup --char-name Squatched | jq '.path'
```

---

## Environment and Detection

macromog searches common install locations automatically. Supply `--ffxi-path` if detection fails — it should point at the FFXI root directory (the one containing the `USER` subdirectory).

### Windows

```
C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI\USER
C:\Program Files\PlayOnline\SquareEnix\FINAL FANTASY XI\USER
```

### Linux

| Setup | Path searched |
|-------|--------------|
| Default Wine prefix | `~/.wine/drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER` |
| Lutris | `~/Games/<game-name>/drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI/USER` (all subfolders of `~/Games` are scanned) |
| Steam / Proton | `~/.steam/steam/steamapps/compatdata/230330/pfx/drive_c/…` |
| Steam (alt path) | `~/.local/share/Steam/steamapps/compatdata/230330/pfx/drive_c/…` |

**Color output** is enabled automatically when writing to a terminal. Set `NO_COLOR=1` or `TERM=dumb` to disable it.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (details on stderr) |

`validate` exits 1 when the file has validation errors, even if it is otherwise well-formed YAML.
