package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

func TestRunConfig_Help(t *testing.T) {
	if got := runConfig([]string{"--help"}, newTextPrinter()); got != 0 {
		t.Errorf("runConfig(--help) = %d, want 0", got)
	}
}

func TestRunConfig_UnknownSubcommand(t *testing.T) {
	if got := runConfig([]string{"nope"}, newTextPrinter()); got != 1 {
		t.Errorf("runConfig(unknown) = %d, want 1", got)
	}
}

func TestRunConfig_Path(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	if got := runConfig([]string{"path"}, p); got != 0 {
		t.Fatalf("runConfig(path) = %d, want 0", got)
	}
	if !strings.Contains(buf.String(), "config.yml") {
		t.Errorf("path output missing config.yml: %q", buf.String())
	}
}

func TestRunConfig_Show(t *testing.T) {
	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatJSON)
	if got := runConfig([]string{"show"}, p); got != 0 {
		t.Fatalf("runConfig(show) = %d, want 0", got)
	}
	if !strings.Contains(buf.String(), `"version"`) {
		t.Errorf("JSON show missing version: %s", buf.String())
	}
}

func resetTestConfig(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := config.Save(path, config.Empty()); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)
}

func TestRunConfig_AddInstall(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	if got := runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter()); got != 0 {
		t.Fatalf("add-install = %d, want 0", got)
	}
	cfg, err := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultInstall != "steam" {
		t.Errorf("default_install = %q, want steam", cfg.DefaultInstall)
	}
}

func TestRunConfig_AddInstall_Errors(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func(t *testing.T) string
	}{
		{
			name: "missing args",
			args: []string{"add-install", "steam"},
		},
		{
			name: "no USER dir",
			args: []string{"add-install", "steam"},
			setup: func(t *testing.T) string {
				resetTestConfig(t)
				return t.TempDir()
			},
		},
		{
			name: "duplicate install",
			args: []string{"add-install", "steam"},
			setup: func(t *testing.T) string {
				resetTestConfig(t)
				ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
				if got := runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter()); got != 0 {
					t.Fatalf("setup add-install = %d", got)
				}
				return ffxiDir
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ffxiDir string
			if tc.setup != nil {
				ffxiDir = tc.setup(t)
			} else {
				resetTestConfig(t)
			}
			args := tc.args
			if ffxiDir != "" {
				args = append(append([]string{}, tc.args...), ffxiDir)
			}
			if got := runConfig(args, newTextPrinter()); got != 1 {
				t.Errorf("runConfig(%v) = %d, want 1", args, got)
			}
		})
	}
}

func TestRunConfig_AddInstall_SecondWithSetDefault(t *testing.T) {
	resetTestConfig(t)
	ffxiA, _, _ := makeFFXITree(t, "a1b2c3d4")
	ffxiB, _, _ := makeFFXITree(t, "e5f6a7b8")

	if got := runConfig([]string{"add-install", "steam", ffxiA}, newTextPrinter()); got != 0 {
		t.Fatalf("first add-install = %d", got)
	}
	if got := runConfig([]string{"add-install", "lutris", "--set-default", ffxiB}, newTextPrinter()); got != 0 {
		t.Fatalf("second add-install = %d", got)
	}

	cfg, err := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultInstall != "lutris" {
		t.Errorf("default_install = %q, want lutris", cfg.DefaultInstall)
	}
}

func TestRunConfig_SetAlias(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	if got := runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter()); got != 0 {
		t.Fatalf("add-install = %d, want 0", got)
	}
	if got := runConfig([]string{"set-alias", "a1b2c3d4", "Squatched"}, newTextPrinter()); got != 0 {
		t.Fatalf("set-alias = %d, want 0", got)
	}
	cfg, err := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Installs["steam"].Characters["a1b2c3d4"].Name != "Squatched" {
		t.Errorf("alias not saved")
	}
	if cfg.Installs["steam"].DefaultCharacter != "a1b2c3d4" {
		t.Errorf("first alias should become default character")
	}
}

func TestRunConfig_SetAlias_SecondDoesNotChangeDefault(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter())
	runConfig([]string{"set-alias", "a1b2c3d4", "Squatched"}, newTextPrinter())

	if got := runConfig([]string{"set-alias", "e5f6a7b8", "AltMule"}, newTextPrinter()); got != 0 {
		t.Fatalf("second set-alias = %d", got)
	}
	cfg, _ := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if cfg.Installs["steam"].DefaultCharacter != "a1b2c3d4" {
		t.Errorf("default_character = %q, want a1b2c3d4", cfg.Installs["steam"].DefaultCharacter)
	}
}

func TestRunConfig_SetAlias_Errors(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter())

	tests := []struct {
		name string
		args []string
	}{
		{name: "missing args", args: []string{"set-alias", "a1b2c3d4"}},
		{name: "empty name", args: []string{"set-alias", "a1b2c3d4", "   "}},
		{name: "invalid char id", args: []string{"set-alias", "ghost", "Nobody"}},
		{name: "no install configured", args: []string{"set-alias", "a1b2c3d4", "Squatched", "--install", "lutris"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := runConfig(tc.args, newTextPrinter()); got != 1 {
				t.Errorf("runConfig(%v) = %d, want 1", tc.args, got)
			}
		})
	}
}

