# Macromog

**Kupo!** Your trusty Moogle macro librarian for Final Fantasy XI and Windower 4.

Macromog allows you to manage your entire collection of FFXI macros using simple, readable YAML files. Export your current macros, make changes in a text editor, validate them, and import them back with automatic backups.

Perfect for players with 40 macro books full of job-specific setups, gear swaps, and complex commands.


## WARNING
While Macromog does **not** alter packets or interfere with or automate any game operations, it is still technically against the terms of service so, use at your own risk.

## Features

- Export your current in-game macros (including custom book names) to a structured `.yml` file! Super easy to read and understand instead of that pesky `.dat` format, you're not a machine. Are you?
- Import from YAML with full validation and automatic backup
- Strict validation against FFXI limits (40 books, 10 sets/book, 20 macros/set, 8-char macro titles, 6 lines max, etc.)
- Supports EN & JP FFXI client locales

**Brought to you by Kupomog**, the helpful Moogle archivist!

## Status

Feature complete! This was developed and tested on Linux though using the EN client under Lutris so pure Windows and JP functionality is unconfirmed.

## Quick Start

### 1. Install

1. Download the latest `Macromog-#.#.#.zip` from the [Releases](https://github.com/squatched/macromog/releases) page.
2. Extract the `Macromog` folder into your Windower4 addons directory (e.g. `C:\Windower4\addons\Macromog\`).
3. To auto-load Macromog every time you launch, add these two lines to `Windower4\scripts\init.txt`:
   ```
   // Enable Macromog
   lua load Macromog
   ```
   If you prefer an addon manager, those work fine too — any method that runs `lua load Macromog` is equivalent. For a one-time manual load without editing `init.txt`, type `//lua load macromog` in the in-game chat box.

Linux and Wine users: Macromog supports you as a primary use case. See [docs/WINE-PATHS.md](docs/WINE-PATHS.md) for setup details.

### 2. Log In

When the addon loads you'll see a welcome message and the first time you log in on a character, you'll get a notice that the character has been registered (we'll get to what that means later):

```
Kupomog at your service, kupo! Type //mmog help for commands.
Character 'Squatched' has been registered with this install, kupo!
```

No manual configuration needed — just play the game, adventurer!

### 3. Export Your Macros

Type in the in-game chat box:

```
//mmog export
```

Macromog reads your macro books and writes them to a timestamped YAML file in your addon's `data` folder:

```
Exported to Squatched_macros_20260626_191500.yml, kupo!
```

### 4. Edit Your Macros

Open the file at `Windower4\addons\Macromog\data\Squatched_macros_20260626_191500.yml` in a text editor. Any simple text editor works but there are way better options for dealing with YAML. If you're not familiar with it, YAML is simply a way to represent data in a structured way that's meant for human use. [Here's](https://quickref.me/yaml.html) a simple reference sheet if you're interested but you shouuld just try it! I'm sure you'll figure it out.

Here's a few text editors that will help you deal with YAML better:
- **[VS Code](https://code.visualstudio.com/)** — free; install the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) for inline validation and highlighting
- **[Notepad++](https://notepad-plus-plus.org/)** — free, lightweight, familiar to many FFXI players

The format is straightforward (use spaces, not tabs):

```yaml
books:
  1:
    name: MyMacros
    sets:
      1:
        ctrl:
          1:
            name: Cure
            contents:
              - /ma "Cure IV" <t>
        alt:
          1:
            name: Dia
            contents:
              - /ma "Dia" <t>
```

Macro titles are limited to 8 characters; each macro can have up to 6 lines. The default export is a full snapshot of all your macro books — import treats it the same way, so books you remove from the YAML will be cleared in-game. Edit freely within what's there, and don't delete books, sets, or macros you want to keep!

### 5. Import Your Macros

Once you've changed your macros however you like them, import them back into the game!

```
//mmog import Squatched_macros_20260626_191500.yml
```

Macromog validates the file, then waits for your next zone:

```
Staged Squatched_macros_20260626_191500.yml. Zone once to apply in-game, kupo!
```

Zone anywhere. On zone-in, Macromog backs up your current macros and applies the changes:

```
Macros successfully applied, kupo! (Pre-import backup at Squatched_a1b2c3d4_backup_20260626_191600)
```

Your macros are updated, kupo! For CLI workflows, configuration details, and advanced features, see the **[User Guide](docs/GUIDE.md)**.

## Documentation

- [User Guide](docs/GUIDE.md) — workflows, configuration system, advanced features
- [FAQ](docs/FAQ.md) — backups, recovery, common mistakes
- [CLI Reference](docs/CLI.md) — full command and flag reference
- [YAML Reference](docs/YAML.md) — macro file and config.yml field-by-field schema
- [DAT file format](docs/DAT-FORMAT.md) — FFXI macro `.dat` and `.ttl` binary layout
- [Contributing](docs/CONTRIBUTING.md) — setup, validation, PR workflow

## Repository Structure

- `macromog.lua` — Main addon
- `docs/` — Technical documentation
- `AGENTS.md` — Development and agent guide (repo root for tooling)

## Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for setup steps, commit message guidelines, and PR workflow.

---

*Macromog is a community-driven Windower 4 addon. Contributions welcome!*
