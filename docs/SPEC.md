# Macromog — Behavior Specification

## Overview

Macromog lets FFXI adventurers manage their macro books via clean, human-editable YAML files, kupo!

**Brought to you by Kupomog**, your friendly Moogle macro archivist.

Macromog ships as two complementary components that share the same YAML format and validation rules:

- **The Plugin** — A Windower 4 addon that provides convenient in-game commands for import, export, backup, and validation while logged in.
- **The CLI** (Command Line Interface) — A standalone binary (`macromog`) that handles all heavy lifting: DAT file parsing, YAML conversion, and schema validation. Works offline with no login required.

The CLI is the source of truth for all conversion and validation logic. The plugin acts as a convenient in-game frontend that delegates to it.

---

## Core Capabilities

Both the plugin and CLI support the following operations:

1. **Export** — Reads current in-game macros from `.dat` files and outputs a sparse YAML file. Scope can be narrowed to specific books, sets, or individual macros.
2. **Import** — Reads and validates a YAML file, automatically creates a timestamped backup, then writes validated data back into the game. Respects the scope embedded in the YAML, clearing any in-scope entries absent from the file.
3. **Template** — Generates a blank YAML file pre-structured for a given scope, ready for the adventurer to fill in.
4. **Validation** — Full schema validation against FFXI constraints. Sparse format support (only defined entries are stored).
5. **Backup** — Creates a timestamped backup of all macro `.dat` files without importing.

---

## YAML Format

The YAML file is the shared data format for both the plugin and CLI. It is **sparse**: only books, sets, and macros that contain data are included.

### File Naming

- Default export filename: `<character_name>_macros_<YYYYMMDD_HHMMSS>.yml`
- Example: `squatched_macros_20260620_033000.yml`
- A timestamp suffix is always added to default filenames to prevent overwriting previous exports.
- An explicit filename argument skips the timestamp: `//macromog export myfile.yml` writes `myfile.yml`.

### Structure

```yaml
version: 1                        # Schema version for future compatibility
character: "squatched"            # Optional: character name (convenience metadata)
exported_at: "2026-06-20T03:30:00Z"  # Optional: ISO 8601 timestamp; omitted in templates

scope:                            # Always present (see "Scoped Export and Import")
  level: full                     # full | book | set | macro
  selections:                     # Present for book/set/macro levels; omitted for full
    - {book: 1}                   # book scope entry
    # - {book: 1, set: 2}         # set scope entry
    # - {book: 1, set: 3, type: ctrl, key: 1}  # macro scope entry

books:
  1:                              # Book index (1–40)
    name: "WHM75NIN"              # Custom book name (max 15 chars, alphanumeric only)
    sets:
      1:                          # Set index (1–10)
        header_unknown: 1234567890  # Optional; bytes 4–7 of DAT header; preserved on round-trip; 0 when generated
        ctrl:
          1:                      # Key index (1–9, then 0)
            name: "Cure"          # Macro button title (max 8 chars, any printable)
            contents:             # Up to 6 lines, max 60 chars each
              - /ma "Cure" <t>
              - /echo "Hope you feel better <t>!"
        alt:
          1:
            name: "Buffs"
            contents:
              - /ma "Protect III" <me> <wait 4>
              - /ma "Shell III" <me> <wait 4>
  5:
    name: "RDM75NIN"
    sets: { ... }
```

### Indexing Rules

- All indices are **1-based** (1–40 for books, 1–10 for sets) to align with in-game designations.
- Ctrl and Alt key indices use the order **1, 2, 3, 4, 5, 6, 7, 8, 9, 0** — matching keyboard layout. The value `10` is never used.

### Auto-Translate Resource Markers

In-game auto-translate phrases and resource references are stored as binary tokens in `.dat` files. On export they appear as printable placeholders:

```yaml
contents:
  - The following is auto-translate cure3: ≺[07021203]≻
  - ≺[02020114]≻≺[02020114]≻Good luck!
```

- `≺` (U+227A) and `≻` (U+227B) delimit a marker; `[XXXXXXXX]` is the 8-digit hex resource ID from the client.
- These markers are unlikely to collide with user-typed macro text.
- **Line-length validation is skipped** for lines containing a resource marker, because the in-game client counts the *expanded* auto-translate text against the 60-character budget, not the placeholder length in YAML.

---

## Constraints

These limits are empirically confirmed against the live FFXI client.

### Macro Books

| Property | Limit | Notes |
|----------|-------|-------|
| Count | 40 | Indices 1–40 |
| Name length | 15 characters | Alphanumeric only. No spaces, symbols, or punctuation. |

