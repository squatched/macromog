# FFXI Macro DAT and TTL File Format

Technical reference for the per-character macro files the FFXI client stores under each character's `USER` folder. Macromog's CLI reads these files on export and writes them on import.

Notation uses **1-based** book and set numbers throughout (matching in-game labels):

| Symbol | Meaning |
|--------|---------|
| B# | Macro book 1–40 |
| S# | Macro set 1–10 within a book |

---

## File inventory

Each character folder contains:

| File(s) | Purpose |
|---------|---------|
| `mcr.dat` … `mcr399.dat` | Macro set payloads (one file per book/set pair) |
| `mcr.ttl` | Custom names for books 1–20 |
| `mcr_2.ttl` | Custom names for books 21–40 |

Other filenames (`nmcr*.dat`, etc.) exist in some installs and are **ignored** by macromog.

### Macro set filenames

Files use a single combined index with **leading zeros dropped** in the filename:

```
index = 10 × (book − 1) + (set − 1)
```

| Book | Set | Index | Filename |
|------|-----|-------|----------|
| 1 | 1 | 0 | `mcr.dat` |
| 1 | 2 | 1 | `mcr1.dat` |
| 6 | 9 | 58 | `mcr58.dat` |
| 6 | 10 | 59 | `mcr59.dat` |
| 33 | 1 | 320 | `mcr320.dat` |
| 40 | 10 | 399 | `mcr399.dat` |

Valid indices are **0–399**. Higher indices (e.g. `mcr400.dat`) are not part of the 40×10 grid and are rejected.

---

## Shared file header (24 bytes)

Both `.dat` macro sets and `.ttl` title files begin with the same 24-byte header:

| Offset | Size | Type | Value |
|--------|------|------|-------|
| 0x00 | 4 | `uint32` LE | Magic / version (`1`) |
| 0x04 | 4 | `uint32` LE | Write timestamp (see below) |
| 0x08 | 16 | `uint8[16]` | MD5 digest of payload (bytes 0x18 onward) |

Macromog checks the magic value on read and writes the correct MD5 on write, but does not verify the MD5 on read.

### Unknown field (bytes 0x04–0x07)

The meaning of this field is not definitively known. What is confirmed from a survey of live character files:

- `0x00000000` — file has never been written by the game (untouched slots consistently have this value)
- Non-zero — set by the game at write time; files written in the same save operation share an identical value
- Copying files between characters (e.g. via the macro save-slot system) preserves the original value verbatim — the game does not update this field on load or copy
- `nmcr*.dat` files written by Windower contain `0x00000001` — Windower uses a constant

The `.ttl` title files use the same field and are always stamped with the same value as `mcr.dat` (book 1, set 1) in the same write operation.

**Working theory — write timestamp:** The field appears to be a 32-bit clock value stamped at write time, but the exact unit and epoch are unclear. Attempting to interpret the observed non-zero values as *seconds since the FFXI JP launch date (2002-05-16)* yields plausible calendar dates for most values:

| Observed value | Seconds-since-JP interpretation |
|---|---|
| `0x00169984` | 2002-06-02 |
| `0x0089e9a0` | 2002-08-28 |
| `0x00912758` | 2002-09-03 |
| `0x01074388` | 2002-12-01 |
| `0x031fd205` | 2004-01-12 |
| `0x0ad50be0` | 2008-02-17 |
| `0x0aebeae0` | 2008-03-05 |

However, two values found in the same character directories (`0x7817334c`, `0xc6204571`) produce dates of 2066 and 2107 under that same interpretation — clearly wrong. Those same values read as milliseconds since Windows session start correspond to 23 and 38 days of uptime, which is plausible but cannot be converted to an absolute timestamp without knowing when the system was booted. No single epoch+unit combination fits all observed values.

POLUtils carries a `// TODO: Figure out when to write 0x80000000` comment, implying the developer observed only two values (`0` and one file that happened to have the high bit set) and assumed it was a boolean flag. The survey above shows it is not.

**Macromog writes `0` for this field.** The game does not validate it on read (confirmed by POLUtils writing `0` without issue for years), and correctly replicating it would require reverse-engineering client internals that remain opaque.

