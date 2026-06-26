# v1.0.0
This is everything I want to have done before calling this thing done.

Ultimately, we should have 3 things available in this release:
- `macromog` (Linux amd64)
- `macromog.exe` (Windows 32-bit)
- `macromog-<version>.zip` (Windower 4 plugin with `macromog.exe` bundled)

## CLI
- ~~Support flags of the form `<flag>=<value>` so `--output=json`.~~ ✓ cobra/pflag
- ~~Enable `-` as an output file (stdout naturally)~~ ✓ no path → stdout
- ~~BUG: `bin/macromog template out.yml --scope B1S3A1-5` outputs a yaml that includes FAR more. Also positionals and flags may never be interleaved. SOLUTION: Migrate to pflags or cobra.~~ ✓ cobra migration complete
- ~~Currently, the bins are not available as releases. We should change that and make the bins available!~~

## Plugin
- ~~Figure out Windower 4 plugin packaging and how to make the bins available to be executed by lua.~~
- ~~Have the plugin pick the right bin, x86 v amd64.~~ ✓ always bundle 32-bit Windows CLI
- ~~Expose export functionality.~~
- ~~Expose validation functionality.~~
- ~~Expose backup functionality.~~
- ~~On startup, leverage config to store FFXI path and character names.~~
- Provide user documentation.
- Package the plugin as a release.

## Bugs
- ~~I logged in, I've obviously zoned because the character is registered, but I can't do any //mmog commands? "Zone once before using any macromog commands, kupo!"~~
- ~~//mmog export -> unknown shorthand flag: 'o' in -o~~
- ~~//mmog backup -> Doesn't tell me where/what the backup is.~~
- ~~//mmog import -> Should probably ask for confirmation since this is destructive~~
- ~~Nowhere in the interface does it tell me that the file paths are in the data folder only. Might be worth surfacing.~~
- ~~Need confirmation when we zone that a character has been associated with their hex id for this install (name it, "install <alias>").~~
- ~~Backups go to CHAR folder rather than into Macromog/data dir.~~ ✓ now write to addon data/ dir; folder named `<charName>_<charID>_backup_<ts>` (charName dropped if unknown)
- ~~Importing in game doesn't work. The macros never get updated, then when you load, the old macros are re-written into the macro.dat files (subsequent exports after loading don't show the imported macros)~~ ✓ re-write .dat files on zone-in (0x0A) so they land after FFXI's zone-out flush but before its zone-in read

## To Test
- //mmog import -> Confirm the macros change. When? Only after zoning? Does zoning overwrite them with what's in the game client so the export command is useless?

# v1+
- CLI config: `color: auto|always|never`
- CLI config: `default_output_format: text|json`
- CLI config: backup directory preference
- CLI config healing: on validation failure, try removing the offending key and re-validating; if still invalid, remove its parent and retry; escalate until valid or empty; offer full reset only as last resort
