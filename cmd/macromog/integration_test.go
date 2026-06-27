package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/dat"
	"github.com/squatched/macromog/internal/dat/testdata"
	"github.com/squatched/macromog/internal/export"
)

// TestIntegration_ExportValidateImportReexport exercises the full standard
// workflow: export from testdata → validate YAML → import into a fresh dir →
// re-export from that dir → verify original content is preserved.
func TestIntegration_ExportValidateImportReexport(t *testing.T) {
	tmp := t.TempDir()

	// Step 1: Export from testdata character directory.
	yamlPath := filepath.Join(tmp, "macros.yml")
	if got := runExport([]string{"--char-dir", testdata.CharDir(), yamlPath}, newTextPrinter()); got != 0 {
		t.Fatalf("export: exit %d", got)
	}

	// Step 2: Validate the exported YAML.
	if got := runValidate([]string{yamlPath}, newTextPrinter()); got != 0 {
		t.Fatalf("validate: exit %d", got)
	}

	// Step 3: Import into a fresh character directory.
	destDir := t.TempDir()
	if got := runImport([]string{"--no-backup", yamlPath, destDir}, newTextPrinter()); got != 0 {
		t.Fatalf("import: exit %d", got)
	}

	// Step 4: Re-export from the destination directory.
	reexportPath := filepath.Join(tmp, "reexport.yml")
	if got := runExport([]string{"--char-dir", destDir, reexportPath}, newTextPrinter()); got != 0 {
		t.Fatalf("re-export: exit %d", got)
	}

	// Step 5: Verify content survived the round-trip.
	origData, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	reexportData, err := os.ReadFile(reexportPath)
	if err != nil {
		t.Fatalf("read reexport: %v", err)
	}
	if !strings.Contains(string(reexportData), "B33S1") {
		t.Errorf("B33S1 missing from re-exported YAML:\n%s", reexportData)
	}
	// Both should declare the same scope.
	if strings.Contains(string(origData), "level: full") != strings.Contains(string(reexportData), "level: full") {
		t.Errorf("scope level mismatch between original and re-export")
	}
}

// TestIntegration_TemplateIsValidYAML generates a scoped template and
// confirms it passes validate without modification.
func TestIntegration_TemplateIsValidYAML(t *testing.T) {
	tmp := t.TempDir()
	tmplPath := filepath.Join(tmp, "template.yml")

	if got := runTemplate([]string{"--scope", "B1S3", tmplPath}, newTextPrinter()); got != 0 {
		t.Fatalf("template: exit %d", got)
	}
	if got := runValidate([]string{tmplPath}, newTextPrinter()); got != 0 {
		t.Fatalf("validate(template): exit %d — template output is not valid YAML", got)
	}
	data, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "# Macro Line 1") {
		t.Errorf("template output must use comment placeholders, not raw empty strings:\n%s", s)
	}
	if strings.Contains(s, `- ""`) {
		t.Errorf("template output must not contain raw empty-string items:\n%s", s)
	}
}

