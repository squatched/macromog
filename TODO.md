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

# Refactor Wine Paths

The Lutris/Wine config-sharing fix works, but the implementation grew by accretion
(one bug, one invariant). The underlying model is simpler than the code suggests.
Follow-up refactor after the working stack is merged.

## Core model (two coordinate systems)

| Layer | Purpose | Example |
|-------|---------|---------|
| **Stored** | What goes in `config.yml` | `/home/squatched/Games/.../pfx/drive_c/Program Files (x86)/...` |
| **Runtime** | What `os.Open` / `ReadDir` need on this GOOS | Linux: same POSIX path; Wine: `Z:\home\squatched\...` |

All path helpers (`hostpath`, `normalizeHostPath`, `OpenPath`, `hostAccessPath`,
`resolveForWine`, `canonicalForWine`) are really **stored ↔ runtime translation**
plus **discover Linux home / Wine prefix once**.

## What is inelegant today

1. **Translation is scattered** — callers must pick `OpenPath` vs `hostAccessPath`
   vs `CanonicalInstallPath` vs `ResolveInstallPath`. Split-brain happened when
   one layer used POSIX stat and another used `Z:`.
2. **`RunningUnderWine()` is implicit state** — heuristics re-derived on every call;
   Wine branches are invisible in Linux CI (how `add-install` once used
   `NormalizePath` while interactive registration used `CanonicalInstallPath`).
3. **Prefix discovery is Lutris-shaped** — `~/Games/*/pfx` probing in
   `findWinePrefixUnderHome` reads as guesswork, not a documented strategy.
4. **Addon logs are misleading** — `registering install ... C:/...` is pre-canonical
   detection; YAML stores POSIX. Rename debug lines (detected vs stored).
5. **No CI fixture for the full round-trip** — unit tests cover helpers, not
   “`C:\...` in → POSIX in YAML → `~/.config/macromog/config.yml` on host.”

## Target shape

### 1. `HostFS` (or `RuntimePaths`) type

Detect once per process:

```go
type HostFS struct {
    GOOS       string
    UnderWine  bool
    LinuxHome  string // /home/squatched
    WinePrefix string // /home/squatched/Games/.../pfx
    ConfigPath string // stored canonical config path
}

func (h *HostFS) Stored(p string) string           // normalize for YAML
func (h *HostFS) Access(p string) (string, error) // for os.Open / ReadDir
```

All filesystem I/O goes through `Access()`. No more “which helper?”

### 2. Single install-registration path in CLI

One `registerInstall(session, name, rawPath)` for cobra `add-install`, interactive
prompts, and auto-detect. Canonicalization lives there only (split already fixed).

### 3. Split discovery from translation

```
discovery/   wine_env.go, lutris.go (or strategy interface)
paths/       host.go (hostpath only), translate.go (C: / Z: / POSIX)
```

### 4. Injectable runtime for tests

Production: `DetectHostFS()`. Tests: `HostFS{UnderWine: true, ...}` so Lutris
layout runs in CI without Wine.

### 5. Thin Lua; clearer logs

`lib/setup.lua` trusts CLI JSON + post-register verify (keep verify). Debug:
`detected root (wine-native):` vs `stored install path (host):`.

### 6. Short doc

`docs/WINE-PATHS.md` (~30 lines): stored POSIX in YAML; Wine accesses `/home/...`
via `Z:`; never `filepath.Join` for `/home/...` on Windows GOOS; prefix discovery
order.

## Keep as-is (do not regress)

- POSIX paths in YAML (Linux + Wine share one file).
- Always map `/home/...` through `Z:` under Wine (no POSIX stat shortcut).
- Post-register verify in `setup.ensure_install`.
- `--output json` before subcommand in `lib/cli.lua`.

## Suggested PR order

1. Merge working wine/config stack (this branch).
2. `refactor(config): introduce HostFS` — mechanical, no behavior change.
3. `test(config): lutris fixture round-trip` — `C:` → YAML → host file.
4. `docs: wine path model` — Kupomog explains it once.
