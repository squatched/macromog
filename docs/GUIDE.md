# Macromog User Guide

Welcome, adventurer! Kupomog here to walk you through everything Macromog can do for your macro library. Macromog is built around two components that share the same functionality and configuration:

- **The Addon** — A Windower 4 addon for convenient in-game export, import, backup, and validation.
- **The CLI** — A standalone (`macromog-windows-386.exe` on Windows, `macromog-linux-amd64` on Linux) for offline macro management, scripting, and multi-character workflows. It's recommended to rename them to something shorter, like `macromog.exe` for Windows or simply `macromog` for Linux to make things easier and clearer. The examples all assume you have done this.

The addon bundles the CLI and delegates all heavy lifting to it. Both use the same config file, so characters registered in-game are immediately available to the CLI and vice versa. The advantage of this is that you can modify your macros without having to log into the game!

To use the CLI on Windows, open **PowerShell** (right-click the Start button → **Terminal** or **Windows PowerShell**). The binary lives at `Windower4\addons\Macromog\bin\macromog.exe` — navigate there, run it by full path, or add the `bin` folder to your system `PATH` so it works from anywhere.

---

## Configuration

FFXI stores each character's macro files under a cryptic hex folder ID inside the `USER` directory — something like `a1b2c3d4` — and the game can be installed almost anywhere depending on whether you're using Steam, a standalone PlayOnline client, Lutris, or Wine on Linux. Macromog's config is how it bridges that messiness: it remembers where your FFXI install lives and maps each character's folder ID to the name you actually know them by. Once that mapping exists, every CLI command accepts `--char-name Squatched` instead of a full path like `--char-dir "C:\...\USER\a1b2c3d4"`.

The good news: **you don't have to set any of this up yourself.**

### Auto-population

You don't need to configure anything manually to get started. When the addon loads for the first time, it detects your FFXI install and registers it. When you log into a character, it maps their opaque folder ID (like `a1b2c3d4`) to a friendly name:

```
Character 'Squatched' has been registered with this install, kupo!
```

After that, `--char-name Squatched` just works from the CLI — no path hunting required.

Each character is registered the first time you log into them with the addon loaded. If you have mules or alts you want to manage via the CLI, log into each one at least once and you'll see the same registration message. Alternatively, you can register them manually with `macromog config set-alias` — see [Managing character aliases](#managing-character-aliases) below.

### Config file location

| Platform | Path |
|----------|------|
| Windows | `%APPDATA%\macromog\config.yml` |
| Linux | `~/.config/macromog/config.yml` |
| Override | Set the `MACROMOG_CONFIG` env var to an absolute path |

Linux users running under Wine share the same config file with the in-game addon when the Linux home directory is visible inside the Wine prefix (common with Lutris). See [WINE-PATHS.md](WINE-PATHS.md) for details.

### Viewing your config

```
macromog.exe config show
macromog.exe config path
```

### Multiple installs

If you run FFXI from more than one install (Steam + a standalone client, for example), register each one:

```
macromog.exe config add-install steam "C:\Program Files (x86)\Steam\steamapps\common\FINAL FANTASY XI Online"
macromog.exe config add-install retail "C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI"

macromog.exe config set-default-install steam

macromog.exe export --install retail --char-name Squatched
```

### Managing character aliases

Characters are auto-registered by the addon. To add or update one manually:

```
macromog.exe list

macromog.exe config set-alias a1b2c3d4 Squatched

macromog.exe config set-alias a1b2c3d4 Squatched --install steam
```

Once an alias exists, you may use `--char-name` instead of `--char-dir` for all commands. Much easier than remembering that ugly weird character ID!

Aliases must be unique within an install (case-insensitive). If you have two characters with the same name on different servers — both living under the same FFXI `USER` folder — you'll need to give them distinct aliases, e.g. `SquatchedAsura` and `SquatchedLeviathan`.

> **Known limitation:** The addon's auto-registration will fail for the second character if the name is already taken. You'll see `Alias setup failed: ...` in chat when you log into them. Work around it by registering the second character manually with a distinct alias:
> ```
> macromog.exe list                                         # find their folder ID
> macromog.exe config set-alias e5f6a7b8 SquatchedLeviathan
> ```

---

## Scope

By default, `export` and `import` operate on all 40 macro books — a full snapshot. **Scope** lets you narrow that to a specific book, set, or even individual macro slot, so you can work on one part of your library without touching the rest.

The `--scope` flag uses a compact selector syntax built from four components:

