package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/squatched/macromog/internal/config"
)

// configSession holds the loaded CLI config and its file path.
type configSession struct {
	path string
	cfg  config.Config
}

func openConfig() (*configSession, error) {
	path, err := config.Path()
	if err != nil {
		return nil, err
	}
	if err := config.Ensure(path); err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		if stdinIsTerminal() {
			fresh, recoverErr := recoverCorruptConfig(path, err)
			if recoverErr != nil {
				return nil, recoverErr
			}
			cfg = fresh
		} else {
			return nil, fmt.Errorf("config.yml is invalid: %v", err)
		}
	}
	return &configSession{path: path, cfg: cfg}, nil
}

func (s *configSession) save() error {
	return config.Save(s.path, s.cfg)
}

func recoverCorruptConfig(path string, loadErr error) (config.Config, error) {
	ew := newErrWriter()
	fmt.Fprintf(ew, "config.yml is invalid: %v\n\n", loadErr)
	fmt.Fprintln(ew, "  [B]ack up corrupt file and start fresh")
	fmt.Fprintln(ew, "  [Q]uit (fix manually)")
	fmt.Fprint(ew, "\nChoice [B/q]: ")

	line, ok := readStdinLine()
	if !ok {
		return config.Config{}, fmt.Errorf("config.yml is invalid: %v", loadErr)
	}
	choice := strings.ToLower(strings.TrimSpace(line))
	if choice == "" || choice == "b" {
		backup := fmt.Sprintf("%s.bak.%s", path, time.Now().UTC().Format("20060102_150405"))
		if err := os.Rename(path, backup); err != nil {
			return config.Config{}, fmt.Errorf("backing up corrupt config: %w", err)
		}
		cfg := config.Empty()
		if err := config.Save(path, cfg); err != nil {
			return config.Config{}, err
		}
		fmt.Fprintf(ew, "Backed up corrupt config to %s and started fresh.\n", backup)
		return cfg, nil
	}
	return config.Config{}, fmt.Errorf("config.yml is invalid: %v", loadErr)
}
