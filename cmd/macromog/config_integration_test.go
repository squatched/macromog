package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
)

// TestIntegration_ConfigWorkflow exercises the documented happy path:
// register install → set alias → list by name → export by name.
func TestIntegration_ConfigWorkflow(t *testing.T) {
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	charID := "a1b2c3d4"
	charDir := filepath.Join(userDir, charID)
	for _, d := range []string{userDir, charDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	src := testdata.CharDir()
	entries, _ := os.ReadDir(src)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(src, e.Name()))
		_ = os.WriteFile(filepath.Join(charDir, e.Name()), data, 0o644)
	}
	_ = os.WriteFile(filepath.Join(charDir, "mcr.dat"), dat.EncodeMacroSet(dat.MacroSet{}), 0o644)

	resetTestConfig(t)

	if got := runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter()); got != 0 {
		t.Fatalf("add-install = %d", got)
	}
	if got := runConfig([]string{"set-alias", charID, "Squatched"}, newTextPrinter()); got != 0 {
		t.Fatalf("set-alias = %d", got)
	}

	var listBuf bytes.Buffer
	if got := runList([]string{"--install", "steam", "--char-name", "Squatched"}, NewPrinter(&listBuf, FormatText)); got != 0 {
		t.Fatalf("list = %d", got)
	}
	if !strings.Contains(listBuf.String(), "Squatched") {
		t.Errorf("list output missing alias:\n%s", listBuf.String())
	}

	out := filepath.Join(t.TempDir(), "macros.yml")
	if got := runExport([]string{"--install", "steam", "--char-name", "Squatched", out}, newTextPrinter()); got != 0 {
		t.Fatalf("export = %d", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "character: Squatched") {
		t.Errorf("export missing character name:\n%s", data)
	}
}