| Component | Meaning | Range |
|-----------|---------|-------|
| `B<n>` | Book n | 1–40 |
| `S<n>` | Set n within the current book | 1–10 |
| `C<n>` | Ctrl key n within the current set | 0–9 |
| `A<n>` | Alt key n within the current set | 0–9 |

Combine them left-to-right, use `-` for ranges, `,` for siblings, and `*` for all:

```
B3           book 3
B1-5         books 1 through 5
B1,3,5       books 1, 3, and 5
B3S2         book 3, set 2
B3S2C1       book 3, set 2, ctrl key 1
B3S2A*       book 3, set 2, all alt keys
```

The scope is embedded in the exported YAML. On import, Macromog only touches what the scope declares — books outside the scope are left completely alone. See the [CLI Reference](CLI.md#scope-selectors) for the full syntax. Don't worry about getting it wrong; the backup has your back, kupo!

---

## CLI Workflows

### The standard workflow

Export, edit, and reimport:

```
macromog.exe export --char-name Squatched

# edit the .yml

macromog.exe import Squatched_macros_20260626_191500.yml
```

`import` validates the file and backs up your current macros automatically before writing anything. If validation fails, nothing is written. Run `macromog.exe validate <file>` separately if you want to check a file without committing to the import.

---

### Scoped update — one book

Export a single book, edit it, import only that book:

```
macromog.exe export --scope B3 --char-name Squatched --output book3.yml

macromog.exe import book3.yml
```

The scope is embedded in the YAML at export time. On import, only book 3 is touched; every other book is left alone. See the [CLI Reference](CLI.md#scope-selectors) for the full selector syntax. This means that if you don't include an item within a scope you specify, then that item will be cleared out when you import it and anything there will be lost (well, it'll be in the handy dandy backup so it's not permanent). For example, if you specify scope `B2S3` and then don't have any Ctrl macros in your `.yml` file, on import, any macros you might have in your Ctrl slots will be destroyed.

### Dry run — preview before writing

See exactly what would change without touching any files:

```
macromog.exe import --dry-run Squatched_macros_20260626_191500.yml
```

### Starting from a blank template

Generate a pre-structured YAML for a specific scope instead of exporting first:

```
macromog.exe template full.yml --char-name Squatched

macromog.exe template b1s3.yml --scope B1S3 --char-name Squatched
```

Fill in what you need, delete the rest, then import as normal.

### Shared macros across characters

Export from one character and push to all others:

```
macromog.exe export --char-name Squatched --output shared.yml

macromog.exe import --all shared.yml
```

### Manual backup

`import` backs up automatically, but you can take a manual snapshot any time:

```
macromog.exe backup --char-name Squatched
macromog.exe backup --char-name Squatched --out C:\Users\you\Desktop\macro-backups
macromog.exe backup --all --out C:\Users\you\Desktop\macro-backups
```

Backups are timestamped folders of raw `.dat` files — no parsing required to restore them. See [FAQ.md](FAQ.md#how-do-i-restore-from-a-backup) for restore instructions.

### List characters and books

Discover what's on disk:

```
macromog.exe list

macromog.exe list --char-name Squatched
```

---

## In-Game Commands

The addon's commands use file paths relative to `Windower4\addons\Macromog\data\`. (`//mmog` is, naturally, the same as using `//macromog`.)

| Command | What it does |
|---------|-------------|
| `//mmog export [filename]` | Export current macros to YAML in the `data` folder, you can optionally specify a filename |
| `//mmog import <filename>` | Validate and stage a YAML file; apply on next zone |
| `//mmog validate <filename>` | Check a YAML file without importing |
| `//mmog backup` | Back up current macro files to the `data` folder |
| `//mmog help` | Print available commands |

**The import zone requirement.** FFXI reads macro files when you zone in; writing to them while they're open can corrupt data. Macromog stages the import and applies it at the next zone-in packet, which guarantees the files are not actively held by the client.

**Only one import can be staged at a time.** Running `//mmog import` a second time before zoning silently replaces the first — only the most recently staged file is applied.

**Export filename.** Without an explicit filename, export produces `<CharName>_macros_<YYYYMMDD_HHMMSS>.yml` in the `data` folder. Pass a name to choose your own: `//mmog export mybook.yml`.

---

## Further Reading

- [CLI Reference](CLI.md) — every command, flag, and output format in detail
- [FAQ](FAQ.md) — backups, recovery, and common mistakes
- [SPEC.md](SPEC.md) — YAML schema and FFXI constraint reference
- [WINE-PATHS.md](WINE-PATHS.md) — Linux and Wine path configuration

Happy macro-wrangling, adventurer. Kupomog will be here if you need him, kupo!
