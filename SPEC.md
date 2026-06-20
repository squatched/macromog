# Macromog - Behavior Specification

## Overview
Macromog is a Windower 4 Lua addon that lets users manage FFXI macros via clean YAML files.

**Brought to you by Kupomog**, your friendly Moogle macro archivist, kupo!

## Core Capabilities
1. **Export**
   - Reads all current in-game macros (including custom book names) from memory or .dat files
   - Outputs to a sparse `<character_name>_macros.yml`
   - Book indices in exported files are always **0-based (0–39)** — this is the canonical on-disk format and unambiguous for round-tripping

2. **Import**
   - Reads and validates a YAML file
   - Automatically creates a timestamped backup of current macros before applying changes
   - Writes validated data back into the game

3. **Validation**
   - Full schema validation against FFXI constraints
   - Sparse format support (only defined entries are stored)

## YAML Structure (Example)

Exported files always use 0-based book indices. Hand-written import files may
use 1-based indices (1–40) — the addon detects and normalizes them on load.

```yaml
books:
  0:                    # Book index — 0-based in exported files (0-39)
    name: "rdm75nin"    # Custom book name (editable in-game)
    sets:
      0:                # Set index (0-9)
        ctrl:
          1:
            name: "Cure"   # Macro button title (max 8 chars)
            contents:
              - "/ma 'Cure IV' <me>"
              - "/wait 1"
        alt: { ... }
      # ... additional sets as needed
  # Only populated books/sets/macros are included
```

## Constraints (Validation Rules)

### Macro Books
- **Count**: Up to **40 books**
- **Index convention**: 1-based by default (1–40 in YAML). If any book uses index 0, the entire file is treated as 0-based (0–39). This lets experienced users write 0-based YAML while keeping a friendly default for everyone else.
- **Name**: User-editable in-game
  - **Recommended max**: **32 characters** (conservative; common usage like "rdm40whm" fits easily)
  - Allowed: Alphanumeric + basic punctuation/underscores/spaces

### Sets
- **10 sets per book** (0–9)
- Sets are numbered (no custom names)

### Macros
- **20 macros per set** (typically grouped as `ctrl[0-9]` + `alt[0-9]`)
- **Macro name (button title)**: **Maximum 8 characters** (official limit)
- **Contents**: Maximum **6 lines** per macro
- **Per-line length**: UI-limited (exact public number undocumented but sufficient for typical commands). 
  - **Recommended validation cap**: **128–200 characters** per line (conservative/safe)
- **Allowed characters per line**: Printable ASCII, double quotes for multi-word names, `< >` target specifiers, spaces, basic punctuation. Japanese support in JP client.

## Commands (Planned)
- `//macromog export [optional filename]`
- `//macromog import <filename>`
- `//macromog backup`
- `//macromog validate <filename>`

## Additional Notes
- Macros are stored per-character in `mcr*.dat` files.
- Full read/write may use memory access or DAT parsing (tools like POLUtils / xi-tinkerer exist as references).
- Support for all FFXI locales (encoding considerations for JP client).
