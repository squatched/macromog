package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

type configShowResult struct {
	Path   string        `json:"path"`
	Config config.Config `json:"config"`
}

func newConfigCmd(state *cliState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <subcommand>",
		Short: "manage installs, character aliases, and CLI preferences",
		Long: `Manage installs, character aliases, and CLI preferences.

Examples:
  macromog config path
  macromog config add-install steam "/path/to/FINAL FANTASY XI"
  macromog config set-alias a1b2c3d4 Squatched
  macromog config set-alias a1b2c3d4 Squatched --install lutris
  macromog config default-offering false`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Usage()
			state.code = 1
			return nil
		},
	}

	cmd.AddCommand(
		newConfigPathCmd(state),
		newConfigShowCmd(state),
		newConfigAddInstallCmd(state),
		newConfigRemoveInstallCmd(state),
		newConfigSetDefaultInstallCmd(state),
		newConfigRemoveDefaultInstallCmd(state),
		newConfigSetAliasCmd(state),
		newConfigRemoveAliasCmd(state),
		newConfigSetDefaultCharacterCmd(state),
		newConfigRemoveDefaultCharacterCmd(state),
		newConfigDefaultOfferingCmd(state),
	)

	return cmd
}

func newConfigPathCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "print config file location",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			path, err := config.Path()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) { fmt.Fprintln(tw, path) })
			p.JSON(struct {
				Path string `json:"path"`
			}{Path: path})
			return nil
		},
	}
}

func newConfigShowCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "dump current config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				data, _ := config.MarshalYAML(session.cfg)
				fmt.Fprint(tw, string(data))
			})
			p.JSON(configShowResult{Path: session.path, Config: session.cfg})
			return nil
		},
	}
}

func newConfigAddInstallCmd(state *cliState) *cobra.Command {
	var setDefault bool
	cmd := &cobra.Command{
		Use:   "add-install <name> <path>",
		Short: "register an FFXI install",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			name, rawPath := args[0], args[1]
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			norm, err := config.CanonicalInstallPath(rawPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			access, err := config.ResolveInstallPath(norm)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			userDir := lister.UserDirFromFFXIPath(access)
			if st, err := os.Stat(userDir); err != nil || !st.IsDir() {
				fmt.Fprintf(os.Stderr, "macromog config: USER directory not found under %s\n", norm)
				state.code = 1
				return nil
			}
			if session.cfg.Installs == nil {
				session.cfg.Installs = make(map[string]config.Install)
			}
			if _, exists := session.cfg.Installs[name]; exists {
				fmt.Fprintf(os.Stderr, "macromog config: install %q already exists\n", name)
				state.code = 1
				return nil
			}
			wasFirst := len(session.cfg.Installs) == 0
			session.cfg.Installs[name] = config.Install{Path: norm}
			if wasFirst || setDefault {
				session.cfg.DefaultInstall = name
			}
			if err := session.save(); err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				if wasFirst || setDefault {
					fmt.Fprintf(tw, "added install %q as default\n", tw.Success(name))
				} else {
					fmt.Fprintf(tw, "added install %q (default install is still %s)\n", tw.Success(name), tw.Highlight(session.cfg.DefaultInstall))
				}
			})
			return nil
		},
	}
	cmd.Flags().BoolVar(&setDefault, "set-default", false, "set as default install")
	return cmd
}

func newConfigRemoveInstallCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-install <name>",
		Short: "remove an install and its aliases",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			name := args[0]
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			if _, ok := session.cfg.Installs[name]; !ok {
				fmt.Fprintf(os.Stderr, "macromog config: install %q not found\n", name)
				state.code = 1
				return nil
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
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				fmt.Fprintf(tw, "removed install %q\n", tw.Highlight(name))
			})
			return nil
		},
	}
}

func newConfigSetDefaultInstallCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "set-default-install <name>",
		Short: "set default_install",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			name := args[0]
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			if _, ok := session.cfg.Installs[name]; !ok {
				fmt.Fprintf(os.Stderr, "macromog config: install %q not found\n", name)
				state.code = 1
				return nil
			}
			session.cfg.DefaultInstall = name
			if err := session.save(); err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				fmt.Fprintf(tw, "default install set to %s\n", tw.Success(name))
			})
			return nil
		},
	}
}

func newConfigRemoveDefaultInstallCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-default-install",
		Short: "remove default_install",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			session.cfg.DefaultInstall = ""
			if err := session.save(); err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) { fmt.Fprintln(tw, "default install removed") })
			return nil
		},
	}
}

