# v1.0.0
This is everything I want to have done before calling this thing done.

Ultimately, we should have 5 things available in this release:
- macromog Windows x86 bin
- macromog Windows amd64 bin
- macromog linux x86 bin
- macromog linux amd64 bin
- archive (.zip?) containing the Windower 4 plugin with the Windows bins included

## CLI
- ~~Support flags of the form `<flag>=<value>` so `--output=json`.~~ ✓ cobra/pflag
- ~~Enable `-` as an output file (stdout naturally)~~ ✓ no path → stdout
- ~~BUG: `bin/macromog template out.yml --scope B1S3A1-5` outputs a yaml that includes FAR more. Also positionals and flags may never be interleaved. SOLUTION: Migrate to pflags or cobra.~~ ✓ cobra migration complete
- Currently, the bins are not available as releases. We should change that and make the bins available!

## Plugin
- Figure out Windower 4 plugin packaging and how to make the bins available to be executed by lua.
- Have the plugin pick the right bin, x86 v amd64.
- Expose export functionality.
- Expose validation functionality.
- Expose backup functionality.
- On startup, leverage config to store FFXI path and character names.
- Provide user documentation.
- Package the plugin as a release.

## Bugs/Rough Edges
- Every time the CLI is invoked, a cmd window pops up very briefly. Not polished/is distracting.
- Remove the x64 version of the Windows build of the CLI. FFXI is a 32 bit application so if they can play FFXI, they can run the 32 bit version and it's not like we need the provisions x64 affords.

# v1+
- CLI config: `color: auto|always|never`
- CLI config: `default_output_format: text|json`
- CLI config: backup directory preference
- CLI config healing: on validation failure, try removing the offending key and re-validating; if still invalid, remove its parent and retry; escalate until valid or empty; offer full reset only as last resort