func TestRunConfig_RemoveAlias_ClearsDefaultCharacter(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter())
	runConfig([]string{"set-alias", "a1b2c3d4", "Squatched"}, newTextPrinter())

	if got := runConfig([]string{"remove-alias", "a1b2c3d4"}, newTextPrinter()); got != 0 {
		t.Fatalf("remove-alias = %d", got)
	}
	cfg, _ := config.Load(os.Getenv("MACROMOG_CONFIG"))
	inst := cfg.Installs["steam"]
	if inst.DefaultCharacter != "" {
		t.Errorf("default_character = %q, want empty", inst.DefaultCharacter)
	}
	if inst.Characters != nil {
		t.Errorf("characters = %v, want nil", inst.Characters)
	}
}

func TestRunConfig_SetAndRemoveDefaultCharacter(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter())
	runConfig([]string{"set-alias", "a1b2c3d4", "Squatched"}, newTextPrinter())
	runConfig([]string{"set-alias", "e5f6a7b8", "AltMule"}, newTextPrinter())

	if got := runConfig([]string{"set-default-character", "e5f6a7b8"}, newTextPrinter()); got != 0 {
		t.Fatalf("set-default-character = %d", got)
	}
	cfg, _ := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if cfg.Installs["steam"].DefaultCharacter != "e5f6a7b8" {
		t.Errorf("default_character = %q, want e5f6a7b8", cfg.Installs["steam"].DefaultCharacter)
	}

	if got := runConfig([]string{"remove-default-character"}, newTextPrinter()); got != 0 {
		t.Fatalf("remove-default-character = %d", got)
	}
	cfg, _ = config.Load(os.Getenv("MACROMOG_CONFIG"))
	if cfg.Installs["steam"].DefaultCharacter != "" {
		t.Errorf("default_character = %q, want empty", cfg.Installs["steam"].DefaultCharacter)
	}
}

func TestRunConfig_SetDefaultInstall(t *testing.T) {
	resetTestConfig(t)
	ffxiA, _, _ := makeFFXITree(t, "a1b2c3d4")
	ffxiB, _, _ := makeFFXITree(t, "e5f6a7b8")
	runConfig([]string{"add-install", "steam", ffxiA}, newTextPrinter())
	runConfig([]string{"add-install", "lutris", ffxiB}, newTextPrinter())

	if got := runConfig([]string{"set-default-install", "lutris"}, newTextPrinter()); got != 0 {
		t.Fatalf("set-default-install = %d", got)
	}
	cfg, _ := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if cfg.DefaultInstall != "lutris" {
		t.Errorf("default_install = %q, want lutris", cfg.DefaultInstall)
	}
}

func TestRunConfig_RemoveDefaultInstall(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter())

	if got := runConfig([]string{"remove-default-install"}, newTextPrinter()); got != 0 {
		t.Fatalf("remove-default-install = %d", got)
	}
	cfg, _ := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if cfg.DefaultInstall != "" {
		t.Errorf("default_install = %q, want empty", cfg.DefaultInstall)
	}
}

func TestRunConfig_DefaultOffering(t *testing.T) {
	resetTestConfig(t)
	for _, args := range [][]string{
		{"default-offering", "false"},
		{"default-offering", "true"},
	} {
		if got := runConfig(args, newTextPrinter()); got != 0 {
			t.Fatalf("runConfig(%v) = %d", args, got)
		}
	}
	cfg, err := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	if !config.DefaultOffering(&cfg) {
		t.Error("expected default offering re-enabled")
	}
}

func TestRunConfig_DefaultOffering_Invalid(t *testing.T) {
	resetTestConfig(t)
	if got := runConfig([]string{"default-offering", "maybe"}, newTextPrinter()); got != 1 {
		t.Errorf("default-offering invalid = %d, want 1", got)
	}
}

func TestRunConfig_RemoveInstallClearsDefault(t *testing.T) {
	resetTestConfig(t)
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	runConfig([]string{"add-install", "steam", ffxiDir}, newTextPrinter())
	if got := runConfig([]string{"remove-install", "steam"}, newTextPrinter()); got != 0 {
		t.Fatalf("remove-install = %d, want 0", got)
	}
	cfg, err := config.Load(os.Getenv("MACROMOG_CONFIG"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultInstall != "" {
		t.Errorf("default_install should be cleared, got %q", cfg.DefaultInstall)
	}
	if cfg.Installs != nil {
		t.Errorf("installs should be nil, got %v", cfg.Installs)
	}
}

func TestRunConfig_ShowText(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	path := filepath.Join(t.TempDir(), "config.yml")
	norm, err := config.NormalizePath(ffxiDir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Version: 1,
		Installs: map[string]config.Install{
			"steam": {Path: norm},
		},
	}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)

	var buf bytes.Buffer
	p := NewPrinter(&buf, FormatText)
	if got := runConfig([]string{"show"}, p); got != 0 {
		t.Fatalf("show = %d, want 0", got)
	}
	if !strings.Contains(buf.String(), "steam") {
		t.Errorf("show output missing steam: %s", buf.String())
	}
}
