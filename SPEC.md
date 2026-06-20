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

1. **Export** — Reads all current in-game macros (including custom book names) from `.dat` files and outputs a sparse `<character_name>_macros.yml`.
2. **Import** — Reads and validates a YAML file, automatically creates a timestamped backup of current macros, then writes validated data back into the game.
3. **Validation** — Full schema validation against FFXI constraints. Sparse format support (only defined entries are stored).
4. **Backup** — Creates a timestamped backup of all macro `.dat` files without importing.

---

## YAML Format

The YAML file is the shared data format for both the plugin and CLI. It is **sparse**: only books, sets, and macros that contain data are included.

### File Naming

- Exported files: `<character_name>_macros.yml` (or user-specified via flag/argument)
- Example: `Hendrimod_macros.yml`

### Structure

```yaml
version: 1                        # Schema version for future compatibility
character: "Hendrimod"            # Optional metadata
exported_at: "2026-06-20T03:30:00Z"

books:
  1:                              # Book index (1–40)
    name: "WHM75NIN"              # Custom book name (max 15 chars, alphanumeric only)
    sets:
      1:                          # Set index (1–10)
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

### Macros

| Property | Limit | Notes |
|----------|-------|-------|
| Macros per set | 20 | 10 `ctrl` + 10 `alt`, indexed 1–9 then 0 |
| Name (button title) | 8 characters | All printable characters allowed |
| Lines per macro | 6 | — |
| Characters per line | 60 | All printable characters allowed |

---

## JSON Schema

The schema is used by the CLI for validation.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/squatched/macromog/schema/macromog.schema.json",
  "title": "Macromog Macro Definition",
  "type": "object",
  "required": ["version", "books"],
  "properties": {
    "version": { "type": "integer", "const": 1 },
    "character": { "type": "string" },
    "exported_at": { "type": "string", "format": "date-time" },
    "books": {
      "type": "object",
      "patternProperties": {
        "^([1-9]|[1-3][0-9]|40)$": { "$ref": "#/$defs/book" }
      },
      "additionalProperties": false
    }
  },
  "$defs": {
    "book": {
      "type": "object",
      "properties": {
        "name": {
          "type": "string",
          "maxLength": 15,
          "pattern": "^[A-Za-z0-9]*$"
        },
        "sets": {
          "type": "object",
          "patternProperties": {
            "^([1-9]|10)$": { "$ref": "#/$defs/set" }
          },
          "additionalProperties": false
        }
      }
    },
    "set": {
      "type": "object",
      "properties": {
        "ctrl": { "$ref": "#/$defs/macroRow" },
        "alt": { "$ref": "#/$defs/macroRow" }
      }
    },
    "macroRow": {
      "type": "object",
      "patternProperties": {
        "^([0-9])$": { "$ref": "#/$defs/macro" }
      },
      "additionalProperties": false
    },
    "macro": {
      "type": "object",
      "properties": {
        "name": { "type": "string", "maxLength": 8 },
        "contents": {
          "type": "array",
          "maxItems": 6,
          "items": { "type": "string", "maxLength": 60 }
        }
      }
    }
  }
}
```

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
| `export` | Export macros from `.dat` files to YAML | `macromog export --char 0x12345678` |
| `import` | Import from YAML into `.dat` files (auto-backups first) | `macromog import mymacros.yml --char 0x12345678` |
| `validate` | Validate a YAML file against the schema | `macromog validate mymacros.yml` |
| `backup` | Create a timestamped backup of all macro `.dat` files | `macromog backup --char 0x12345678` |
| `list` | List detected characters and macro books | `macromog list` |

### Flags

| Flag | Description |
|------|-------------|
| `--ffxi-path <path>` | Path to FFXI install (auto-detected if possible) |
| `--char <id>` | Character folder (hex ID or path) |
| `--output <file>` / `-o` | Output file for export |
| `--backup` | Auto-backup before import (default: true) |
| `--force` | Overwrite without confirmation |
| `--dry-run` | Show what would happen without writing files |

### UX Notes

- Auto-detect FFXI install path on Windows from common installation locations.
- Friendly, descriptive error messages with corrective suggestions.
- Progress output for large exports.
- Colorized terminal output.

---

## Additional Notes

- Macros are stored per-character in `mcr*.dat` files under the character's folder.
- Reference implementations for DAT parsing: POLUtils, xi-tinkerer.
- Japanese client encoding should be handled when reading/writing character name metadata.
