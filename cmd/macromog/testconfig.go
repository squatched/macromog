package main

import (
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/config"
)

func setTestConfig(t *testing.T, ffxiPath string, aliases map[string]string) string {
	t.Helper()
	norm, err := config.NormalizePath(ffxiPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Version:        1,
		DefaultInstall: "test",
		Installs: map[string]config.Install{
			"test": {Path: norm},
		},
	}
	if len(aliases) > 0 {
		chars := make(map[string]config.Character, len(aliases))
		for id, name := range aliases {
			chars[id] = config.Character{Name: name}
		}
		inst := cfg.Installs["test"]
		inst.Characters = chars
		cfg.Installs["test"] = inst
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)
	return path
}