### Sets

| Property | Limit | Notes |
|----------|-------|-------|
| Sets per book | 10 | Indices 1–10 |
| Name | — | Sets have no custom names in-game. YAML comments may be used for personal reference. |
| `header_unknown` | Optional `uint32` | Bytes 4–7 of the DAT header (purpose unclear; possibly a write timestamp). Preserved on export; defaults to `0` when generating YAML by hand. The game does not validate this field on read. |

### Macros

| Property | Limit | Notes |
|----------|-------|-------|
| Macros per set | 20 | 10 `ctrl` + 10 `alt`, indexed 1–9 then 0 |
| Name (button title) | 8 characters | Printable characters only; no tabs, newlines, or control codes. The DAT format stores text as Shift-JIS (CJK = 2 bytes each), so the practical limit for Japanese text is 4 CJK characters. |
| Lines per macro | 6 | — |
| Characters per line | 60 | Printable characters only; no tabs, newlines, or control codes. |

---

## The Plugin

The Windower 4 plugin provides in-game commands that delegate to the CLI. It is optional — the CLI is fully functional without it.

### Commands

| Command | Description |
|---------|-------------|
| `//macromog export [filename]` | Export current macros to YAML |
| `//macromog import <filename>` | Import and apply macros from YAML |
| `//macromog backup` | Create a timestamped backup of current macro files |
| `//macromog validate <filename>` | Validate a YAML file without applying it |

---

## The CLI

The CLI (`macromog`) is a standalone binary for offline-first macro management.

### Commands

| Command | Description | Example |
|---------|-------------|---------|
| `export` | Export macros from `.dat` files to YAML | `macromog export /path/to/USER/a1b2c3d4` |
| `import` | Import from YAML into `.dat` files (auto-backups first) | `macromog import mymacros.yml` |
| `template` | Generate a blank YAML template for a given scope | `macromog template out.yml --scope B1S3` |
| `validate` | Validate a YAML file against FFXI constraints | `macromog validate mymacros.yml` |
| `backup` | Create a timestamped backup of all macro `.dat` files | `macromog backup` |
| `list` | List detected characters and macro books | `macromog list` |

### Flags

| Flag | Description |
|------|-------------|
| `--ffxi-path <path>` | Path to FFXI install (auto-detected if possible) |
| `<char-dir>` | Character USER directory (positional) |
| `--char-dir <path>` | Character USER directory; bypasses selection |
| `--char-name <name>` | Character alias (set with `macromog alias`) |
| `--output <file>` / `-o` | Output file for export |
| `--scope <selector>` | Scope selector (repeatable; see "Scoped Export and Import") |
| `--no-backup` | Skip auto-backup before import |
| `--dry-run` | Show what would happen without writing files |

### UX Notes

- Auto-detect FFXI install path on Windows from common installation locations.
- Friendly, descriptive error messages with corrective suggestions.
- Colorized terminal output.

---

## Additional Notes

- Macros are stored per-character in `mcr*.dat` files under the character's USER folder. See [DAT-FORMAT.md](DAT-FORMAT.md) for the on-disk binary layout.

## Scoped Export and Import

The scope system controls which portion of a character's macro library a YAML file has authority over. The scope is embedded in the YAML at export time. On import, the embedded scope tells macromog exactly what territory it may alter — and equally important, what it must leave alone.

### Scope Levels

| Level | Authority | Within scope | Outside scope |
|-------|-----------|--------------|---------------|
| `full` | All 40 books | Content written; absent books deleted | — (nothing is outside) |
| `book` | Specified books only | Content written; absent sets deleted; empty books' `.dat` files deleted | Other books untouched |
| `set` | Specified (book, set) pairs | Entire set overwritten; absent macro slots zeroed | Other sets untouched |
| `macro` | Specified individual slots | Only those slots updated | Everything else untouched; nothing deleted |

**Empty book handling**: When a book is within scope but has no content in the YAML, all of its `mcr*.dat` files are deleted and its title is cleared from the `.ttl` file. This matches FFXI's own behavior — the client does not create `.dat` files for empty books.

### Scope Selector Syntax

Both `export` and `import` accept one or more `--scope` flags using this selector syntax:

```
--scope <selector>[,<selector>...]
```

A selector is built left-to-right from these components:

