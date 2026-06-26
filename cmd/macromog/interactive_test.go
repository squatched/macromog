package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

func setInteractiveStdin(t *testing.T, input string) {
	t.Helper()
	interactiveStdin.isTTY = func() bool { return true }
	interactiveStdin.r = strings.NewReader(input)
	interactiveStdin.scanner = nil
	t.Cleanup(restoreInteractiveStdin)
}

func TestMaybeRegisterInstall(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantInstall bool
	}{
		{
			name:        "decline",
			input:       "n\n",
			wantInstall: false,
		},
		{
			name:        "accept default suggested name",
			input:       "\n\n",
			wantName:    "default",
			wantInstall: true,
		},
		{
			name:        "accept with custom name",
			input:       "yes\nmyinstall\n",
			wantName:    "myinstall",
			wantInstall: true,
		},
		{
			name:        "EOF on confirmation",
			input:       "",
			wantInstall: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
			session := testSession(t, config.Empty())
			setInteractiveStdin(t, tc.input)

			ctx, err := maybeRegisterInstall(session, ffxiDir)
			if err != nil {
				t.Fatalf("maybeRegisterInstall: %v", err)
			}
			if tc.wantInstall {
				if ctx.installName != tc.wantName {
					t.Errorf("installName = %q, want %q", ctx.installName, tc.wantName)
				}
				if ctx.install == nil {
					t.Fatal("expected install context")
				}
				cfg, err := config.Load(session.path)
				if err != nil {
					t.Fatal(err)
				}
				if _, ok := cfg.Installs[tc.wantName]; !ok {
					t.Errorf("install %q not saved", tc.wantName)
				}
			} else {
				if ctx.install != nil {
					t.Errorf("expected no install registration, got %q", ctx.installName)
				}
			}
		})
	}
}

func TestMaybeRegisterInstall_CI(t *testing.T) {
	t.Setenv("CI", "1")
	t.Cleanup(restoreInteractiveStdin)

	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	session := testSession(t, config.Empty())

	_, err := maybeRegisterInstall(session, ffxiDir)
	if err == nil {
		t.Fatal("expected error in CI mode, got nil")
	}
	if !strings.Contains(err.Error(), ffxiDir) {
		t.Errorf("error should mention detected path; got: %v", err)
	}
}

func TestMaybeRegisterInstall_SecondDoesNotChangeDefault(t *testing.T) {
	ffxiA, _, _ := makeFFXITree(t, "a1b2c3d4")
	ffxiB, _, _ := makeFFXITree(t, "e5f6a7b8")
	normA, _ := config.NormalizePath(ffxiA)
	cfg := config.Config{
		Version:        1,
		DefaultInstall: "first",
		Installs: map[string]config.Install{
			"first": {Path: normA},
		},
	}
	session := testSession(t, cfg)
	setInteractiveStdin(t, "\nother\n")

	ctx, err := maybeRegisterInstall(session, ffxiB)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.installName != "other" {
		t.Errorf("installName = %q, want other", ctx.installName)
	}
	loaded, _ := config.Load(session.path)
	if loaded.DefaultInstall != "first" {
		t.Errorf("default_install = %q, want first", loaded.DefaultInstall)
	}
}

func TestPromptInstallSelect(t *testing.T) {
	ffxiA, _, _ := makeFFXITree(t, "a1b2c3d4")
	ffxiB, _, _ := makeFFXITree(t, "e5f6a7b8")
	normA, _ := config.NormalizePath(ffxiA)
	normB, _ := config.NormalizePath(ffxiB)
	cfg := config.Config{
		Version: 1,
		Installs: map[string]config.Install{
			"steam":  {Path: normA},
			"lutris": {Path: normB},
		},
	}
	session := testSession(t, cfg)
	setInteractiveStdin(t, "2\n")

	ctx, err := promptInstallSelect(session, []string{"steam", "lutris"})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.installName != "lutris" {
		t.Errorf("installName = %q, want lutris", ctx.installName)
	}
}

func TestRecoverCorruptConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(path, []byte("version: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("backup and reset", func(t *testing.T) {
		setInteractiveStdin(t, "b\n")
		cfg, err := recoverCorruptConfig(path, errInvalidConfig)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Version != 1 {
			t.Errorf("version = %d, want 1", cfg.Version)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		var hasBackup, hasFresh bool
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "config.yml.bak.") {
				hasBackup = true
			}
			if e.Name() == "config.yml" {
				hasFresh = true
			}
		}
		if !hasBackup {
			t.Error("expected backup file")
		}
		if !hasFresh {
			t.Error("expected fresh config.yml")
		}
	})

	t.Run("quit", func(t *testing.T) {
		if err := os.WriteFile(path, []byte("version: 99\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		setInteractiveStdin(t, "q\n")
		if _, err := recoverCorruptConfig(path, errInvalidConfig); err == nil {
			t.Fatal("expected error on quit")
		}
	})
}

func TestOpenConfig_InteractiveRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("version: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)
	setInteractiveStdin(t, "\n") // default B

	session, err := openConfig()
	if err != nil {
		t.Fatal(err)
	}
	if session.cfg.Version != 1 {
		t.Errorf("version = %d, want 1", session.cfg.Version)
	}
}

func TestPromptCharSelect_Interactive(t *testing.T) {
	ffxiDir, userDir, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	norm, _ := config.NormalizePath(ffxiDir)
	cfg := config.Config{
		Version: 1,
		Installs: map[string]config.Install{
			"test": {
				Path: norm,
				Characters: map[string]config.Character{
					"a1b2c3d4": {Name: "Squatched"},
					"e5f6a7b8": {Name: "Alt"},
				},
			},
		},
	}
	session := testSession(t, cfg)
	setInteractiveStdin(t, "2\n")

	chars, err := lister.DiscoverCharacters(userDir)
	if err != nil {
		t.Fatal(err)
	}
	inst := cfg.Installs["test"]
	dirs, err := promptCharSelect(chars, &inst, "test", session)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 || filepath.Base(dirs[0]) != "e5f6a7b8" {
		t.Errorf("dirs = %v, want e5f6a7b8", dirs)
	}
}

func TestPromptConfiguredCharSelect_Interactive(t *testing.T) {
	ffxiDir, userDir, _ := makeFFXITree(t, "a1b2c3d4", "e5f6a7b8")
	norm, _ := config.NormalizePath(ffxiDir)
	inst := config.Install{
		Path: norm,
		Characters: map[string]config.Character{
			"a1b2c3d4": {Name: "Squatched"},
			"e5f6a7b8": {Name: "AltMule"},
		},
	}
	session := testSession(t, config.Config{Version: 1})
	setInteractiveStdin(t, "1\n")

	dirs, err := promptConfiguredCharSelect(&inst, userDir, session)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) != 1 {
		t.Fatalf("dirs = %v", dirs)
	}
	// Sorted by name: AltMule first, Squatched second — pick 1 = AltMule
	if filepath.Base(dirs[0]) != "e5f6a7b8" {
		t.Errorf("got %q, want e5f6a7b8", filepath.Base(dirs[0]))
	}
}

func TestResolveInstall_RegistersInteractively(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	session := testSession(t, config.Empty())
	setInteractiveStdin(t, "\n\n") // accept, default name

	ctx, err := resolveInstall(session, installOpts{ffxiPath: ffxiDir})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.install == nil {
		t.Fatal("expected registered install")
	}
	cfg, _ := config.Load(session.path)
	if len(cfg.Installs) != 1 {
		t.Errorf("installs = %v, want one entry", cfg.Installs)
	}
}

var errInvalidConfig = errInvalidConfigSentinel{}

type errInvalidConfigSentinel struct{}

func (errInvalidConfigSentinel) Error() string {
	return "unsupported config version 99 (expected 1)"
}
