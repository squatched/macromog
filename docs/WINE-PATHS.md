# Wine path model

Macromog keeps **one config file** for both the Linux CLI and the in-game
addon running under Wine. Paths in `config.yml` are always **stored POSIX**
paths on the Linux host (for example `/home/you/Games/.../pfx/drive_c/...`).

## Two coordinate systems

| Layer | Used for | Example |
|-------|----------|---------|
| **Stored** | `config.yml`, CLI JSON | `/home/you/Games/ffxi/pfx/drive_c/Program Files (x86)/...` |
| **Runtime** | `os.Open`, `ReadDir` on this GOOS | Linux: same path; Wine: `Z:\home\you\Games\...` |

The Go `HostFS` type detects the runtime once and translates between them:

- `Stored()` — canonicalize a detected or user-supplied path for YAML
- `Access()` — convert a stored path for filesystem calls

## Rules that must not regress

1. **POSIX in YAML** — never store `C:\...` in config; Wine and Linux share one file.
2. **`/home/...` always maps through `Z:` under Wine** — do not stat POSIX paths directly in the prefix; that caused split-brain reads/writes.
3. **Use `hostpath()` for Linux host segments** — never `filepath.Join` a `/home/...` path on Windows GOOS (it becomes `\home\...`).
4. **Post-register verify** — the addon re-reads config after `add-install` before marking ready.

## Prefix discovery (Lutris-shaped)

When `WINEPREFIX` is unset, discovery scans `~/Games/*` for a prefix containing
the FFXI `USER` tree, then falls back to `drive_c`, then `~/.wine`. This matches
common Lutris layouts but is heuristic — set `WINEPREFIX` when auto-detection fails.

## Debugging

Addon debug logs distinguish:

- `detected ffxi root (wine-native):` — path from Windower before CLI canonicalization
- `stored install path (host):` — POSIX path returned in config JSON after registration