| Component | Meaning | Example |
|-----------|---------|---------|
| `B<n>` | Book n (1–40) | `B1`, `B5` |
| `S<n>` | Set n (1–10) within the current book | `B1S3` |
| `C<n>` | Ctrl key n (0–9) within the current set | `B1S3C2` |
| `A<n>` | Alt key n (0–9) within the current set | `B1S3A1` |
| `<n>-<m>` | Range at the current component level | `B1-5`, `B1S2-4` |
| `*` | All valid values at the current level | `B*`, `B1S*`, `B1S3C*` |

**Comma siblings**: A comma within a selector creates a sibling at the same level. A bare number after a comma inherits the previous component type:

```
B1,3,5       → books 1, 3, and 5
B1S2,4       → book 1, sets 2 and 4
B1S3A1,3     → book 1, set 3, alt keys 1 and 3
B1S3A1,C2    → book 1, set 3, alt+1 and ctrl+2
```

Multiple `--scope` flags are combined for disjoint selectors:

```
--scope B1S3A1 --scope B5S2C4
```

Mixing levels across `--scope` flags (e.g., `--scope B1` with `--scope B2S3`) is an error.

**Scope level inference**: The level is determined by the deepest component present in any selector.

| Deepest component | Inferred scope level |
|-------------------|---------------------|
| `C` or `A` | macro |
| `S` (no `C`/`A`) | set |
| `B` (no `S`/`C`/`A`) | book |
| `*` alone, or `B*` without further qualification | full |

### Export with `--scope`

```sh
macromog export --scope B1,5        # books 1 and 5 only
macromog export --scope B1S3        # book 1, set 3 only
macromog export --scope B1S3C1 --scope B5S2A3   # two specific macros
macromog export                     # no flag → full scope
```

Only content within the scope is written to the YAML. The `scope` field is always embedded in the output, reflecting what was exported.

**Default (no `--scope`)**: Full scope — all non-empty macros are exported, and `scope: {level: full}` is written. This is the standard workflow.

### Import Scope Behavior

Import reads the `scope` field from the YAML and applies clearing within that scope only.

**No `--scope` flag on import** (standard workflow):

| YAML `scope` field | Import behavior |
|-------------------|-----------------|
| `level: full` | Full clearing — absent books deleted, standard workflow |
| `level: book` | Only scoped books touched; others untouched |
| `level: set` | Only scoped sets touched; others untouched |
| `level: macro` | Only scoped slots updated; nothing deleted |
| Absent (legacy file) | Write-only — files are written but nothing is deleted (backward compatible) |

**`--scope` flag on import** overrides the YAML's embedded scope:

```sh
macromog import macros.yml --scope B1   # import only book 1, even from a full-scope YAML
macromog import macros.yml --scope *    # force full scope on a legacy or narrow YAML
```

If the import `--scope` **exceeds** the YAML's embedded scope — claiming authority over books, sets, or macros that have no content in the YAML — a confirmation prompt is required:

```
Warning: scope override expands authority to book 4, which has no content in
this YAML — book 4 will be cleared. Proceed? [y/N]
```

If the import `--scope` is a **subset or equal** of the YAML's embedded scope, no confirmation is needed; the narrower scope is applied silently.

### The Standard Workflow (Full Scope)

```sh
macromog export                    # → squatched_macros_20260620_033000.yml
#   (edit the YAML — add, change, remove macros as desired)
macromog import squatched_macros_20260620_033000.yml
```

Because the export embedded `scope: {level: full}`, import has authority over all 40 books. Books removed from the YAML have their `.dat` files deleted. No warnings, no prompts.

---

## Template Command

`macromog template` generates a blank YAML file pre-structured for a given scope. Every macro slot within scope is present with an empty `name` and six empty `contents` lines, ready for the adventurer to fill in and remove what they don't need.

```sh
macromog template out.yml                  # full template: all 40 books, 10 sets, 20 macros
macromog template out.yml --scope B1S3     # only book 1, set 3
macromog template out.yml --scope B1S3A1,C2    # only those two macro slots
macromog template out.yml --char-name Squatched  # embed character name in template
```

Template files include `version` and `scope`. They do **not** include `exported_at` (templates are not exports). The `character` field is included only when `--char-name` is provided.

### Template Structure Example

`macromog template out.yml --scope B1S3A1,C2`:

```yaml
version: 1
scope:
  level: macro
  selections:
    - {book: 1, set: 3, type: alt, key: 1}
    - {book: 1, set: 3, type: ctrl, key: 2}
books:
  1:
    name: ""
    sets:
      3:
        alt:
          1:
            name: ""
            contents:
              - ""
              - ""
              - ""
              - ""
              - ""
              - ""
        ctrl:
          2:
            name: ""
            contents:
              - ""
              - ""
              - ""
              - ""
              - ""
              - ""
```
