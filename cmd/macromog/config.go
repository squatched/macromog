package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

const configUsage = `Usage: macromog config <subcommand> [args]

Manage installs, character aliases, and CLI preferences.

Subcommands:
  path                              print config file location
  show                              dump current config
  add-install <name> <path>         register an install [--set-default]
  remove-install <name>             remove an install and its aliases
  set-default-install <name>        set default_install
  remove-default-install            remove default_install
  set-alias <char-id> <name>        give a character a friendly name [--install <name>] [--set-default]
  remove-alias <char-id>            remove an alias [--install <name>]
  set-default-character <char-id>   set default_character [--install <name>]
  remove-default-character          remove default_character [--install <name>]
  default-offering <true|false>     enable or disable default-setting tips

Examples:
  macromog config path
  macromog config add-install steam "/path/to/FINAL FANTASY XI"
  macromog config set-alias a1b2c3d4 Squatched
  macromog config set-alias a1b2c3d4 Squatched --install lutris
  macromog config default-offering false
`

type configShowResult struct {
	Path   string        `json:"path"`
	Config config.Config `json:"config"`
}

func runConfig(args []string, p *Printer) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(os.Stdout, configUsage)
		return 0
	}

	switch args[0] {
	case "path":
		return configPath(p)
	case "show":
		return configShow(p)
	case "add-install":
		return configAddInstall(args[1:], p)
	case "remove-install":
		return configRemoveInstall(args[1:], p)
	case "set-default-install":
		return configSetDefaultInstall(args[1:], p)
	case "remove-default-install":
		return configRemoveDefaultInstall(p)
	case "set-alias":
		return configSetAlias(args[1:], p)
	case "remove-alias":
		return configRemoveAlias(args[1:], p)
	case "set-default-character":
		return configSetDefaultCharacter(args[1:], p)
	case "remove-default-character":
		return configRemoveDefaultCharacter(args[1:], p)
	case "default-offering":
		return configDefaultOffering(args[1:], p)
	default:
		fmt.Fprintf(os.Stderr, "macromog config: unknown subcommand %q\n\n%s", args[0], configUsage)
		return 1
	}
}

func configPath(p *Printer) int {
	path, err := config.Path()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintln(tw, path)
	})
	p.JSON(struct {
		Path string `json:"path"`
	}{Path: path})
	return 0
}

func configShow(p *Printer) int {
	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		data, _ := config.MarshalYAML(session.cfg)
		fmt.Fprint(tw, string(data))
	})
	p.JSON(configShowResult{Path: session.path, Config: session.cfg})
	return 0
}

func configAddInstall(args []string, p *Printer) int {
	setDefault, positional, err := parseConfigFlags(args, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(positional) != 2 {
		fmt.Fprintln(os.Stderr, "macromog config add-install: expected <name> <path>")
		return 1
	}
	name, rawPath := positional[0], positional[1]

	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	norm, err := config.NormalizePath(rawPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	userDir := lister.UserDirFromFFXIPath(norm)
	if st, err := os.Stat(userDir); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "macromog config: USER directory not found under %s\n", norm)
		return 1
	}
	if session.cfg.Installs == nil {
		session.cfg.Installs = make(map[string]config.Install)
	}
	if _, exists := session.cfg.Installs[name]; exists {
		fmt.Fprintf(os.Stderr, "macromog config: install %q already exists\n", name)
		return 1
	}
	wasFirst := len(session.cfg.Installs) == 0
	session.cfg.Installs[name] = config.Install{Path: norm}
	if wasFirst || setDefault {
		session.cfg.DefaultInstall = name
	}
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		if wasFirst || setDefault {
			fmt.Fprintf(tw, "added install %q as default\n", tw.Success(name))
		} else {
			fmt.Fprintf(tw, "added install %q (default install is still %s)\n", tw.Success(name), tw.Highlight(session.cfg.DefaultInstall))
		}
	})
	return 0
}

func configRemoveInstall(args []string, p *Printer) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "macromog config remove-install: expected <name>")
		return 1
	}
	name := args[0]
	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	if _, ok := session.cfg.Installs[name]; !ok {
		fmt.Fprintf(os.Stderr, "macromog config: install %q not found\n", name)
		return 1
	}
	delete(session.cfg.Installs, name)
	if session.cfg.DefaultInstall == name {
		session.cfg.DefaultInstall = ""
	}
	if len(session.cfg.Installs) == 0 {
		session.cfg.Installs = nil
	}
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "removed install %q\n", tw.Highlight(name))
	})
	return 0
}

func configSetDefaultInstall(args []string, p *Printer) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "macromog config set-default-install: expected <name>")
		return 1
	}
	name := args[0]
	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	if _, ok := session.cfg.Installs[name]; !ok {
		fmt.Fprintf(os.Stderr, "macromog config: install %q not found\n", name)
		return 1
	}
	session.cfg.DefaultInstall = name
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "default install set to %s\n", tw.Success(name))
	})
	return 0
}

func configRemoveDefaultInstall(p *Printer) int {
	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	session.cfg.DefaultInstall = ""
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintln(tw, "default install removed")
	})
	return 0
}

