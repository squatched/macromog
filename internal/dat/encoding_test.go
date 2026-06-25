package dat_test

import (
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
)

const specialStart = "\u227a"

func TestDecodeText_ASCII(t *testing.T) {
	got := dat.DecodeText([]byte("/ma \"Cure\" <me>\x00garbage"))
	if got != `/ma "Cure" <me>` {
		t.Errorf("got %q", got)
	}
}

func TestDecodeText_Empty(t *testing.T) {
	if got := dat.DecodeText(nil); got != "" {
		t.Errorf("nil = %q, want empty", got)
	}
	if got := dat.DecodeText([]byte{0}); got != "" {
		t.Errorf("NUL-only = %q, want empty", got)
	}
}

func TestDecodeText_ResourceMarker(t *testing.T) {
	raw := []byte{0xFD, 0x07, 0x02, 0x12, 0x03, 0xFD}
	got := dat.DecodeText(raw)
	if !strings.Contains(got, "[07021203]") {
		t.Errorf("got %q", got)
	}
	if !strings.HasPrefix(got, specialStart) {
		t.Errorf("expected special marker start, got %q", got)
	}
}

func TestDecodeText_AutoTransRegion(t *testing.T) {
	raw := []byte{0xEF, 0x27, 0xEF, 0x28}
	got := dat.DecodeText(raw)
	if !strings.Contains(got, "autotrans:start") || !strings.Contains(got, "autotrans:end") {
		t.Errorf("got %q", got)
	}
}

func TestDecodeText_ElementalMarkers(t *testing.T) {
	raw := []byte{0xEF, 0x1F, 0xEF, 0x26}
	got := dat.DecodeText(raw)
	if !strings.Contains(got, "element:0") || !strings.Contains(got, "element:7") {
		t.Errorf("got %q", got)
	}
}

func TestDecodeText_ShiftJIS(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		want string
	}{
		{
			name: "hiragana",
			// あい (82 A0, 82 A2)
			raw:  []byte{0x82, 0xA0, 0x82, 0xA2},
			want: "あい",
		},
		{
			name: "katakana",
			// アイ (83 41, 83 43)
			raw:  []byte{0x83, 0x41, 0x83, 0x43},
			want: "アイ",
		},
		{
			name: "katakana past skipped 0x7F trail",
			// ムメ (83 80, 83 81) — trail bytes at/above 0x80
			raw:  []byte{0x83, 0x80, 0x83, 0x81},
			want: "ムメ",
		},
		{
			name: "mixed ascii and kana",
			raw:  []byte{'/', 'm', 'a', ' ', 0x82, 0xA0, 0x82, 0xA2},
			want: "/ma あい",
		},
		{
			name: "unknown sjis pair",
			raw:  []byte{0x82, 0x40},
			want: specialStart + "byte:8240" + "\u227b",
		},
		{
			name: "lone lead byte",
			raw:  []byte{0x82},
			want: specialStart + "byte:82" + "\u227b",
		},
		{
			name: "high ascii without valid trail",
			raw:  []byte{0xE0, 0x20},
			want: specialStart + "byte:E0" + "\u227b" + " ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dat.DecodeText(tt.raw); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
