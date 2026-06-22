package dat

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
)

func TestEncodeText_Empty(t *testing.T) {
	out := EncodeText("", 10)
	if len(out) != 10 {
		t.Fatalf("len = %d, want 10", len(out))
	}
	for _, b := range out {
		if b != 0 {
			t.Fatal("expected all-zero output for empty string")
		}
	}
}

func TestEncodeText_ASCII(t *testing.T) {
	s := "/ma \"Cure IV\" <me>"
	out := EncodeText(s, LineSize) // 61 bytes
	if len(out) != LineSize {
		t.Fatalf("len = %d, want %d", len(out), LineSize)
	}
	// bytes 0..len(s)-1 should match ASCII
	for i, b := range []byte(s) {
		if out[i] != b {
			t.Fatalf("out[%d] = 0x%02X, want 0x%02X", i, out[i], b)
		}
	}
	// last byte must be NUL
	if out[LineSize-1] != 0 {
		t.Error("last byte must be NUL")
	}
}

func TestEncodeText_NULTerminator(t *testing.T) {
	// A string that exactly fills limit (size-1) bytes should still have NUL at end
	s := string(bytes.Repeat([]byte{'x'}, 60)) // 60 chars, limit = 60
	out := EncodeText(s, LineSize)
	if out[LineSize-1] != 0 {
		t.Error("last byte must be NUL even when content fills limit")
	}
	for i := 0; i < 60; i++ {
		if out[i] != 'x' {
			t.Fatalf("out[%d] = 0x%02X, want 'x'", i, out[i])
		}
	}
}

func TestEncodeText_ResourceMarker(t *testing.T) {
	s := "≺[07021203]≻"
	out := EncodeText(s, LineSize)
	want := []byte{0xFD, 0x07, 0x02, 0x12, 0x03, 0xFD}
	if !bytes.HasPrefix(out, want) {
		t.Errorf("got prefix %v, want %v", out[:6], want)
	}
}

func TestEncodeText_ElementalMarker(t *testing.T) {
	for n := 0; n <= 7; n++ {
		s := "≺element:" + string(rune('0'+n)) + "≻"
		out := EncodeText(s, LineSize)
		if out[0] != 0xEF || out[1] != byte(0x1F+n) {
			t.Errorf("element:%d: got [0x%02X 0x%02X], want [0xEF 0x%02X]", n, out[0], out[1], 0x1F+n)
		}
	}
}

func TestEncodeText_AutoTranslate(t *testing.T) {
	cases := []struct {
		token string
		b1    byte
	}{
		{"≺autotrans:start≻", 0x27},
		{"≺autotrans:end≻", 0x28},
	}
	for _, c := range cases {
		out := EncodeText(c.token, LineSize)
		if out[0] != 0xEF || out[1] != c.b1 {
			t.Errorf("%s: got [0x%02X 0x%02X], want [0xEF 0x%02X]", c.token, out[0], out[1], c.b1)
		}
	}
}

func TestEncodeText_ByteMarker(t *testing.T) {
	// Single byte marker
	out1 := EncodeText("≺byte:EF≻", LineSize)
	if out1[0] != 0xEF {
		t.Errorf("single byte marker: got 0x%02X, want 0xEF", out1[0])
	}
	// Double byte marker
	out2 := EncodeText("≺byte:EF1F≻", LineSize)
	if out2[0] != 0xEF || out2[1] != 0x1F {
		t.Errorf("double byte marker: got [0x%02X 0x%02X], want [0xEF 0x1F]", out2[0], out2[1])
	}
}

func TestEncodeText_HiraganaRoundTrip(t *testing.T) {
	// ひ is U+3072 (base = 0x3072 - 0x3041 = 49)
	s := "ひ"
	out := EncodeText(s, 10)
	decoded := DecodeText(out)
	if decoded != s {
		t.Errorf("hiragana round-trip: got %q, want %q", decoded, s)
	}
}

func TestEncodeText_KatakanaRoundTrip(t *testing.T) {
	// ア is U+30A2 (base = 1)
	s := "ア"
	out := EncodeText(s, 10)
	decoded := DecodeText(out)
	if decoded != s {
		t.Errorf("katakana round-trip: got %q, want %q", decoded, s)
	}
}

func TestEncodeMacroSet_Size(t *testing.T) {
	var set MacroSet
	out := EncodeMacroSet(set)
	if len(out) != MacroSetFileSize {
		t.Errorf("EncodeMacroSet len = %d, want %d", len(out), MacroSetFileSize)
	}
}

func TestEncodeMacroSet_Header(t *testing.T) {
	var set MacroSet
	out := EncodeMacroSet(set)

	magic := binary.LittleEndian.Uint32(out[0:4])
	if magic != MagicVersion {
		t.Errorf("magic = %d, want %d", magic, MagicVersion)
	}

	// bytes 4–7: unknown field, written as 0 when MacroSet.HeaderUnknown is zero
	flag := binary.LittleEndian.Uint32(out[4:8])
	if flag != 0 {
		t.Errorf("header unknown bytes = 0x%08X, want 0x00000000", flag)
	}

	// bytes 8–23: MD5 of the payload (bytes 24+)
	wantMD5 := md5.Sum(out[HeaderSize:])
	var gotMD5 [16]byte
	copy(gotMD5[:], out[8:24])
	if gotMD5 != wantMD5 {
		t.Errorf("header MD5 = %x, want %x", gotMD5, wantMD5)
	}
}

