# Frequently Asked Questions

---

## Backups and Recovery

### Does macromog back up my macros before importing?

Yes. Every `import` (unless you pass `--no-backup`) creates a timestamped copy of your character's macro files before writing anything. The backup path is printed at the end of the import so you always know where to find it.

### What does the backup include?

All `mcr*.dat` files (macro sets) and `*.ttl` files (book titles) from your character folder, copied byte-for-byte. No parsing is involved — the files are reproduced exactly as FFXI wrote them.

### Where are my backups stored?

By default inside your character's folder:

```
C:\...\USER\a1b2c3d4\backups\Squatched_a1b2c3d4_backup_YYYYMMDD_HHMMSS\
```

To store backups elsewhere, pass `--backup-dir` on `import`. To create a standalone backup any time without importing:

```
macromog.exe backup --char-name Squatched
macromog.exe backup --char-name Squatched --out C:\Users\you\Documents\macro-backups
```

### How do I restore from a backup?

Quit FFXI first. In PowerShell, copy the backup files back to your character folder:

```powershell
Copy-Item "C:\...\USER\a1b2c3d4\backups\Squatched_a1b2c3d4_backup_20260626_191500\*" `
          "C:\...\USER\a1b2c3d4\" -Force
```

Replace the path with the actual backup directory printed by `import` or `backup`. Start FFXI after the files are in place.

Linux users: see [WINE-PATHS.md](WINE-PATHS.md) for path details; the restore is a plain `cp` of the same files.

### Why does macromog back up DAT files instead of exporting YAML?

Restoring from a DAT backup is a plain file copy — no parsing, no import step, no code path risk. Restoring from a YAML export would require running `import` again, meaning the same code that might have caused the problem is now your recovery path. DAT backup is the safer safety net.

YAML exports are still valuable for human-readable snapshots you can version-control, diff, or share — just not as a substitute for the pre-import backup.

### I used `--no-backup` and something went wrong. Can I recover?

Maybe. If you previously ran `macromog.exe export`, that YAML file captures your macros as of that export — re-run `macromog.exe import <that-file>` to restore them. The result may not be identical if macros were changed in-game after that export.

Going forward, leave backup enabled. The default is on for a reason, kupo.

### Should I make a manual backup before a big reorganization?

Yes, kupo! `macromog.exe backup` is the right tool for a pre-flight snapshot before a complex editing session — even though `import` auto-backs up too, having an explicit checkpoint before you start editing is a good habit.

---

## Setup and Troubleshooting

### The addon says "Macromog install is not configured yet, kupo!" — what do I do?

Macromog couldn't detect your FFXI install automatically. Register it manually with the CLI:

```
macromog.exe config add-install default "C:\Program Files (x86)\PlayOnline\SquareEnix\FINAL FANTASY XI"
```

Replace the path with wherever your FFXI install actually lives. After that, reload the addon (`//lua unload macromog` then `//lua load macromog`) and it should be ready.

### The addon says "Zone once before using Macromog commands, kupo!"

Macromog waits for a zone-in event before allowing any commands, so it knows your character is fully loaded. Zone to any area and the commands will work.

### I see "Alias setup failed: ..." in chat when logging in

Your character's name is already registered as an alias for a different character in the same install — most likely you have two characters with the same name on different servers. Register the second one manually with a distinct alias:

```
macromog.exe list
macromog.exe config set-alias <folder-id> SquatchedLeviathan
```

See [Managing character aliases](GUIDE.md#managing-character-aliases) in the User Guide.

### I staged an import and zoned but the macros didn't change

The zone-in apply failed. You'll see `Zone-in macro apply failed: ...` in chat with the error detail. Your macros are untouched — fix the issue (usually a validation error) and run `//mmog import` again, kupo!

### I ran `//mmog import` twice before zoning — which one gets applied?

Only the most recently staged file. Running `//mmog import` a second time silently replaces the first — only one import can be staged at a time.

---

## YAML Editing

### My import failed with a validation error — how do I find out what's wrong?

Run `macromog.exe validate <file>` for a detailed error report listing every problem by book, set, and slot. Fix the errors and re-import.

### What happens to macros I don't include in the YAML?

It depends on the scope embedded in the file:

- **Full scope** (the default): books absent from the YAML are deleted. If you export everything, remove a book, and import, that book is gone. Only edit macros within books you want to keep — don't remove book entries unless you intend to clear them.
- **Partial scope** (e.g. `--scope B3`): only the scoped books/sets are touched. Everything outside the scope is left completely alone, whether or not it appears in the file.

### Do I need to include all 40 books in the YAML?

No — the format is sparse, and Macromog only writes what's in the file. But be aware of the scope behavior above: with full scope (the default), books you omit entirely will be cleared on import. If you only want to update a subset of your books, use `--scope` when exporting so the YAML's authority is narrowed accordingly.

### My Japanese macro titles are getting cut off

Macro titles have an 8-character limit, but the underlying DAT format uses Shift-JIS encoding where each CJK character takes 2 bytes. In practice this means Japanese titles are limited to 4 characters (4 × 2 bytes = 8). Validation will catch titles that exceed this limit.

### Can I use special characters or symbols in macro names?

Any printable character is allowed in macro names (tabs, newlines, and control codes are not). There is no restriction on symbols or punctuation. Keep the 8-character limit in mind, and remember that CJK characters each count as 2 bytes toward that limit.

---

## Common Mistakes

### I imported the wrong file and overwrote macros I needed.

Don't panic, adventurer — `import` always backs up before writing. Check your character folder's `backups` subfolder for the timestamped backup it created. Restore from it using the steps under [How do I restore from a backup?](#how-do-i-restore-from-a-backup).

### Import changed books I didn't intend to touch.

This happens when the YAML has a `full` scope and books you left out get cleared. Use `--scope` to limit what `import` is allowed to touch:

```
macromog.exe import --scope B3 macros.yml
```

See [Scope Selectors](CLI.md#scope-selectors) for the full syntax.

### I'm not sure what an import would do before I commit.

Use `--dry-run` — it validates the YAML and shows exactly which files would be written, without touching anything:

```
macromog.exe import --dry-run macros.yml
```

---

## Compatibility

### Does Macromog intercept or modify network traffic?

No. Macromog only reads and writes local files — it never injects, modifies, or inspects game packets. The addon listens for Windower's zone-in event (the same way most Windower addons do) solely to know when it's safe to write macro files, not to interact with the network.

### Does Macromog work with private servers?

Yes. Macromog reads and writes the same `.dat` and `.ttl` file format that retail FFXI uses. It doesn't communicate with any server — it only touches local files.

### Does Macromog support both the English and Japanese FFXI clients?

Yes. Both the EN and JP clients use the same binary DAT format with Shift-JIS text encoding. Macromog handles both. Macro adventures know no borders, kupo!

### Does Macromog handle macros with auto-translate in them?

Yes, for round-trips. On export, binary auto-translate tokens are converted to printable placeholders like `≺[07021203]≻`. On import, those placeholders are converted back to the binary format the game expects. Macros with auto-translate export and import cleanly.

Macromog does not know what a token ID means — it cannot tell you that `07021203` is `<Cure III>`. There is no way to author a new auto-translate token from scratch in YAML without first exporting an existing macro that contains it.

**Line-length caveat**: Macromog skips the 60-character line check for any line containing a `≺[...]≻` marker, because the game counts the *expanded* auto-translate text against that limit, not the placeholder. If you edit a line that contains auto-translate tokens, it is your responsibility to keep the expanded result within 60 characters — Macromog cannot verify this.
