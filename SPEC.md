# Macromog - Behavior Specification

## Overview
Macromog is a Windower 4 Lua addon that lets users manage FFXI macros via clean YAML files.

**Brought to you by Kupomog**, your friendly Moogle macro archivist, kupo!

## Core Capabilities
1. **Export**
   - Reads all current in-game macros (including custom book names) from memory or .dat files
   - Outputs to a sparse `<character_name>_macros.yml`

2. **Import**
   - Reads and validates a YAML file
   - Automatically creates a timestamped backup of current macros before applying changes
   - Writes validated data back into the game

3. **Validation**
   - Full schema validation against FFXI constraints
   - Sparse format support (only defined entries are stored)

## YAML Structure (Example)
```yaml
books:
  0:                    # Book index (0-39)
    name: "rdm75nin"    # Custom book name (editable in-game)
    sets:
      0:                # Set index (0-9)
        ctrl:
          0:
            name: "Cure"   # Macro button title (max 8 chars)
            contents:
              - "/ma 'Cure IV' <me>"
              - "/wait 1"
        alt: { ... }
      # ... additional sets as needed
  # Only populated books/sets/macros are included
  
