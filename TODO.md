# v1.0.0
This is everything I want to have done before calling this thing done.

Ultimately, we should have 5 things available in this release:
- macromog Windows x86 bin
- macromog Windows amd64 bin
- macromog linux x86 bin
- macromog linux amd64 bin
- archive (.zip?) containing the Windower 4 plugin with the Windows bins included

## CLI
- When exporting densely, we need to _NOT_ include double quotes by default around everything I think. Will that create malformed YAML? The reasoning is that in general, we don't want that because it creates the expectation that it should be there and users will mess themselves up by using single quotes inside of the double quotes we provide when double quotes are what's needed around spell names in macros.
- Add a `--dense` flag to export. Whatever scope is exported, all macros should exist in the output yaml whether they're empty or not.
- Support flags of the form `<flag>=<value>` so `--output=json`.
- Add a configuration system that stores in appdata or ~/.config that stores FFXI install dir(s) so if the user's FFXI is in a place we haven't thought of, their usage of the tool is much more ergonomic. We should also consider moving character aliases there rather than poluting the FFXI install dir.
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
