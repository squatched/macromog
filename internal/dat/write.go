package dat

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

// EncodeText encodes a Go string to a fixed-size NUL-terminated FFXI binary
// field. Special markers (≺...≻) are re-encoded to their binary form;
// Hiragana and Katakana are encoded to Shift-JIS. Unrepresentable characters
// are dropped. The output is exactly size bytes, zero-padded.
func EncodeText(s string, size int) []byte {
	out := make([]byte, size) // zero-initialized; last byte is NUL terminator
	if size <= 0 || s == "" {
		return out
	}
	limit := size - 1 // reserve last byte for NUL
	pos := 0

	for len(s) > 0 && pos < limit {
		// Special marker: ≺token≻
		if strings.HasPrefix(s, specialMarkerStart) {
			rest := s[len(specialMarkerStart):]
			end := strings.Index(rest, specialMarkerEnd)
			if end < 0 {
				s = rest // malformed — skip the opener
				continue
			}
			token := rest[:end]
			for _, b := range encodeSpecialMarker(token) {
				if pos >= limit {
					break
				}
				out[pos] = b
				pos++
			}
			s = rest[end+len(specialMarkerEnd):]
			continue
		}

		r, n := utf8.DecodeRuneInString(s)
		s = s[n:]
		if r == utf8.RuneError {
			continue
		}

		// ASCII
		if r < 0x80 {
			out[pos] = byte(r)
			pos++
			continue
		}

		// Shift-JIS (Hiragana / Katakana)
		if lead, trail, ok := unicodeToShiftJIS(r); ok {
			if pos+1 < limit { // need 2 bytes before the NUL slot
				out[pos] = lead
				out[pos+1] = trail
				pos += 2
			}
			continue
		}
		// Unrepresentable character — drop silently
	}
	return out
}

// encodeSpecialMarker converts a decoded marker token (the text between ≺ and
// ≻) back to its raw FFXI binary bytes.
func encodeSpecialMarker(token string) []byte {
	switch {
	case token == "autotrans:start":
		return []byte{0xEF, 0x27}
	case token == "autotrans:end":
		return []byte{0xEF, 0x28}
	case strings.HasPrefix(token, "element:"):
		n, err := strconv.Atoi(token[8:])
		if err != nil || n < 0 || n > 7 {
			return nil
		}
		return []byte{0xEF, byte(0x1F + n)}
	case len(token) == 10 && token[0] == '[' && token[9] == ']':
		id, err := strconv.ParseUint(token[1:9], 16, 32)
		if err != nil {
			return nil
		}
		return []byte{
			0xFD,
			byte(id >> 24), byte(id >> 16), byte(id >> 8), byte(id),
			0xFD,
		}
	case strings.HasPrefix(token, "byte:"):
		hex := token[5:]
		switch len(hex) {
		case 2:
			b, err := strconv.ParseUint(hex, 16, 8)
			if err == nil {
				return []byte{byte(b)}
			}
		case 4:
			b1, err1 := strconv.ParseUint(hex[:2], 16, 8)
			b2, err2 := strconv.ParseUint(hex[2:], 16, 8)
			if err1 == nil && err2 == nil {
				return []byte{byte(b1), byte(b2)}
			}
		}
	}
	return nil
}

// unicodeToShiftJIS encodes a Hiragana or Katakana rune to a Shift-JIS pair.
// Mirrors the decode table in shiftJISToUnicode.
func unicodeToShiftJIS(r rune) (lead, trail byte, ok bool) {
	// Hiragana U+3041..U+3041+85 → lead=0x82, trail=0x9F+base
	if r >= 0x3041 && r < 0x3041+86 {
		return 0x82, byte(0x9F + int(r) - 0x3041), true
	}
	// Katakana U+30A1..U+30A1+85 → lead=0x83, trail=0x40+base (skip 0x7F)
	if r >= 0x30A1 && r < 0x30A1+86 {
		t := byte(0x40 + int(r) - 0x30A1)
		if t >= 0x7F {
			t++ // 0x7F is not a valid Shift-JIS trail byte
		}
		return 0x83, t, true
	}
	return 0, 0, false
}

// EncodeMacroSet serializes a MacroSet to the 7624-byte on-disk binary format.
// The header includes the MD5 checksum of the payload (bytes 24–7623) as the
// game and POLUtils expect.
func EncodeMacroSet(set MacroSet) []byte {
	out := make([]byte, MacroSetFileSize)
	binary.LittleEndian.PutUint32(out[0:4], MagicVersion)
	binary.LittleEndian.PutUint32(out[4:8], set.HeaderUnknown)
	// bytes 8–23: MD5, filled in below

	offset := HeaderSize
	var all [MacrosPerSet]Macro
	copy(all[:10], set.Ctrl[:])
	copy(all[10:], set.Alt[:])

	for _, m := range all {
		// 4-byte prefix: zeros (already zero-initialized)
		offset += MacroPrefixSize
		for i := 0; i < LineCount; i++ {
			copy(out[offset:], EncodeText(m.Contents[i], LineSize))
			offset += LineSize
		}
		copy(out[offset:], EncodeText(m.Name, NameSize))
		offset += NameSize
	}

	sum := md5.Sum(out[HeaderSize:])
	copy(out[8:24], sum[:])
	return out
}

// WriteMacroSetFile encodes set and writes it to path, creating or overwriting.
func WriteMacroSetFile(path string, set MacroSet) error {
	return os.WriteFile(path, EncodeMacroSet(set), 0o644)
}

// WriteBookTitles writes book names to mcr.ttl (books 1–20) and mcr_2.ttl
// (books 21–40) in dir. Matches game behavior: if all names in a group are
// empty, the corresponding file is deleted (or left absent if it never existed).
func WriteBookTitles(dir string, titles [MaxBooks]string) error {
	if err := writeTitleFile(filepath.Join(dir, "mcr.ttl"), titles[:20]); err != nil {
		return err
	}
	return writeTitleFile(filepath.Join(dir, "mcr_2.ttl"), titles[20:])
}

func writeTitleFile(path string, names []string) error {
	allEmpty := true
	for _, n := range names {
		if n != "" {
			allEmpty = false
			break
		}
	}
	if allEmpty {
		// Match game behavior: do not keep a title file when all entries are empty.
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", filepath.Base(path), err)
		}
		return nil
	}

	out := make([]byte, HeaderSize+len(names)*BookNameSize)
	binary.LittleEndian.PutUint32(out[0:4], MagicVersion)
	// bytes 4–7: unknown flag; write 0
	// bytes 8–23: MD5, filled in below
	for i, name := range names {
		copy(out[HeaderSize+i*BookNameSize:], EncodeText(name, BookNameSize))
	}
	sum := md5.Sum(out[HeaderSize:])
	copy(out[8:24], sum[:])
	return os.WriteFile(path, out, 0o644)
}
