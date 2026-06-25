package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const CurrentVersion = 1

// Config is the top-level macromog CLI configuration.
type Config struct {
	Version        int                `yaml:"version" json:"version"`
	Preferences    *Preferences       `yaml:"preferences,omitempty" json:"preferences,omitempty"`
	DefaultInstall string             `yaml:"default_install,omitempty" json:"default_install,omitempty"`
	Installs       map[string]Install `yaml:"installs,omitempty" json:"installs,omitempty"`
}

// Preferences holds CLI behavior preferences.
type Preferences struct {
	DefaultOffering *bool `yaml:"default_offering,omitempty" json:"default_offering,omitempty"`
}

// Install describes one registered FFXI install root.
type Install struct {
	Path             string               `yaml:"path" json:"path"`
	DefaultCharacter string               `yaml:"default_character,omitempty" json:"default_character,omitempty"`
	Characters       map[string]Character `yaml:"characters,omitempty" json:"characters,omitempty"`
}

// Character holds a friendly alias for one USER folder ID.
type Character struct {
	Name string `yaml:"name" json:"name"`
}

// Path returns the canonical config file location for the current environment.
// MACROMOG_CONFIG overrides the default when set to an absolute path.
// Under Wine with a mapped Linux home, this is the host XDG path (/home/...).
func Path() (string, error) {
	return ActiveHostFS().ConfigPath()
}

func configFileInDir(dir string) string {
	dir = normalizeHostPath(dir)
	if strings.HasPrefix(dir, "/home/") {
		return hostpath(dir, "config.yml")
	}
	return filepath.Join(dir, "config.yml")
}

// Empty returns a fresh config with only version set.
func Empty() Config {
	return Config{Version: CurrentVersion}
}

// MarshalYAML returns the YAML encoding of cfg.
func MarshalYAML(cfg Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}

// Load reads and validates the config file at path.
// A missing file is treated as an empty config.
func Load(path string) (Config, error) {
	openPath, err := OpenPath(path)
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(openPath)
	if os.IsNotExist(err) {
		return Empty(), nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	if err := Validate(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Save writes cfg to path atomically (temp file, then rename).
func Save(path string, cfg Config) error {
	if err := Validate(&cfg); err != nil {
		return err
	}
	openPath, err := OpenPath(path)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(openPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "config-*.yml")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, openPath)
}

// Ensure creates an empty config file when path does not exist yet.
func Ensure(path string) error {
	openPath, err := OpenPath(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(openPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return Save(path, Empty())
}

// Validate checks semantic constraints on cfg.
func Validate(cfg *Config) error {
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}
	if cfg.Version != CurrentVersion {
		return fmt.Errorf("unsupported config version %d (expected %d)", cfg.Version, CurrentVersion)
	}
	if cfg.DefaultInstall != "" {
		if cfg.Installs == nil {
			return fmt.Errorf("default_install %q references unknown install", cfg.DefaultInstall)
		}
		if _, ok := cfg.Installs[cfg.DefaultInstall]; !ok {
			return fmt.Errorf("default_install %q references unknown install", cfg.DefaultInstall)
		}
	}
	for name, inst := range cfg.Installs {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("install name must not be empty")
		}
		if strings.TrimSpace(inst.Path) == "" {
			return fmt.Errorf("install %q: path must not be empty", name)
		}
		norm, err := CanonicalInstallPath(inst.Path)
		if err != nil {
			return fmt.Errorf("install %q: %w", name, err)
		}
		if norm != inst.Path {
			return fmt.Errorf("install %q: path must be absolute and normalized (got %q)", name, inst.Path)
		}
		if inst.DefaultCharacter != "" {
			if inst.Characters == nil {
				return fmt.Errorf("install %q: default_character %q is not configured", name, inst.DefaultCharacter)
			}
			if _, ok := inst.Characters[inst.DefaultCharacter]; !ok {
				return fmt.Errorf("install %q: default_character %q is not configured", name, inst.DefaultCharacter)
			}
		}
		seenNames := make(map[string]string)
		for id, ch := range inst.Characters {
			if strings.TrimSpace(id) == "" {
				return fmt.Errorf("install %q: character id must not be empty", name)
			}
			if strings.TrimSpace(ch.Name) == "" {
				return fmt.Errorf("install %q: alias for %q must not be empty", name, id)
			}
			lower := strings.ToLower(ch.Name)
			if prev, ok := seenNames[lower]; ok && prev != id {
				return fmt.Errorf("install %q: duplicate alias %q", name, ch.Name)
			}
			seenNames[lower] = id
		}
	}
	if cfg.Preferences != nil && cfg.Preferences.DefaultOffering != nil {
		// bool is always valid once present
	}
	return nil
}

// NormalizePath expands ~, makes the path absolute, and removes trailing slashes.
// Symlinks are preserved as written.
func NormalizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	expanded := path
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		expanded = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(abs, string(filepath.Separator)), nil
}

// DefaultOffering reports whether default-setting tips are enabled.
// Absent preference means enabled (true).
func DefaultOffering(cfg *Config) bool {
	if cfg.Preferences == nil || cfg.Preferences.DefaultOffering == nil {
		return true
	}
	return *cfg.Preferences.DefaultOffering
}

// ParseBool parses flexible boolean strings for default-offering.
func ParseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "t", "yes", "y":
		return true, nil
	case "false", "f", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q (use true/false, yes/no, y/n)", s)
	}
}

