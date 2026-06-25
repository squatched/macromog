package main

import (
	"fmt"
	"os"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

type registerInstallOpts struct {
	setDefault bool
}

func registerInstall(session *configSession, name, rawPath string, opts registerInstallOpts) error {
	norm, err := config.CanonicalInstallPath(rawPath)
	if err != nil {
		return err
	}
	access, err := config.ResolveInstallPath(norm)
	if err != nil {
		return err
	}
	userDir := lister.UserDirFromFFXIPath(access)
	if st, err := os.Stat(userDir); err != nil || !st.IsDir() {
		return fmt.Errorf("USER directory not found under %s", norm)
	}
	if session.cfg.Installs == nil {
		session.cfg.Installs = make(map[string]config.Install)
	}
	if _, exists := session.cfg.Installs[name]; exists {
		return fmt.Errorf("install %q already exists", name)
	}
	wasFirst := len(session.cfg.Installs) == 0
	session.cfg.Installs[name] = config.Install{Path: norm}
	if wasFirst || opts.setDefault {
		session.cfg.DefaultInstall = name
	}
	return session.save()
}
