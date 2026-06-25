package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"

	"github.com/spf13/cobra"
	"github.com/squatched/macromog/internal/config"
	"github.com/squatched/macromog/internal/lister"
)

type debugEnvResult struct {
	Environment []string `json:"environment"`
}

type debugPathsResult struct {
	GOOS               string `json:"goos"`
	GOARCH             string `json:"goarch"`
	RunningUnderWine   bool   `json:"running_under_wine"`
	ConfigPath         string `json:"config_path"`
	ConfigPathErr      string `json:"config_path_error,omitempty"`
	ConfigOpenPath     string `json:"config_open_path"`
	ConfigOpenPathErr  string `json:"config_open_path_error,omitempty"`
	LinuxHome          string `json:"linux_home,omitempty"`
	LinuxHomeFound     bool   `json:"linux_home_found"`
	WinePrefix         string `json:"wine_prefix,omitempty"`
	WinePrefixErr      string `json:"wine_prefix_error,omitempty"`
	UserHomeDir        string `json:"user_home_dir,omitempty"`
	UserHomeDirErr     string `json:"user_home_dir_error,omitempty"`
	UserConfigDir      string `json:"user_config_dir,omitempty"`
	UserConfigDirErr   string `json:"user_config_dir_error,omitempty"`
	MacromogConfigEnv  string `json:"macromog_config_env,omitempty"`
	DetectedUserDir    string `json:"detected_user_dir,omitempty"`
	DetectedUserDirErr string `json:"detected_user_dir_error,omitempty"`
}

type debugAllResult struct {
	Paths debugPathsResult `json:"paths"`
	Env   debugEnvResult   `json:"env"`
}

func newDebugCmd(state *cliState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "diagnostic probes for path and environment detection",
	}
	cmd.AddCommand(
		newDebugEnvCmd(state),
		newDebugPathsCmd(state),
		newDebugAllCmd(state),
	)
	return cmd
}

func newDebugEnvCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "dump the full process environment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return emitDebugEnv(state)
		},
	}
}

func newDebugPathsCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "paths",
		Short: "dump config and install path resolution probes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return emitDebugPaths(state)
		},
	}
}

func newDebugAllCmd(state *cliState) *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "dump paths and full environment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := collectDebugPaths()
			env := collectDebugEnv()
			state.printer.Text(func(tw *TextWriter) {
				writeDebugPathsText(tw, paths)
				fmt.Fprintln(tw)
				writeDebugEnvText(tw, env.Environment)
			})
			state.printer.JSON(debugAllResult{Paths: paths, Env: env})
			return nil
		},
	}
}

func emitDebugEnv(state *cliState) error {
	env := collectDebugEnv()
	state.printer.Text(func(tw *TextWriter) {
		writeDebugEnvText(tw, env.Environment)
	})
	state.printer.JSON(env)
	return nil
}

func emitDebugPaths(state *cliState) error {
	paths := collectDebugPaths()
	state.printer.Text(func(tw *TextWriter) {
		writeDebugPathsText(tw, paths)
	})
	state.printer.JSON(paths)
	return nil
}

func collectDebugEnv() debugEnvResult {
	env := append([]string(nil), os.Environ()...)
	sort.Strings(env)
	return debugEnvResult{Environment: env}
}

func collectDebugPaths() debugPathsResult {
	out := debugPathsResult{
		GOOS:              runtime.GOOS,
		GOARCH:            runtime.GOARCH,
		RunningUnderWine:  config.RunningUnderWine(),
		MacromogConfigEnv: os.Getenv("MACROMOG_CONFIG"),
	}

	if home, ok := config.LinuxHomeForSharedConfig(); ok {
		out.LinuxHome = home
		out.LinuxHomeFound = true
	}

	if path, err := config.Path(); err != nil {
		out.ConfigPathErr = err.Error()
	} else {
		out.ConfigPath = path
		if open, err := config.OpenPath(path); err != nil {
			out.ConfigOpenPathErr = err.Error()
		} else {
			out.ConfigOpenPath = open
		}
	}

	if prefix, err := config.WinePrefixDir(); err != nil {
		out.WinePrefixErr = err.Error()
	} else {
		out.WinePrefix = prefix
	}

	if home, err := os.UserHomeDir(); err != nil {
		out.UserHomeDirErr = err.Error()
	} else {
		out.UserHomeDir = home
	}

	if cfgDir, err := os.UserConfigDir(); err != nil {
		out.UserConfigDirErr = err.Error()
	} else {
		out.UserConfigDir = cfgDir
	}

	if userDir, err := lister.DetectUserDir(); err != nil {
		out.DetectedUserDirErr = err.Error()
	} else {
		out.DetectedUserDir = userDir
	}

	return out
}

func writeDebugEnvText(tw *TextWriter, env []string) {
	fmt.Fprintln(tw, "environment:")
	for _, line := range env {
		fmt.Fprintln(tw, line)
	}
}

func writeDebugPathsText(tw *TextWriter, paths debugPathsResult) {
	lines := []struct {
		key string
		val string
	}{
		{"goos", paths.GOOS},
		{"goarch", paths.GOARCH},
		{"running_under_wine", fmt.Sprintf("%v", paths.RunningUnderWine)},
		{"macromog_config", paths.MacromogConfigEnv},
		{"config_path", paths.ConfigPath},
		{"config_open_path", paths.ConfigOpenPath},
		{"linux_home", paths.LinuxHome},
		{"wine_prefix", paths.WinePrefix},
		{"user_home_dir", paths.UserHomeDir},
		{"user_config_dir", paths.UserConfigDir},
		{"detected_user_dir", paths.DetectedUserDir},
	}
	fmt.Fprintln(tw, "paths:")
	for _, line := range lines {
		if line.val != "" {
			fmt.Fprintf(tw, "  %s: %s\n", line.key, line.val)
		}
	}
	for _, errLine := range []struct {
		key string
		val string
	}{
		{"config_path_error", paths.ConfigPathErr},
		{"config_open_path_error", paths.ConfigOpenPathErr},
		{"wine_prefix_error", paths.WinePrefixErr},
		{"user_home_dir_error", paths.UserHomeDirErr},
		{"user_config_dir_error", paths.UserConfigDirErr},
		{"detected_user_dir_error", paths.DetectedUserDirErr},
	} {
		if errLine.val != "" {
			fmt.Fprintf(tw, "  %s: %s\n", errLine.key, errLine.val)
		}
	}
	if !paths.LinuxHomeFound && paths.LinuxHome == "" {
		fmt.Fprintf(tw, "  linux_home_found: %v\n", paths.LinuxHomeFound)
	}
}