// FindInstallByPath returns the install name whose normalized path matches path.
func FindInstallByPath(cfg *Config, path string) (string, *Install, error) {
	norm, err := CanonicalInstallPath(path)
	if err != nil {
		return "", nil, err
	}
	for name, inst := range cfg.Installs {
		instNorm, err := CanonicalInstallPath(inst.Path)
		if err != nil {
			continue
		}
		if instNorm == norm {
			instCopy := inst
			return name, &instCopy, nil
		}
	}
	return "", nil, nil
}

// ResolveAlias finds the folder ID for a friendly name within inst (case-insensitive).
func ResolveAlias(inst *Install, name string) (string, error) {
	if inst == nil || len(inst.Characters) == 0 {
		return "", fmt.Errorf("no character found with name %q", name)
	}
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("character name must not be empty")
	}
	lower := strings.ToLower(name)
	var matches []string
	for id, ch := range inst.Characters {
		if strings.ToLower(ch.Name) == lower {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no character found with name %q", name)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple characters named %q: %v", name, matches)
	}
}

// LookupName returns the alias for id, or id itself when unset.
func LookupName(inst *Install, id string) string {
	if inst == nil {
		return id
	}
	if ch, ok := inst.Characters[id]; ok && ch.Name != "" {
		return ch.Name
	}
	return id
}

// SuggestInstallName derives a short install name from a filesystem path.
func SuggestInstallName(cfg *Config, path string) string {
	lower := strings.ToLower(path)
	candidates := []struct {
		substr string
		name   string
	}{
		{"steam", "steam"},
		{"lutris", "lutris"},
		{".wine", "wine"},
		{"wine", "wine"},
		{"playonline", "playonline"},
	}
	for _, c := range candidates {
		if strings.Contains(lower, c.substr) {
			return uniqueInstallName(cfg, c.name)
		}
	}
	return uniqueInstallName(cfg, "default")
}

func uniqueInstallName(cfg *Config, base string) string {
	if cfg.Installs == nil {
		return base
	}
	if _, ok := cfg.Installs[base]; !ok {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s.%d", base, i)
		if _, ok := cfg.Installs[candidate]; !ok {
			return candidate
		}
	}
}

// InstallNames returns sorted install names from cfg.
func InstallNames(cfg *Config) []string {
	if len(cfg.Installs) == 0 {
		return nil
	}
	names := make([]string, 0, len(cfg.Installs))
	for name := range cfg.Installs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
