package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "macromog-config-*")
	if err != nil {
		os.Exit(1)
	}
	path := filepath.Join(dir, "config.yml")
	if err := config.Save(path, config.Empty()); err != nil {
		os.Exit(1)
	}
	os.Setenv("MACROMOG_CONFIG", path)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
