package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

func TestOpenConfig_CreatesOnFirstUse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	t.Setenv("MACROMOG_CONFIG", path)

	session, err := openConfig()
	if err != nil {
		t.Fatal(err)
	}
	if session.path != path {
		t.Errorf("path = %q, want %q", session.path, path)
	}
	if session.cfg.Version != 1 {
		t.Errorf("version = %d, want 1", session.cfg.Version)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "version: 1") {
		t.Errorf("file not created: %s", data)
	}
}

func TestOpenConfig_InvalidNonInteractive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("version: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)

	_, err := openConfig()
	if err == nil {
		t.Fatal("expected invalid config error")
	}
	if !strings.Contains(err.Error(), "config.yml is invalid") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenConfig_LoadsValidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	ffxi := filepath.Join(t.TempDir(), "ffxi")
	norm, err := config.NormalizePath(ffxi)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Version:        1,
		DefaultInstall: "steam",
		Installs: map[string]config.Install{
			"steam": {Path: norm},
		},
	}
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)

	session, err := openConfig()
	if err != nil {
		t.Fatal(err)
	}
	if session.cfg.DefaultInstall != "steam" {
		t.Errorf("default_install = %q, want steam", session.cfg.DefaultInstall)
	}
}
