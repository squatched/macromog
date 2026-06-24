package dat_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
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
	if !strings.Contains(line2, "[07021203]") {
		t.Errorf("line2 missing Cure III resource marker: %q", line2)
	}

	line3 := set.Ctrl[9].Contents[2]
	if !strings.Contains(line3, "[02020114]") {
		t.Errorf("line3 missing Good luck marker: %q", line3)
	}
	if !strings.Contains(line3, "Good luck!") {
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

func TestReadMacroSet_BadMagic(t *testing.T) {
	data := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(data[0:4], 99)
	_, err := dat.ReadMacroSet(data)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestReadMacroSet_ShiftJIS(t *testing.T) {
	// Synthetic JP macro: katakana name アイ, hiragana line あい
	data := buildMacroSetFile(t, [20]macroSpec{{
		name:  string([]byte{0x83, 0x41, 0x83, 0x43}), // アイ
		lines: [6]string{string([]byte{0x82, 0xA0, 0x82, 0xA2})}, // あい
	}})
	set, err := dat.ReadMacroSet(data)
	if err != nil {
		t.Fatal(err)
	}
	if set.Ctrl[0].Name != "アイ" {
		t.Errorf("name = %q, want アイ", set.Ctrl[0].Name)
	}
	if set.Ctrl[0].Contents[0] != "あい" {
		t.Errorf("line1 = %q, want あい", set.Ctrl[0].Contents[0])
	}
	if set.Ctrl[0].Empty() {
		t.Error("JP macro should not be empty")
	}
}

func TestReadBookTitles_BadMagic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mcr.ttl")
	if err := os.WriteFile(path, []byte{2, 0, 0, 0}, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := dat.ReadBookTitles(dir)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestReadBookTitles_PayloadSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mcr.ttl")
	// 24-byte header + 17-byte payload (not a multiple of 16)
	data := make([]byte, 41)
	binary.LittleEndian.PutUint32(data[0:4], dat.MagicVersion)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := dat.ReadBookTitles(dir)
	if err == nil {
		t.Fatal("expected error for misaligned payload")
	}
}

func TestMacro_Empty(t *testing.T) {
	var m dat.Macro
	if !m.Empty() {
		t.Error("zero macro should be empty")
	}
	m.Name = "x"
	if m.Empty() {
		t.Error("named macro should not be empty")
	}
	m = dat.Macro{Contents: [6]string{"", "line"}}
	if m.Empty() {
		t.Error("macro with line content should not be empty")
	}
}

type macroSpec struct {
	name  string
	lines [6]string
}

func buildMacroSetFile(t *testing.T, macros [20]macroSpec) []byte {
	t.Helper()
	data := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(data[0:4], dat.MagicVersion)
	offset := dat.HeaderSize
	for i := 0; i < dat.MacrosPerSet; i++ {
		writeMacro(data[offset:], macros[i])
		offset += dat.MacroSize
	}
	return data
}

func writeMacro(buf []byte, spec macroSpec) {
	pos := dat.MacroPrefixSize
	for i := 0; i < dat.LineCount; i++ {
		line := spec.lines[i]
		copy(buf[pos:], line)
		pos += dat.LineSize
	}
	copy(buf[pos:], spec.name)
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

func TestDiscoverMacroFiles_NumericOrder(t *testing.T) {
	dir := t.TempDir()
	blank := make([]byte, dat.MacroSetFileSize)
	binary.LittleEndian.PutUint32(blank[0:4], dat.MagicVersion)
	for _, name := range []string{"mcr10.dat", "mcr1.dat", "mcr2.dat", "mcr400.dat"} {
		if err := os.WriteFile(filepath.Join(dir, name), blank, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := dat.DiscoverMacroFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3 (mcr400.dat ignored)", len(files))
	}
	want := []string{"mcr1.dat", "mcr2.dat", "mcr10.dat"}
	for i, name := range want {
		if filepath.Base(files[i]) != name {
			t.Errorf("files[%d] = %q, want %q", i, filepath.Base(files[i]), name)
		}
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
