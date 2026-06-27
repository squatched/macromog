# Frequently Asked Questions

---

## Backups and Recovery

### Does macromog back up my macros before importing?

Yes. Every `import` (unless you pass `--no-backup`) creates a timestamped copy of your character's macro files before writing anything. The backup path is printed at the end of the import so you always know where to find it.

### What does the backup include?

All `mcr*.dat` files (macro sets) and `*.ttl` files (book titles) from your character folder, copied byte-for-byte. No parsing is involved — the files are reproduced exactly as FFXI wrote them.

### Where are my backups stored?

By default: `<char-dir>/backups/<CharName>_<charID>_backup_YYYYMMDD_HHMMSS/`

Use `--backup-dir <path>` on `import` to store backups somewhere else. To create a standalone backup at any time, without importing anything:

```sh
macromog backup --char-name Squatched
macromog backup --char-name Squatched --out ~/macro-backups
```

### How do I restore from a backup?

Quit FFXI first, then copy the files from the backup folder back to your character folder.

**Windows (PowerShell):**

```powershell
Copy-Item "C:\...\USER\a1b2c3d4\backups\Squatched_a1b2c3d4_backup_20260626_191500\*" `
          "C:\...\USER\a1b2c3d4\" -Force
```

**Linux:**

```sh
cp ~/.wine/.../USER/a1b2c3d4/backups/Squatched_a1b2c3d4_backup_20260626_191500/* \
   ~/.wine/.../USER/a1b2c3d4/
```

Replace the path with the actual backup directory printed by `import` or `backup`. Start FFXI after the files are in place.

### Why does macromog back up DAT files instead of exporting YAML?

Restoring from a DAT backup is a plain file copy — no parsing, no import step, no code path risk. Restoring from a YAML export would require running `import` again, meaning the same code that might have caused the problem is now your recovery path. DAT backup is the safer safety net.

YAML exports are still valuable for human-readable snapshots you can version-control, diff, or share — just not as a substitute for the pre-import backup.

### I used `--no-backup` and something went wrong. Can I recover?

Maybe. If you previously ran `macromog export`, that YAML file captures your macros as of that export — re-run `macromog import <that-file>` to restore them. The result may not be identical if macros were changed in-game after that export.

Going forward, leave backup enabled. The default is on for a reason, kupo.

### Should I make a manual backup before a big reorganization?

Yes. `macromog backup` is the right tool for a pre-flight snapshot before a complex editing session — even though `import` auto-backs up too, having an explicit checkpoint before you start editing is a good habit.

---

## Common Mistakes

### I imported the wrong file and overwrote macros I needed.

Check `<char-dir>/backups/` for the timestamped backup that `import` created before it wrote anything. Restore from it using the steps under [How do I restore from a backup?](#how-do-i-restore-from-a-backup).

### Import changed books I didn't intend to touch.

This happens when the YAML has a `full` scope and books you left out get cleared. Use `--scope` to limit what `import` is allowed to touch:

```sh
# Only update book 3; leave everything else alone
macromog import --scope B3 macros.yml
```

See [Scope Selectors](CLI.md#scope-selectors) for the full syntax.

### I'm not sure what an import would do before I commit.

Use `--dry-run` — it validates the YAML and shows exactly which files would be written, without touching anything:

```sh
macromog import --dry-run macros.yml
```
