# Macromog - Behavior Specification

## Overview
Macromog is a Windower 4 Lua addon that lets users manage FFXI macros via YAML files.

## Core Capabilities
1. **Export**
   - Reads all current in-game macros from memory
   - Outputs to `<character_name>_macros.yml` (sparse format)
   - Only includes defined/non-empty entries

2. **Import**
   - Reads and validates a YAML file
   - Offers/automatically creates a timestamped backup of current macros
   - Writes validated macros back into game memory
   - Refreshes UI where possible

3. **Validation**
   - Enforces FFXI constraints
   - Custom schema checks

## YAML Structure (Example)
```yaml
books:
  0:
    name: "WHM Support"
    sets:
      0:
        ctrl:
          0:
            name: "Cure"
            contents:
              - "/ma 'Cure IV' <me>"
              - "/wait 1"
        alt:
          0:
            name: "Esuna"
            contents: ["..."]
      # ... more sets as needed
  # Only populated books are present
