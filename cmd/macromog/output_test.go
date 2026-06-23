package main

import (
	"testing"
)

func TestExtractOutputFormat(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantFormat OutputFormat
		wantArgs   []string
		wantErr    bool
	}{
		{
			name:       "no args",
			args:       nil,
			wantFormat: FormatText,
			wantArgs:   nil,
		},
		{
			name:       "no output flag",
			args:       []string{"macromog", "list"},
			wantFormat: FormatText,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "json before subcommand",
			args:       []string{"macromog", "--output", "json", "list"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "text before subcommand",
			args:       []string{"macromog", "--output", "text", "list"},
			wantFormat: FormatText,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "json after subcommand",
			args:       []string{"macromog", "list", "--output", "json"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "list"},
		},
		{
			name:       "json after subcommand with other flags",
			args:       []string{"macromog", "list", "--char", "/foo", "--output", "json"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "list", "--char", "/foo"},
		},
		{
			name:       "subcommand --output file passthrough",
			args:       []string{"macromog", "export", "--output", "macros.yml"},
			wantFormat: FormatText,
			wantArgs:   []string{"macromog", "export", "--output", "macros.yml"},
		},
		{
			name:       "global json + subcommand --output file",
			args:       []string{"macromog", "--output", "json", "export", "--output", "macros.yml"},
			wantFormat: FormatJSON,
			wantArgs:   []string{"macromog", "export", "--output", "macros.yml"},
		},
		{
			name:    "unknown format before subcommand",
			args:    []string{"macromog", "--output", "xml", "list"},
			wantErr: true,
		},
		{
			name:    "missing value before subcommand",
			args:    []string{"macromog", "--output"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			format, args, err := extractOutputFormat(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (format=%q, args=%v)", format, args)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if format != tc.wantFormat {
				t.Errorf("format = %q, want %q", format, tc.wantFormat)
			}
			if len(args) != len(tc.wantArgs) {
				t.Errorf("args = %v, want %v", args, tc.wantArgs)
				return
			}
			for i := range args {
				if args[i] != tc.wantArgs[i] {
					t.Errorf("args[%d] = %q, want %q", i, args[i], tc.wantArgs[i])
				}
			}
		})
	}
}

func TestRun_OutputJSON_PostSubcommand(t *testing.T) {
	// Verify the full dispatch path: --output json placed after the subcommand.
	if got := run([]string{"macromog", "validate", "--output", "json", writeTemp(t, "version: 1\nbooks: {}\n")}); got != 0 {
		t.Errorf("run(validate --output json <valid>) = %d, want 0", got)
	}
}
