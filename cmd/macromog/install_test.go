package main

import (
	"path/filepath"
	"testing"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

func testSession(t *testing.T, cfg config.Config) *configSession {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := config.Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MACROMOG_CONFIG", path)
	return &configSession{path: path, cfg: cfg}
}

func TestResolveInstall_MutuallyExclusive(t *testing.T) {
	session := testSession(t, config.Empty())
	_, err := resolveInstall(session, installOpts{ffxiPath: "/a", installName: "steam"})
	if err == nil {
		t.Fatal("expected mutual exclusion error")
	}
}

func TestResolveInstall_NamedInstall(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
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
	session := testSession(t, cfg)

	ctx, err := resolveInstall(session, installOpts{installName: "steam"})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.installName != "steam" || ctx.install == nil {
		t.Fatalf("ctx = %+v", ctx)
	}
	if ctx.ffxiPath != norm {
		t.Errorf("ffxiPath = %q, want %q", ctx.ffxiPath, norm)
	}
}

func TestResolveInstall_NamedInstall_NotFound(t *testing.T) {
	session := testSession(t, config.Empty())
	_, err := resolveInstall(session, installOpts{installName: "ghost"})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestResolveInstall_ExplicitPath_Registered(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	norm, err := config.NormalizePath(ffxiDir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Version:        1,
		DefaultInstall: "steam",
		Installs: map[string]config.Install{
			"steam": {
				Path: norm,
				Characters: map[string]config.Character{
					"a1b2c3d4": {Name: "Squatched"},
				},
			},
		},
	}
	session := testSession(t, cfg)

	ctx, err := resolveInstall(session, installOpts{ffxiPath: ffxiDir})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.installName != "steam" {
		t.Errorf("installName = %q, want steam", ctx.installName)
	}
	if ctx.install == nil || ctx.install.Characters["a1b2c3d4"].Name != "Squatched" {
		t.Errorf("install context missing aliases: %+v", ctx.install)
	}
}

func TestResolveInstall_ExplicitPath_Unregistered_NonInteractive(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	norm, err := config.NormalizePath(ffxiDir)
	if err != nil {
		t.Fatal(err)
	}
	session := testSession(t, config.Empty())

	ctx, err := resolveInstall(session, installOpts{ffxiPath: ffxiDir})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.ffxiPath != norm {
		t.Errorf("ffxiPath = %q, want %q", ctx.ffxiPath, norm)
	}
	if ctx.install != nil {
		t.Error("unregistered path should not attach install context")
	}
}

func TestResolveInstall_DefaultInstall(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	norm, err := config.NormalizePath(ffxiDir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Version:        1,
		DefaultInstall: "steam",
		Installs: map[string]config.Install{
			"steam":  {Path: norm},
			"lutris": {Path: absPath(t, t.TempDir(), "lutris")},
		},
	}
	session := testSession(t, cfg)

	ctx, err := resolveInstall(session, installOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.installName != "steam" {
		t.Errorf("installName = %q, want steam", ctx.installName)
	}
}

func TestResolveInstall_SingleInstall(t *testing.T) {
	ffxiDir, _, _ := makeFFXITree(t, "a1b2c3d4")
	norm, err := config.NormalizePath(ffxiDir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Version: 1,
		Installs: map[string]config.Install{
			"only": {Path: norm},
		},
	}
	session := testSession(t, cfg)

	ctx, err := resolveInstall(session, installOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.installName != "only" {
		t.Errorf("installName = %q, want only", ctx.installName)
	}
}

func TestResolveInstall_MultipleInstalls_NonInteractive(t *testing.T) {
	if _, err := lister.DetectUserDir(); err == nil {
		t.Skip("FFXI is auto-detectable on this machine; multi-install error path requires detection to fail")
	}

	cfg := config.Config{
		Version: 1,
		Installs: map[string]config.Install{
			"steam":  {Path: absPath(t, t.TempDir(), "steam")},
			"lutris": {Path: absPath(t, t.TempDir(), "lutris")},
		},
	}
	session := testSession(t, cfg)

	_, err := resolveInstall(session, installOpts{})
	if err == nil {
		t.Fatal("expected error when multiple installs and no default")
	}
}

func TestResolveInstall_ExplicitPath_NoUSER(t *testing.T) {
	session := testSession(t, config.Empty())
	_, err := resolveInstall(session, installOpts{ffxiPath: t.TempDir()})
	if err == nil {
		t.Fatal("expected USER not found error")
	}
}

func absPath(t *testing.T, elems ...string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join(elems...))
	if err != nil {
		t.Fatal(err)
	}
	return p
}
