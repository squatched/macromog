package dat_test

import (
	"testing"

	"github.com/squatched/macromog/internal/dat"
)

func TestFileIndex(t *testing.T) {
	tests := []struct {
		book, set, want int
	}{
		{1, 1, 0},
		{1, 2, 1},
		{6, 9, 58},
		{6, 10, 59},
		{33, 1, 320},
		{40, 10, 399},
	}
	for _, tt := range tests {
		if got := dat.FileIndex(tt.book, tt.set); got != tt.want {
			t.Errorf("FileIndex(%d,%d) = %d, want %d", tt.book, tt.set, got, tt.want)
		}
	}
}

func TestParseFileIndex(t *testing.T) {
	book, set := dat.ParseFileIndex(320)
	if book != 33 || set != 1 {
		t.Errorf("ParseFileIndex(320) = (%d,%d), want (33,1)", book, set)
	}
}

func TestMacroFileName(t *testing.T) {
	if got := dat.MacroFileName(1, 1); got != "mcr.dat" {
		t.Errorf("MacroFileName(1,1) = %q, want mcr.dat", got)
	}
	if got := dat.MacroFileName(33, 1); got != "mcr320.dat" {
		t.Errorf("MacroFileName(33,1) = %q, want mcr320.dat", got)
	}
}

func TestParseMacroFileName(t *testing.T) {
	cases := map[string]int{
		"mcr.dat":    0,
		"mcr1.dat":   1,
		"mcr320.dat": 320,
	}
	for name, want := range cases {
		got, ok := dat.ParseMacroFileName(name)
		if !ok || got != want {
			t.Errorf("ParseMacroFileName(%q) = (%d,%v), want (%d,true)", name, got, ok, want)
		}
	}
	for _, name := range []string{"nmcr.dat", "mcr.ttl", "foo.dat"} {
		if _, ok := dat.ParseMacroFileName(name); ok {
			t.Errorf("ParseMacroFileName(%q) should be false", name)
		}
	}
}

func TestYAMLKey(t *testing.T) {
	want := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	for slot, w := range want {
		if got := dat.YAMLKey(slot); got != w {
			t.Errorf("YAMLKey(%d) = %d, want %d", slot, got, w)
		}
	}
}