func newConfigSetAliasCmd(state *cliState) *cobra.Command {
	var installFlag string
	var setDefault bool
	cmd := &cobra.Command{
		Use:   "set-alias <char-id> <name>",
		Short: "give a character a friendly name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			charID, name := args[0], args[1]
			if strings.TrimSpace(name) == "" {
				fmt.Fprintln(os.Stderr, "macromog config set-alias: name must not be empty")
				state.code = 1
				return nil
			}
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			installName, inst, err := resolveConfigInstall(session, installFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			access, err := config.ResolveInstallPath(inst.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			userDir := lister.UserDirFromFFXIPath(access)
			charDir := filepath.Join(userDir, charID)
			if st, err := os.Stat(charDir); err != nil || !st.IsDir() {
				fmt.Fprintf(os.Stderr, "macromog config: %q is not a valid character directory in %s\n", charID, userDir)
				state.code = 1
				return nil
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
				state.code = 1
				return nil
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
			return nil
		},
	}
	cmd.Flags().StringVar(&installFlag, "install", "", "named FFXI install from config")
	cmd.Flags().BoolVar(&setDefault, "set-default", false, "set as default character for this install")
	return cmd
}

func newConfigRemoveAliasCmd(state *cliState) *cobra.Command {
	var installFlag string
	cmd := &cobra.Command{
		Use:   "remove-alias <char-id>",
		Short: "remove a character alias",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			charID := args[0]
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			installName, inst, err := resolveConfigInstall(session, installFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			if inst.Characters == nil {
				fmt.Fprintf(os.Stderr, "macromog config: no alias set for %q\n", charID)
				state.code = 1
				return nil
			}
			if _, ok := inst.Characters[charID]; !ok {
				fmt.Fprintf(os.Stderr, "macromog config: no alias set for %q\n", charID)
				state.code = 1
				return nil
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
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				fmt.Fprintf(tw, "alias removed: %s\n", tw.Highlight(charID))
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&installFlag, "install", "", "named FFXI install from config")
	return cmd
}

func newConfigSetDefaultCharacterCmd(state *cliState) *cobra.Command {
	var installFlag string
	cmd := &cobra.Command{
		Use:   "set-default-character <char-id>",
		Short: "set default_character for an install",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			charID := args[0]
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			installName, inst, err := resolveConfigInstall(session, installFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			if inst.Characters == nil {
				fmt.Fprintf(os.Stderr, "macromog config: character %q is not configured\n", charID)
				state.code = 1
				return nil
			}
			if _, ok := inst.Characters[charID]; !ok {
				fmt.Fprintf(os.Stderr, "macromog config: character %q is not configured\n", charID)
				state.code = 1
				return nil
			}
			inst.DefaultCharacter = charID
			session.cfg.Installs[installName] = *inst
			if err := session.save(); err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				fmt.Fprintf(tw, "default character set to %s\n", tw.Success(charID))
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&installFlag, "install", "", "named FFXI install from config")
	return cmd
}

func newConfigRemoveDefaultCharacterCmd(state *cliState) *cobra.Command {
	var installFlag string
	cmd := &cobra.Command{
		Use:   "remove-default-character",
		Short: "remove default_character for an install",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			installName, inst, err := resolveConfigInstall(session, installFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			inst.DefaultCharacter = ""
			session.cfg.Installs[installName] = *inst
			if err := session.save(); err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) { fmt.Fprintln(tw, "default character removed") })
			return nil
		},
	}
	cmd.Flags().StringVar(&installFlag, "install", "", "named FFXI install from config")
	return cmd
}

func newConfigDefaultOfferingCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "default-offering <true|false>",
		Short: "enable or disable default-setting tips",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := state.printer
			value, err := config.ParseBool(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			session, err := openConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			if session.cfg.Preferences == nil {
				session.cfg.Preferences = &config.Preferences{}
			}
			session.cfg.Preferences.DefaultOffering = &value
			if err := session.save(); err != nil {
				fmt.Fprintf(os.Stderr, "macromog config: %v\n", err)
				state.code = 1
				return nil
			}
			p.Text(func(tw *TextWriter) {
				if value {
					fmt.Fprintln(tw, "Default offering enabled. Install and character selection prompts are unchanged; default-setting tips are shown.")
				} else {
					fmt.Fprintln(tw, "Default offering disabled. Install and character selection prompts are unchanged; default-setting tips are suppressed.")
				}
			})
			return nil
		},
	}
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

func runConfig(args []string, p *Printer) int {
	state := &cliState{printer: p, out: os.Stdout}
	return execWithState(newConfigCmd(state), args, state)
}