func configSetAlias(args []string, p *Printer) int {
	installFlag := ""
	stringFlags := map[string]*string{"install": &installFlag}
	setDefault, remaining, err := parseConfigFlags(args, stringFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(remaining) != 2 {
		fmt.Fprintln(os.Stderr, "macromog config set-alias: expected <char-id> <name>")
		return 1
	}
	charID, name := remaining[0], remaining[1]
	if strings.TrimSpace(name) == "" {
		fmt.Fprintln(os.Stderr, "macromog config set-alias: name must not be empty")
		return 1
	}

	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	installName, inst, err := resolveConfigInstall(session, installFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	userDir := lister.UserDirFromFFXIPath(inst.Path)
	charDir := filepath.Join(userDir, charID)
	if st, err := os.Stat(charDir); err != nil || !st.IsDir() {
		fmt.Fprintf(os.Stderr, "macromog config: %q is not a valid character directory in %s\n", charID, userDir)
		return 1
	}
	if inst.Characters == nil {
		inst.Characters = make(map[string]config.Character)
	}
	wasFirst := len(inst.Characters) == 0
	oldDefault := inst.DefaultCharacter
	inst.Characters[charID] = config.Character{Name: name}
	if wasFirst || setDefault {
		inst.DefaultCharacter = charID
	}
	session.cfg.Installs[installName] = *inst
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "alias set: %s → %s\n", tw.Highlight(charID), tw.Success(name))
		if !wasFirst && !setDefault && oldDefault != "" && oldDefault != charID {
			fmt.Fprintf(tw, "default character is still %s. Use 'macromog config set-default-character %s", tw.Highlight(oldDefault), charID)
			if installName != session.cfg.DefaultInstall {
				fmt.Fprintf(tw, " --install %s", installName)
			}
			fmt.Fprintln(tw, "' to change.")
		}
	})
	return 0
}

func configRemoveAlias(args []string, p *Printer) int {
	installFlag := ""
	stringFlags := map[string]*string{"install": &installFlag}
	_, remaining, err := parseConfigFlags(args, stringFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(remaining) != 1 {
		fmt.Fprintln(os.Stderr, "macromog config remove-alias: expected <char-id>")
		return 1
	}
	charID := remaining[0]

	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	installName, inst, err := resolveConfigInstall(session, installFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	if inst.Characters == nil {
		fmt.Fprintf(os.Stderr, "macromog config: no alias set for %q\n", charID)
		return 1
	}
	if _, ok := inst.Characters[charID]; !ok {
		fmt.Fprintf(os.Stderr, "macromog config: no alias set for %q\n", charID)
		return 1
	}
	delete(inst.Characters, charID)
	if inst.DefaultCharacter == charID {
		inst.DefaultCharacter = ""
	}
	if len(inst.Characters) == 0 {
		inst.Characters = nil
	}
	session.cfg.Installs[installName] = *inst
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "alias removed: %s\n", tw.Highlight(charID))
	})
	return 0
}

func configSetDefaultCharacter(args []string, p *Printer) int {
	installFlag := ""
	stringFlags := map[string]*string{"install": &installFlag}
	_, remaining, err := parseConfigFlags(args, stringFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(remaining) != 1 {
		fmt.Fprintln(os.Stderr, "macromog config set-default-character: expected <char-id>")
		return 1
	}
	charID := remaining[0]

	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	installName, inst, err := resolveConfigInstall(session, installFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	if inst.Characters == nil {
		fmt.Fprintf(os.Stderr, "macromog config: character %q is not configured\n", charID)
		return 1
	}
	if _, ok := inst.Characters[charID]; !ok {
		fmt.Fprintf(os.Stderr, "macromog config: character %q is not configured\n", charID)
		return 1
	}
	inst.DefaultCharacter = charID
	session.cfg.Installs[installName] = *inst
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintf(tw, "default character set to %s\n", tw.Success(charID))
	})
	return 0
}

func configRemoveDefaultCharacter(args []string, p *Printer) int {
	installFlag := ""
	stringFlags := map[string]*string{"install": &installFlag}
	_, remaining, err := parseConfigFlags(args, stringFlags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(remaining) != 0 {
		fmt.Fprintln(os.Stderr, "macromog config remove-default-character: unexpected arguments")
		return 1
	}

	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	installName, inst, err := resolveConfigInstall(session, installFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	inst.DefaultCharacter = ""
	session.cfg.Installs[installName] = *inst
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		fmt.Fprintln(tw, "default character removed")
	})
	return 0
}

func configDefaultOffering(args []string, p *Printer) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "macromog config default-offering: expected <true|false>")
		return 1
	}
	value, err := config.ParseBool(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	session, err := openConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	if session.cfg.Preferences == nil {
		session.cfg.Preferences = &config.Preferences{}
	}
	session.cfg.Preferences.DefaultOffering = &value
	if err := session.save(); err != nil {
		fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
		return 1
	}
	p.Text(func(tw *TextWriter) {
		if value {
			fmt.Fprintln(tw, "Default offering enabled. Install and character selection prompts are unchanged; default-setting tips are shown.")
		} else {
			fmt.Fprintln(tw, "Default offering disabled. Install and character selection prompts are unchanged; default-setting tips are suppressed.")
		}
	})
	return 0
}

func resolveConfigInstall(session *configSession, installFlag string) (string, *config.Install, error) {
	name := installFlag
	if name == "" {
		name = session.cfg.DefaultInstall
	}
	if name == "" {
		names := config.InstallNames(&session.cfg)
		if len(names) == 1 {
			name = names[0]
		} else {
			return "", nil, fmt.Errorf("--install is required when default_install is not set")
		}
	}
	inst, ok := session.cfg.Installs[name]
	if !ok {
		return "", nil, fmt.Errorf("install %q not found in config", name)
	}
	return name, &inst, nil
}