---

## Macro set files (`mcr*.dat`)

### Size

Every macro set file is exactly **7624 bytes**:

```
24-byte header + 20 macros × 380 bytes = 7624
```

### Macro struct (380 bytes)

Macros are packed C-style structs with **no inter-field padding**:

```
uint32  prefix     always 0
char    line[6][61]   six lines, 60 usable chars + NUL each
char    name[10]      NUL-terminated; up to 8 content bytes; zeroed to fill
```

| Field | Size | Notes |
|-------|------|-------|
| `prefix` | 4 | Always `0` in observed files |
| Each line | 61 | NUL-terminated Shift-JIS text (see [Text encoding](#text-encoding)) |
| Name | 10 | Fixed `char[10]`; NUL-terminated Shift-JIS; up to 8 content bytes |

Early hypotheses assumed a 9-byte name field; live client files use **10 bytes**.

### Macro order within a set

Each file stores **20 macros** in this order:

1. `ctrl[0]` … `ctrl[9]` — Ctrl keys 1–10
2. `alt[0]` … `alt[9]` — Alt keys 1–10

YAML export remaps slot indices to keyboard order **1–9, then 0** (Ctrl/Alt 10 → key `0`).

### Empty slots

Unused macros are all zero bytes. Export omits empty macros (sparse YAML).

---

## Book title files (`mcr.ttl`, `mcr_2.ttl`)

Custom book names live in title files, not inside individual `mcr*.dat` sets.

| File | Books covered |
|------|---------------|
| `mcr.ttl` | 1–20 |
| `mcr_2.ttl` | 21–40 |

Layout:

```
24-byte header (same as macro sets)
char title[20][16]   per file — 15 usable chars + NUL per book name
```

Title files may be absent; export proceeds with empty book names for those ranges.

---

## Text encoding

All macro lines and names are **Shift-JIS** byte strings in fixed-size `char[]` arrays. The same struct layout is used on EN and JP clients; only the encoded bytes differ.

### ASCII

Bytes `< 0x80` pass through as single-byte characters.

### Shift-JIS

Double-byte sequences use standard Shift-JIS lead/trail ranges. Trail byte `0x7F` is never used; decoders must account for that gap when mapping katakana trails at or above `0x80`.

### FFXI extension bytes

The client embeds binary tokens inside macro text. On export, macromog converts them to printable placeholders (see [SPEC.md](SPEC.md#auto-translate-resource-markers) for YAML form).

| Pattern | Meaning | Export placeholder |
|---------|---------|-------------------|
| `0xFD` + 4-byte ID + `0xFD` | Spell, item, or auto-translate resource | `≺[XXXXXXXX]≻` |
| `0xEF 0x27` | Auto-translate region start | `≺autotrans:start≻` |
| `0xEF 0x28` | Auto-translate region end | `≺autotrans:end≻` |
| `0xEF 0x1F`–`0x26` | Elemental symbols | `≺element:N≻` |

Resource IDs are locale-independent; the client resolves display text per language at render time.

### Japanese practical limits

Because fields are byte-bounded, not glyph-bounded:

| Field | Byte budget | JP practical limit |
|-------|-------------|-------------------|
| Macro name | 8 bytes | 4 kana/kanji (2 bytes each) |
| Macro line | 60 bytes | 30 full-width characters (fewer when mixed with ASCII) |
| Book name | 15 bytes | 7–8 CJK characters depending on code points |

---

## Client locale

Struct sizes and field order are the same across regional clients. JP installs store Shift-JIS in the same `char[]` arrays; EN installs store ASCII and extension tokens. No JP-specific layout variant has been observed.

---

## Test fixtures

Anonymized sample files used by CLI tests live at:

```
internal/dat/testdata/char/
```

Includes representative macro sets (`mcr320.dat`–`mcr329.dat`, pathological `mcr58.dat`, struct-layout `mcr59.dat`) and matching `mcr.ttl` / `mcr_2.ttl` title files.

---

## References

Macromog's parser constants are defined in `internal/dat/dat.go`. Related prior art includes POLUtils and xi-tinkerer macro readers; macromog was validated against live client files and those references.