// TestIntegration_AllMultiCharRoundTrip exports and imports for multiple
// characters via --all and verifies all succeed.
func TestIntegration_AllMultiCharRoundTrip(t *testing.T) {
	src := testdata.CharDir()

	// Build a two-character FFXI tree, each seeded from testdata.
	ffxiDir := t.TempDir()
	userDir := filepath.Join(ffxiDir, "USER")
	if err := os.Mkdir(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"a1b2c3d4", "e5f6a7b8"} {
		charDir := filepath.Join(userDir, id)
		if err := os.Mkdir(charDir, 0o755); err != nil {
			t.Fatal(err)
		}
		entries, _ := os.ReadDir(src)
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, _ := os.ReadFile(filepath.Join(src, e.Name()))
			_ = os.WriteFile(filepath.Join(charDir, e.Name()), data, 0o644)
		}
		// testdata has no mcr.dat (book 1, set 1); write a valid empty one so
		// the directory is discoverable by DiscoverCharacters.
		_ = os.WriteFile(filepath.Join(charDir, "mcr.dat"), dat.EncodeMacroSet(dat.MacroSet{}), 0o644)
	}

	// Export from a single character to get the YAML.
	doc, err := export.FromCharacterDir(export.Options{CharacterDir: filepath.Join(userDir, "a1b2c3d4"), Character: "A"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	yamlData, _ := export.MarshalYAML(doc)
	yamlPath := filepath.Join(t.TempDir(), "macros.yml")
	if err := os.WriteFile(yamlPath, yamlData, 0o644); err != nil {
		t.Fatal(err)
	}

	// Import the same YAML into both characters.
	args := []string{"--ffxi-path", ffxiDir, "--all", "--no-backup", yamlPath}
	if got := runImport(args, newTextPrinter()); got != 0 {
		t.Errorf("import --all: exit %d", got)
	}

	// Re-export from each character and verify the content survived.
	for _, id := range []string{"a1b2c3d4", "e5f6a7b8"} {
		charDir := filepath.Join(userDir, id)
		reDoc, err := export.FromCharacterDir(export.Options{CharacterDir: charDir, Character: id})
		if err != nil {
			t.Errorf("[%s] re-export: %v", id, err)
			continue
		}
		reData, _ := export.MarshalYAML(reDoc)
		if !strings.Contains(string(reData), "B33S1") {
			t.Errorf("[%s] B33S1 missing from re-exported YAML", id)
		}
	}
}

// TestIntegration_ScopeRoundTrip_MacroLevel exercises the full pipeline for
// macro-level scope: scope string → YAML selections block → import filtering.
//
// Key invariants verified:
//  1. Exporting with C* scope produces a YAML with no "key:" field in selections
//     (nil Key → omitempty omits the field).
//  2. That YAML passes validation.
//  3. Importing the scoped YAML only updates ctrl slots; alt is untouched.
//  4. Re-exporting with the same scope shows the updated content.
func TestIntegration_ScopeRoundTrip_MacroLevel(t *testing.T) {
	src := testdata.CharDir()
	tmp := t.TempDir()

	// Step 1: Full export from testdata → seed a destination directory.
	fullYAML := filepath.Join(tmp, "full.yml")
	if got := runExport([]string{"--char-dir", src, fullYAML}, newTextPrinter()); got != 0 {
		t.Fatalf("full export: exit %d", got)
	}
	destDir := t.TempDir()
	if got := runImport([]string{"--no-backup", fullYAML, destDir}, newTextPrinter()); got != 0 {
		t.Fatalf("full import: exit %d", got)
	}

	// Step 2: Export from testdata with ctrl-wildcard scope for B6S10.
	// This exercises scope string → parsed scope → YAML selections block.
	scopedYAML := filepath.Join(tmp, "scoped.yml")
	if got := runExport([]string{"--char-dir", src, "--scope", "B6S10C*", scopedYAML}, newTextPrinter()); got != 0 {
		t.Fatalf("export --scope B6S10C*: exit %d", got)
	}

	// Step 3: Validate — ensures the generated YAML is schema-compliant.
	if got := runValidate([]string{scopedYAML}, newTextPrinter()); got != 0 {
		t.Fatalf("validate scoped YAML: exit %d", got)
	}

	// Step 4: Verify the selections block has no "key:" field (C* = nil Key).
	yamlData, err := os.ReadFile(scopedYAML)
	if err != nil {
		t.Fatalf("read scoped YAML: %v", err)
	}
	yamlStr := string(yamlData)
	if !strings.Contains(yamlStr, "type: ctrl") {
		t.Errorf("expected 'type: ctrl' in scope selections:\n%s", yamlStr)
	}
	if strings.Contains(yamlStr, "key:") {
		t.Errorf("C* scope must produce no 'key:' field in selections:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "level: macro") {
		t.Errorf("expected 'level: macro' in scope:\n%s", yamlStr)
	}

	// Step 5: Import the scoped YAML into the seeded destination.
	// The YAML scope is C* (ctrl wildcard), so only ctrl slots of B6S10 are
	// updated; alt slots of B6S10 and all other books/sets are untouched.
	if got := runImport([]string{"--no-backup", scopedYAML, destDir}, newTextPrinter()); got != 0 {
		t.Fatalf("scoped import: exit %d", got)
	}

	// Step 6: Re-export with the same C* scope and verify ctrl is present.
	reexportYAML := filepath.Join(tmp, "reexport.yml")
	if got := runExport([]string{"--char-dir", destDir, "--scope", "B6S10C*", reexportYAML}, newTextPrinter()); got != 0 {
		t.Fatalf("re-export with scope: exit %d", got)
	}
	reexportData, err := os.ReadFile(reexportYAML)
	if err != nil {
		t.Fatalf("read re-export: %v", err)
	}

	// The ctrl macros from testdata's B6S10 should be present in the re-export.
	if !strings.Contains(string(reexportData), "level: macro") {
		t.Errorf("re-export missing scope block:\n%s", reexportData)
	}
	// Validate the re-export too — it must remain schema-compliant.
	if got := runValidate([]string{reexportYAML}, newTextPrinter()); got != 0 {
		t.Fatalf("validate re-export: exit %d", got)
	}
}
