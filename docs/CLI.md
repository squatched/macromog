# macromog CLI — User Guide

Welcome, adventurer! This guide covers everything you need to manage your FFXI macro books with the `macromog` command-line tool, kupo!

---

## Installation

Download the binary for your platform from the [Releases](https://github.com/squatched/macromog/releases) page and place it somewhere on your `PATH`.

| Platform | Binary |
|----------|--------|
| Windows (32-bit) | `macromog.exe` |
| Linux (64-bit) | `macromog` |

FFXI is a 32-bit client, so the Windower plugin bundles `macromog.exe`. Linux
releases target amd64 only.

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

macromog auto-detects your FFXI install on Windows and Linux. Register installs and give your characters real names with [`config`](#configuration) — much nicer than memorizing those cryptic letter-and-number folder IDs in USER. If detection fails, supply `--ffxi-path` pointing at your FFXI root (the folder that contains `USER`).

---

## Commands

| Command | What it does |
|---------|-------------|
| [`config`](#configuration) | Manage installs, character aliases, and CLI preferences |
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
| `--ffxi-path <path>` | Path to your FFXI install root (raw path only). See [Install resolution](#install-resolution). |
| `--install <name>` | Named FFXI install from config (e.g. `steam`, `lutris`). See [Install resolution](#install-resolution). |
| `--char-dir <path>` | Path to a character's folder inside USER (e.g. `/path/to/USER/a1b2c3d4`). Bypasses character selection entirely. |
| `--char-name <name>` | Friendly character name from config — the sort you'd actually remember. Bypasses selection. |

---

## Configuration

macromog keeps your CLI preferences in a YAML config file, kupo — installs, character names, and a few habits so you don't have to re-type them every command. The file is created automatically on first use and updated as you go.

### Config file location

| Context | Path |
|---------|------|
| Linux | `~/.config/macromog/config.yml` |
| Windows | `%APPDATA%\macromog\config.yml` |
| Running under Wine with a mapped Linux home | `~/.config/macromog/config.yml` (same file as the host shell) |
| Override | `MACROMOG_CONFIG` environment variable (absolute path to config file) |

When the Linux home directory is visible inside a Wine prefix (including Lutris installs where `Z:\home\<user>\.config` exists), macromog prefers the host XDG path so a shell `macromog` and the in-game addon share one config file and POSIX install paths. If detection does not apply on your setup, set `MACROMOG_CONFIG` explicitly in both environments.

### Config schema

```yaml
version: 1

preferences:
  default_offering: true    # optional; absent = true (see Default offering)

default_install: steam      # optional; omitted when unset

installs:
  steam:
    path: /home/adventurer/.steam/steamapps/compatdata/230330/pfx/drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI
    default_character: a1b2c3d4    # optional; folder ID as FFXI stores it
    characters:                     # optional; omitted when empty
      a1b2c3d4:
        name: Squatched
      e5f6a7b8:
        name: AltMule

  lutris:
    path: /home/adventurer/Games/ffxi/drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI
    default_character: a1b2c3d4
    characters:
      a1b2c3d4:
        name: Squatched
```

**Named installs** — register each FFXI root (Steam Proton prefix, Lutris bottle, native Windows install, etc.) under a short name. Use `--install <name>` to select one without typing the full path.

**Character aliases** — FFXI stashes each character in USER under an opaque folder ID (those baffling strings like `a1b2c3d4`). Aliases map a real name you'll recognize to that ID. Scoped per install: the same friendly name on Steam and Lutris can point at different folders.

**Defaults** — the first install you register becomes `default_install`; the first alias on an install becomes `default_character` for that install. Later additions do not change an existing default unless you pass `--set-default` or run an explicit set-default command. When a default is removed, the key is omitted from the file rather than left empty.

**Paths** — stored as absolute, normalized paths: `~` expanded, made absolute, trailing slashes removed. Symlinks are preserved as written (not resolved to their targets).

### Install resolution

macromog picks the FFXI install root in this order:

1. **`--ffxi-path <path>`** — use this path. If it matches a registered install, that install's aliases apply. In an interactive session, an unknown path triggers [install registration](#install-registration).
2. **`--install <name>`** — look up a named install in config.
3. **`default_install`** — if present in config.
4. **Single install in config** — if exactly one install is registered, use it.
5. **Auto-detect** — search common locations (see [Environment and Detection](#environment-and-detection)). If the result matches a registered install path, use that install's context. An unknown path in an interactive session triggers install registration.
6. **Multiple installs, no default** — interactive selection prompt (see below).
7. **Error** — with a hint to run `macromog config add-install`.

`--ffxi-path` accepts filesystem paths only, not install names. Use `--install` for names.

### Character selection

Within the active install, macromog picks the character in this order:

1. **`--char-dir <path>`** — use this folder directly.
2. **`--char-name <name>`** — resolve alias from config for the active install.
3. **`--all`** — every character in the USER folder.
4. **`default_character`** — folder ID stored for the active install, if present (macromog shows the alias in messages when you have one).
5. **Single configured alias** — if exactly one character is listed in config for this install, use it.
6. **USER scan** — discover characters from the filesystem:
   - One character found: use it (prints a notice).
   - Multiple characters on an interactive terminal: selection prompt.
   - Multiple characters on non-interactive input: error — use `--char-dir`, `--char-name`, or `--all`.
7. **Multiple configured aliases, no default** — interactive selection prompt.

When a selection prompt appears because no default is set, macromog may show an optional tip (controlled by [default offering](#default-offering)):

```
Multiple installs configured. Which install for this command?
  [1] steam
  [2] lutris
Enter number:

Tip: macromog config set-default-install steam skips this prompt.
```

Character prompts use the same pattern: *"Which character for this command?"* — never implying you are setting a default in that moment.

### Install registration

When macromog finds a path that is not yet in config (via `--ffxi-path` or auto-detect) and stdin is interactive:

```
Path not in config. Register as install? [Y/n]
Name [steam]:
```

The suggested name is derived from the path (`steam`, `lutris`, `wine`, `playonline`, `default`, or `install`; collisions become `steam.2`, `steam.3`, …). Press Enter to accept the suggestion or type a different name.

- **First registered install** becomes `default_install` automatically.
- **Second install** does not change the default. macromog prints:

  > Added install `lutris`. Default install is still `steam`. Use `macromog config set-default-install lutris` to change.

- **Non-interactive sessions** — use the path for the current command only; no registration prompt.

If you decline registration, the next interactive run with the same path asks again.

### Default offering

`preferences.default_offering` controls whether macromog shows tips about setting defaults when you choose an install or character without one configured. Selection prompts themselves are unchanged — only the optional tips are affected.

```sh
macromog config default-offering false
macromog config default-offering true
```

Accepted values: `true`, `false`, `t`, `f`, `yes`, `no`, `y`, `n` (case-insensitive; stored as `true` or `false`).

macromog prints an informational message whenever this preference changes:

```
Default offering disabled. Install and character selection prompts are unchanged; default-setting tips are suppressed.
```

When absent from config, default offering is enabled (`true`).

To always choose install and character explicitly, remove defaults and disable offering:

```sh
macromog config remove-default-install
macromog config remove-default-character --install steam
macromog config default-offering false
```

Or pass `--install` and `--char-name` on every command.

### Config lifecycle

| Event | Behavior |
|-------|----------|
| First `macromog` invocation | Create `config.yml` with `version: 1` only |
| Normal writes | Atomic write (temp file, then rename) |
| Invalid config (parse or semantic error) | Specific error message; command aborts |

**Interactive recovery** when config cannot be loaded:

```
config.yml is invalid: <reason>

  [B]ack up corrupt file and start fresh
  [Q]uit (fix manually)

Choice [B/q]:
```

Choosing **B** renames the corrupt file to `config.yml.bak.<timestamp>` and writes a fresh empty config. Choosing **Q** exits without modifying the file. Non-interactive sessions exit with an error and do not auto-reset.

Removing an install or alias also removes any `default_install` or `default_character` that pointed at it.

### `config` command

```sh
macromog config <subcommand> [args]
```

| Subcommand | Purpose |
|------------|---------|
| `path` | Print config file location |
| `show` | Dump current config |
| `add-install <name> <path> [--set-default]` | Register an install |
| `remove-install <name>` | Remove an install and its aliases |
| `set-default-install <name>` | Set `default_install` |
| `remove-default-install` | Remove `default_install` |
| `set-alias <char-id> <name> [--install <name>] [--set-default]` | Give a character a friendly name |
| `remove-alias <char-id> [--install <name>]` | Remove an alias |
| `set-default-character <char-id> [--install <name>]` | Set `default_character` by folder ID |
| `remove-default-character [--install <name>]` | Remove `default_character` for an install |
| `default-offering <true\|false>` | Enable or disable default-setting tips |

`--install` on config subcommands selects which install to affect; omitted flags use `default_install`.

**Examples:**

You'll need a character's folder ID once to create an alias — run `macromog list` if you don't have it handy. After that, `--char-name` is all you need day to day.

```sh
# Show where config lives
macromog config path

# Register a Steam Proton prefix
macromog config add-install steam "/home/adventurer/.steam/steamapps/compatdata/230330/pfx/drive_c/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI"

# Turn a1b2c3d4 into a name you'll recognize
macromog config set-alias a1b2c3d4 Squatched

# Alias on a specific install
macromog config set-alias a1b2c3d4 Squatched --install lutris

# Export using a named install and alias
macromog export --install steam --char-name Squatched -o macros.yml
```

When a second alias is added to an install, macromog notes that the default character is unchanged and points to `config set-default-character`.

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

## Multi-character selection

When a command operates on more than one character (`--all`, or an interactive prompt that accepts multiple selections), macromog shows:

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

Single-character resolution order is documented under [Character selection](#character-selection) in Configuration.

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
- `--ffxi-path`, `--install`, `--char-dir`, `--char-name`, `--all` — see [Global Flags](#global-flags) and [Configuration](#configuration)
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
- `--ffxi-path`, `--install`, `--char-dir`, `--char-name`, `--all` — see [Global Flags](#global-flags) and [Configuration](#configuration)
- `--no-backup` — skip the automatic backup (use with care; see [FAQ](FAQ.md#backups-and-recovery))
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
- `--ffxi-path`, `--install`, `--char-dir`, `--char-name` — see [Global Flags](#global-flags) and [Configuration](#configuration)

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

**`config show`:**
```json
{ "path": "/home/adventurer/.config/macromog/config.yml", "config": { "version": 1, "installs": { ... } } }
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

macromog searches common install locations automatically when config does not supply an install. Register paths with [`config add-install`](#config-command) so detection is rarely needed. Supply `--ffxi-path` if detection fails — it should point at the FFXI root directory (the one containing the `USER` subdirectory).

### Windows

| Setup | Path searched |
|-------|--------------|
| Standard PlayOnline install | `C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI\USER` |
| PlayOnline (64-bit Program Files) | `C:\Program Files\PlayOnline\SquareEnix\FINAL FANTASY XI\USER` |
| Steam (default install) | `C:\Program Files (x86)\Steam\steamapps\common\FINAL FANTASY XI Online\USER` |
| Steam (64-bit Program Files) | `C:\Program Files\Steam\steamapps\common\FINAL FANTASY XI Online\USER` |

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
