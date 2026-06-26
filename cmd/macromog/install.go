package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

// installContext is the resolved FFXI install for the current command.
type installContext struct {
	ffxiPath    string
	installName string
	install     *config.Install
}

type installOpts struct {
	ffxiPath    string
	installName string
}

func resolveInstall(session *configSession, opts installOpts) (installContext, error) {
	if opts.ffxiPath != "" && opts.installName != "" {
		return installContext{}, fmt.Errorf("--ffxi-path and --install are mutually exclusive")
	}

	if opts.ffxiPath != "" {
		return resolveExplicitPath(session, opts.ffxiPath, true)
	}
	if opts.installName != "" {
		return resolveNamedInstall(session, opts.installName)
	}
	if session.cfg.DefaultInstall != "" {
		return installFromName(session, session.cfg.DefaultInstall)
	}
	if names := config.InstallNames(&session.cfg); len(names) == 1 {
		return installFromName(session, names[0])
	}

	detected, err := lister.DetectUserDir()
	if err == nil {
		ffxiPath := filepath.Dir(detected)
		return resolveExplicitPath(session, ffxiPath, false)
	}

	names := config.InstallNames(&session.cfg)
	if len(names) > 1 {
		return promptInstallSelect(session, names)
	}

	return installContext{}, fmt.Errorf("FFXI install not found; run 'macromog config add-install' or use --ffxi-path")
}

func resolveExplicitPath(session *configSession, rawPath string, explicit bool) (installContext, error) {
	norm, err := config.CanonicalInstallPath(rawPath)
	if err != nil {
		return installContext{}, err
	}
	access, err := config.ResolveInstallPath(norm)
	if err != nil {
		return installContext{}, err
	}
	userDir := lister.UserDirFromFFXIPath(access)
	if st, err := os.Stat(userDir); err != nil || !st.IsDir() {
		return installContext{}, fmt.Errorf("USER directory not found under %s", norm)
	}
	name, inst, err := config.FindInstallByPath(&session.cfg, norm)
	if err != nil {
		return installContext{}, err
	}
	if inst != nil {
		return installContext{ffxiPath: access, installName: name, install: inst}, nil
	}
	return maybeRegisterInstall(session, norm, explicit)
}

func resolveNamedInstall(session *configSession, name string) (installContext, error) {
	inst, ok := session.cfg.Installs[name]
	if !ok {
		return installContext{}, fmt.Errorf("install %q not found in config", name)
	}
	access, err := config.ResolveInstallPath(inst.Path)
	if err != nil {
		return installContext{}, err
	}
	return installContext{ffxiPath: access, installName: name, install: &inst}, nil
}

func installFromName(session *configSession, name string) (installContext, error) {
	inst, ok := session.cfg.Installs[name]
	if !ok {
		return installContext{}, fmt.Errorf("default install %q not found in config", name)
	}
	access, err := config.ResolveInstallPath(inst.Path)
	if err != nil {
		return installContext{}, err
	}
	return installContext{ffxiPath: access, installName: name, install: &inst}, nil
}

// maybeRegisterInstall offers to register canonicalPath if stdin is a TTY. In
// CI mode, auto-detected (explicit=false) unregistered paths produce an error;
// explicitly-supplied paths are trusted and used as-is.
func maybeRegisterInstall(session *configSession, canonicalPath string, explicit bool) (installContext, error) {
	access, err := config.ResolveInstallPath(canonicalPath)
	if err != nil {
		return installContext{}, err
	}
	if !stdinIsTerminal() {
		if !explicit && isCI() {
			return installContext{}, fmt.Errorf(
				"auto-detected FFXI install at %s is not in config; run 'macromog config add-install' or use --ffxi-path",
				canonicalPath,
			)
		}
		return installContext{ffxiPath: access}, nil
	}
	ew := newErrWriter()
	fmt.Fprintf(ew, "Auto-detected FFXI install at %s, not in config. Register? [Y/n] ", ew.Highlight(canonicalPath))
	answerLine, ok := readStdinLine()
	if !ok {
		return installContext{ffxiPath: access}, nil
	}
	answer := strings.ToLower(strings.TrimSpace(answerLine))
	if answer == "n" || answer == "no" {
		return installContext{ffxiPath: access}, nil
	}

	suggested := config.SuggestInstallName(&session.cfg, canonicalPath)
	fmt.Fprintf(ew, "Name [%s]: ", suggested)
	name := suggested
	if nameLine, ok := readStdinLine(); ok {
		if typed := strings.TrimSpace(nameLine); typed != "" {
			name = typed
		}
	}
	if err := addInstallToConfig(session, name, canonicalPath); err != nil {
		return installContext{}, err
	}
	inst := session.cfg.Installs[name]
	return installContext{ffxiPath: access, installName: name, install: &inst}, nil
}

func addInstallToConfig(session *configSession, name, rawPath string) error {
	wasFirst := len(session.cfg.Installs) == 0
	if err := registerInstall(session, name, rawPath, registerInstallOpts{}); err != nil {
		return err
	}
	ew := newErrWriter()
	if wasFirst || session.cfg.DefaultInstall == name {
		fmt.Fprintf(ew, "Added install %q as default.\n", name)
	} else {
		fmt.Fprintf(ew, "Added install %q. Default install is still %q. Use 'macromog config set-default-install %s' to change.\n",
			name, session.cfg.DefaultInstall, name)
	}
	return nil
}

func promptInstallSelect(session *configSession, names []string) (installContext, error) {
	if !stdinIsTerminal() {
		return installContext{}, fmt.Errorf("multiple installs configured; use --install to specify one")
	}
	ew := newErrWriter()
	fmt.Fprintln(ew, ew.Bold("Multiple installs configured. Which install for this command?"))
	for i, name := range names {
		fmt.Fprintf(ew, "  %s %s\n", ew.Muted(fmt.Sprintf("[%d]", i+1)), ew.Bold(name))
	}
	fmt.Fprint(ew, "Enter number: ")

	line, ok := readStdinLine()
	if !ok {
		return installContext{}, fmt.Errorf("invalid selection; use --install to specify one")
	}
	indices, err := parseSelection(line, len(names))
	if err != nil || len(indices) != 1 {
		return installContext{}, fmt.Errorf("invalid selection; use --install to specify one")
	}
	selected := names[indices[0]]
	if config.DefaultOffering(&session.cfg) && session.cfg.DefaultInstall == "" {
		fmt.Fprintf(ew, "\nTip: macromog config set-default-install %s skips this prompt.\n", selected)
	}
	return installFromName(session, selected)
}