func TestEncodeMacroSet_RoundTrip(t *testing.T) {
	original := MacroSet{}
	original.Ctrl[0] = Macro{Name: "Cure", Contents: [LineCount]string{"/ma \"Cure IV\" <me>", "/wait 1"}}
	original.Alt[9] = Macro{Name: "WS", Contents: [LineCount]string{"/ws \"Savage Blade\" <t>"}}

	encoded := EncodeMacroSet(original)
	decoded, err := ReadMacroSet(encoded)
	if err != nil {
		t.Fatalf("ReadMacroSet: %v", err)
	}

	if decoded.Ctrl[0].Name != original.Ctrl[0].Name {
		t.Errorf("ctrl[0].Name = %q, want %q", decoded.Ctrl[0].Name, original.Ctrl[0].Name)
	}
	if decoded.Ctrl[0].Contents[0] != original.Ctrl[0].Contents[0] {
		t.Errorf("ctrl[0].Contents[0] = %q, want %q", decoded.Ctrl[0].Contents[0], original.Ctrl[0].Contents[0])
	}
	if decoded.Alt[9].Name != original.Alt[9].Name {
		t.Errorf("alt[9].Name = %q, want %q", decoded.Alt[9].Name, original.Alt[9].Name)
	}
}

func TestEncodeMacroSet_HeaderUnknownPreserved(t *testing.T) {
	set := MacroSet{HeaderUnknown: 0xDEADBEEF}
	set.Ctrl[0] = Macro{Name: "Test"}

	encoded := EncodeMacroSet(set)

	got := binary.LittleEndian.Uint32(encoded[4:8])
	if got != 0xDEADBEEF {
		t.Errorf("header unknown bytes = 0x%08X, want 0xDEADBEEF", got)
	}

	decoded, err := ReadMacroSet(encoded)
	if err != nil {
		t.Fatalf("ReadMacroSet: %v", err)
	}
	if decoded.HeaderUnknown != 0xDEADBEEF {
		t.Errorf("decoded HeaderUnknown = 0x%08X, want 0xDEADBEEF", decoded.HeaderUnknown)
	}
}

func TestEncodeMacroSet_FromTestdata(t *testing.T) {
	path := filepath.Join(testdata.CharDir(), "mcr320.dat")
	original, err := ReadMacroSetFile(path)
	if err != nil {
		t.Fatalf("ReadMacroSetFile: %v", err)
	}

	encoded := EncodeMacroSet(original)
	roundtripped, err := ReadMacroSet(encoded)
	if err != nil {
		t.Fatalf("ReadMacroSet after encode: %v", err)
	}

	for i := 0; i < 10; i++ {
		if roundtripped.Ctrl[i] != original.Ctrl[i] {
			t.Errorf("Ctrl[%d] differs after round-trip", i)
		}
		if roundtripped.Alt[i] != original.Alt[i] {
			t.Errorf("Alt[%d] differs after round-trip", i)
		}
	}
}

func TestWriteMacroSetFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mcr.dat")

	set := MacroSet{}
	set.Ctrl[0] = Macro{Name: "Test", Contents: [LineCount]string{"/echo hello"}}

	if err := WriteMacroSetFile(path, set); err != nil {
		t.Fatalf("WriteMacroSetFile: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Size() != MacroSetFileSize {
		t.Errorf("file size = %d, want %d", fi.Size(), MacroSetFileSize)
	}

	read, err := ReadMacroSetFile(path)
	if err != nil {
		t.Fatalf("ReadMacroSetFile: %v", err)
	}
	if read.Ctrl[0].Name != "Test" {
		t.Errorf("Ctrl[0].Name = %q, want \"Test\"", read.Ctrl[0].Name)
	}
}

func TestWriteBookTitles(t *testing.T) {
	dir := t.TempDir()

	var titles [MaxBooks]string
	titles[0] = "WHM75"
	titles[19] = "RDM75NIN"
	titles[20] = "BLM90"

	if err := WriteBookTitles(dir, titles); err != nil {
		t.Fatalf("WriteBookTitles: %v", err)
	}

	read, err := ReadBookTitles(dir)
	if err != nil {
		t.Fatalf("ReadBookTitles: %v", err)
	}

	cases := []struct{ idx int; want string }{
		{0, "WHM75"},
		{19, "RDM75NIN"},
		{20, "BLM90"},
		{1, ""},
	}
	for _, c := range cases {
		if read[c.idx] != c.want {
			t.Errorf("titles[%d] = %q, want %q", c.idx, read[c.idx], c.want)
		}
	}
}
