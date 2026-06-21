package dat_test

import (
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
)

func TestDecodeText_ASCII(t *testing.T) {
	got := dat.DecodeText([]byte("/ma \"Cure\" <me>\x00garbage"))
	if got != `/ma "Cure" <me>` {
		t.Errorf("got %q", got)
	}
}

func TestDecodeText_ResourceMarker(t *testing.T) {
	raw := []byte{0xFD, 0x07, 0x02, 0x12, 0x03, 0xFD}
	got := dat.DecodeText(raw)
	if !strings.Contains(got, "[07021203]") {
		t.Errorf("got %q", got)
	}
	if !strings.HasPrefix(got, "\u227a") {
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