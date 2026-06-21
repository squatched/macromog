package dat

import (
	"crypto/md5"
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/dat/testdata"
)

// TestEncodeMacroSet_PayloadBytes verifies that re-encoding a decoded MacroSet
// produces bytes that are identical to the original in every content field.
// Specifically it checks:
//   - The MD5 checksum at bytes 8–23 (must match the original)
//   - For each of the 20 macros: the 4-byte prefix (always 0)
//   - Each of the 6 line fields through the NUL terminator (content + NUL)
//   - The name field through the NUL terminator
//
// Bytes 4–7 (unknown flag) are excluded: originals may carry 0x80000000.
func TestEncodeMacroSet_PayloadBytes(t *testing.T) {
	entries, err := os.ReadDir(testdata.CharDir())
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	tested := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if _, ok := ParseMacroFileName(e.Name()); !ok {
			continue
		}
		path := filepath.Join(testdata.CharDir(), e.Name())
		original, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}

		ms, err := ReadMacroSet(original)
		if err != nil {
			t.Fatalf("decode %s: %v", e.Name(), err)
		}

		reencoded := EncodeMacroSet(ms)

		// Verify the re-encoded MD5 matches the original's stored checksum.
		// (Both should equal MD5 of bytes 24+.)
		wantMD5 := md5.Sum(original[HeaderSize:])
		gotMD5 := md5.Sum(reencoded[HeaderSize:])
		if wantMD5 != gotMD5 {
			t.Errorf("%s: MD5 mismatch: orig payload hash %x, re-encoded payload hash %x",
				e.Name(), wantMD5, gotMD5)
		}
		// The stored checksum in the re-encoded file must also be correct.
		var storedMD5 [16]byte
		copy(storedMD5[:], reencoded[8:24])
		if storedMD5 != gotMD5 {
			t.Errorf("%s: stored MD5 in header %x does not match payload hash %x",
				e.Name(), storedMD5, gotMD5)
		}

		compareMacroPayloads(t, e.Name(), original, reencoded)
		tested++
	}
	if tested == 0 {
		t.Fatal("no .dat files found in testdata")
	}
	t.Logf("compared payload bytes across %d .dat file(s)", tested)
}

// compareMacroPayloads compares each content field in the two byte slices.
// Both must be MacroSetFileSize bytes. The check covers:
//   - 4-byte prefix per macro (always 0)
//   - 6 × LineSize bytes per macro: bytes [0, nulPos] where nulPos is
//     the NUL position in the *original* field (content + terminator)
//   - NameSize bytes per macro: same approach
func compareMacroPayloads(t *testing.T, name string, orig, got []byte) {
	t.Helper()
	if len(orig) != MacroSetFileSize || len(got) != MacroSetFileSize {
		t.Errorf("%s: unexpected sizes orig=%d got=%d", name, len(orig), len(got))
		return
	}

	off := HeaderSize
	for i := 0; i < MacrosPerSet; i++ {
		// 4-byte prefix — must be zero in both
		for j := 0; j < MacroPrefixSize; j++ {
			if orig[off+j] != 0 || got[off+j] != 0 {
				t.Errorf("%s macro %d prefix[%d]: orig=0x%02X got=0x%02X (both should be 0)",
					name, i, j, orig[off+j], got[off+j])
			}
		}
		off += MacroPrefixSize

		// 6 line fields
		for ln := 0; ln < LineCount; ln++ {
			origField := orig[off : off+LineSize]
			gotField := got[off : off+LineSize]
			compareField(t, name, i, "line", ln, origField, gotField)
			off += LineSize
		}

		// name field
		origField := orig[off : off+NameSize]
		gotField := got[off : off+NameSize]
		compareField(t, name, i, "name", -1, origField, gotField)
		off += NameSize
	}
}

// compareField checks that the two fields agree on every byte up to and
// including the NUL terminator position found in orig.
func compareField(t *testing.T, file string, macro int, kind string, idx int, orig, got []byte) {
	t.Helper()
	nulPos := len(orig)
	for i, b := range orig {
		if b == 0 {
			nulPos = i
			break
		}
	}
	// compare content bytes + NUL
	for i := 0; i <= nulPos && i < len(orig); i++ {
		if orig[i] != got[i] {
			label := kind
			if idx >= 0 {
				label = "line[" + string(rune('0'+idx)) + "]"
			}
			t.Errorf("%s macro %d %s byte[%d]: orig=0x%02X got=0x%02X",
				file, macro, label, i, orig[i], got[i])
		}
	}
}
