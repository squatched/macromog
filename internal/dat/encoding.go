package dat

import (
	"fmt"
	"strings"
)

// FFXI special-string markers mirror POLUtils FFXIEncoding conventions.
// ≺ (U+227A) and ≻ (U+227B) wrap decoded extension tokens.
const (
	specialMarkerStart = "\u227a"
	specialMarkerEnd   = "\u227b"
)

// DecodeText converts a NUL-terminated FFXI macro text field to a Go string.
// ASCII and FFXI extension bytes (auto-translate resources, elemental markers)
// are handled; valid Shift-JIS double-byte sequences are passed through as the
// decoded rune when the pair is recognized.
func DecodeText(raw []byte) string {
	end := len(raw)
	for i, b := range raw {
		if b == 0 {
			end = i
			break
		}
	}
	if end == 0 {
		return ""
	}

	var sb strings.Builder
	data := raw[:end]
	for i := 0; i < len(data); {
		b := data[i]

		// Elemental symbol: 0xEF followed by 0x1F..0x26
		if b == 0xEF && i+1 < len(data) && data[i+1] >= 0x1F && data[i+1] <= 0x26 {
			sb.WriteString(specialMarkerStart)
			sb.WriteString(fmt.Sprintf("element:%d", data[i+1]-0x1F))
			sb.WriteString(specialMarkerEnd)
			i += 2
			continue
		}

		// Auto-translator region start/end: 0xEF 0x27 or 0xEF 0x28
		if b == 0xEF && i+1 < len(data) && (data[i+1] == 0x27 || data[i+1] == 0x28) {
			sb.WriteString(specialMarkerStart)
			sb.WriteString("autotrans:")
			if data[i+1] == 0x27 {
				sb.WriteString("start")
			} else {
				sb.WriteString("end")
			}
			sb.WriteString(specialMarkerEnd)
			i += 2
			continue
		}

		// Resource reference (spells, items, auto-translate phrases): 0xFD <id> 0xFD
		if b == 0xFD && i+5 < len(data) && data[i+5] == 0xFD {
			id := uint32(data[i+1])<<24 | uint32(data[i+2])<<16 | uint32(data[i+3])<<8 | uint32(data[i+4])
			sb.WriteString(specialMarkerStart)
			sb.WriteString(fmt.Sprintf("[%08X]", id))
			sb.WriteString(specialMarkerEnd)
			i += 6
			continue
		}

		// ASCII
		if b < 0x80 {
			sb.WriteByte(b)
			i++
			continue
		}

		// Shift-JIS double-byte (simplified; full tables live in POLUtils)
		if i+1 < len(data) && isShiftJISLead(b) && isShiftJISTrail(data[i+1]) {
			r := decodeShiftJIS(b, data[i+1])
			if r != '\uFFFD' {
				sb.WriteRune(r)
			} else {
				sb.WriteString(specialMarkerStart)
				sb.WriteString(fmt.Sprintf("byte:%02X%02X", b, data[i+1]))
				sb.WriteString(specialMarkerEnd)
			}
			i += 2
			continue
		}

		sb.WriteString(specialMarkerStart)
		sb.WriteString(fmt.Sprintf("byte:%02X", b))
		sb.WriteString(specialMarkerEnd)
		i++
	}

	return sb.String()
}

func isShiftJISLead(b byte) bool {
	return (b >= 0x81 && b <= 0x9F) || (b >= 0xE0 && b <= 0xEF)
}

func isShiftJISTrail(b byte) bool {
	return (b >= 0x40 && b <= 0x7E) || (b >= 0x80 && b <= 0xFC)
}

// decodeShiftJIS provides minimal Shift-JIS decoding for common macro text.
// Unknown pairs return U+FFFD so callers can emit a byte marker instead.
func decodeShiftJIS(lead, trail byte) rune {
	if r, ok := shiftJISToUnicode(lead, trail); ok {
		return r
	}
	return '\uFFFD'
}

// shiftJISToUnicode maps a Shift-JIS pair to Unicode. The table is intentionally
// small; unlisted pairs are surfaced as byte markers in DecodeText.
func shiftJISToUnicode(lead, trail byte) (rune, bool) {
	// Hiragana block (lead 0x82, trail 0x9F-0xF1) — common in JP macros.
	if lead == 0x82 && trail >= 0x9F {
		base := int(trail) - 0x9F
		if base >= 0 && base < 86 {
			return rune(0x3041 + base), true // ぁ..ゖ approx range
		}
	}
	// Katakana block (lead 0x83, trail 0x40-0x96). SJIS skips 0x7F as a trail byte.
	if lead == 0x83 && trail >= 0x40 {
		base := int(trail) - 0x40
		if trail > 0x7E {
			base--
		}
		if base >= 0 && base < 86 {
			return rune(0x30A1 + base), true
		}
	}
	return 0, false
}