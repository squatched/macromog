package dat_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestReadMacroSet_Book33Set1(t *testing.T) {
	charDir := testdata.CharDir()
	data, err := os.ReadFile(filepath.Join(charDir, "mcr320.dat"))
	if err != nil {
		t.Fatal(err)
	}
	set, err := dat.ReadMacroSet(data)
	if err != nil {
		t.Fatal(err)
	}
	if set.Ctrl[0].Name != "B33S1" {
		t.Errorf("ctrl[0].Name = %q, want B33S1", set.Ctrl[0].Name)
	}
	for i := 1; i < 10; i++ {
		if !set.Ctrl[i].Empty() {
			t.Errorf("ctrl[%d] should be empty", i)
		}
	}
	for i := 0; i < 10; i++ {
		if !set.Alt[i].Empty() {
			t.Errorf("alt[%d] should be empty", i)
		}
	}
}

func TestReadMacroSet_StructTest(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdata.CharDir(), "mcr59.dat"))
	if err != nil {
		t.Fatal(err)
	}
	set, err := dat.ReadMacroSet(data)
	if err != nil {
		t.Fatal(err)
	}

	if set.Ctrl[0].Name != "Ctrl1" {
		t.Errorf("ctrl[0].Name = %q, want Ctrl1", set.Ctrl[0].Name)
	}
	if set.Ctrl[9].Name != "Ctrl0" {
		t.Errorf("ctrl[9].Name = %q, want Ctrl0", set.Ctrl[9].Name)
	}
	if set.Ctrl[0].Contents[0] != "Line 1" {
		t.Errorf("ctrl[0] line1 = %q", set.Ctrl[0].Contents[0])
	}

	if !set.Alt[0].Empty() {
		t.Errorf("alt[0] should be empty (unused alt slot 1), got %#v", set.Alt[0])
	}
	if set.Alt[1].Name != "Line1" || set.Alt[1].Contents[0] != "Contents" {
		t.Errorf("alt[1] = name %q contents %#v", set.Alt[1].Name, set.Alt[1].Contents)
	}
	if set.Alt[2].Name != "Line2" || set.Alt[2].Contents[1] != "Line 2" {
		t.Errorf("alt[2] = name %q contents %#v", set.Alt[2].Name, set.Alt[2].Contents)
	}
	if set.Alt[3].Contents[1] != "Line 2" || set.Alt[3].Contents[3] != "Line 4" {
		t.Errorf("alt[3] skip-lines = %#v", set.Alt[3].Contents)
	}
}

func TestReadMacroSet_Pathological(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdata.CharDir(), "mcr58.dat"))
	if err != nil {
		t.Fatal(err)
	}
	set, err := dat.ReadMacroSet(data)
	if err != nil {
		t.Fatal(err)
	}

	if set.Ctrl[8].Name != "12345678" {
		t.Errorf("ctrl[8].Name = %q", set.Ctrl[8].Name)
	}
	if set.Ctrl[9].Name != "Testing" {
		t.Errorf("ctrl[9].Name = %q, want Testing", set.Ctrl[9].Name)
	}

	line2 := set.Ctrl[9].Contents[1]
	if want := "The following is auto-translate cure3: "; line2[:len(want)] != want {
		t.Errorf("line2 prefix = %q", line2[:len(want)])
	}
	if !contains(line2, "[07021203]") {
		t.Errorf("line2 missing Cure III resource marker: %q", line2)
	}

	line3 := set.Ctrl[9].Contents[2]
	if !contains(line3, "[02020114]") {
		t.Errorf("line3 missing Good luck marker: %q", line3)
	}
	if !contains(line3, "Good luck!") {
		t.Errorf("line3 missing typed text: %q", line3)
	}
}

func TestReadBookTitles(t *testing.T) {
	titles, err := dat.ReadBookTitles(testdata.CharDir())
	if err != nil {
		t.Fatal(err)
	}
	if titles[0] != "Macros01" {
		t.Errorf("book 1 = %q, want Macros01", titles[0])
	}
	if titles[32] != "Book33" {
		t.Errorf("book 33 = %q, want Book33", titles[32])
	}
	if titles[39] != "Book40" {
		t.Errorf("book 40 = %q, want Book40", titles[39])
	}
}

func TestReadMacroSet_InvalidSize(t *testing.T) {
	_, err := dat.ReadMacroSet([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for short file")
	}
}

func TestDiscoverMacroFiles(t *testing.T) {
	files, err := dat.DiscoverMacroFiles(testdata.CharDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 12 {
		t.Fatalf("got %d files, want 12", len(files))
	}
}

func TestReadMacroSetFile(t *testing.T) {
	set, err := dat.ReadMacroSetFile(filepath.Join(testdata.CharDir(), "mcr320.dat"))
	if err != nil {
		t.Fatal(err)
	}
	if set.Ctrl[0].Name != "B33S1" {
		t.Errorf("name = %q", set.Ctrl[0].Name)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}