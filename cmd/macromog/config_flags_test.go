package main

import (
	"testing"
)

func TestParseConfigFlags(t *testing.T) {
	install := ""
	stringFlags := map[string]*string{"install": &install}

	tests := []struct {
		name        string
		args        []string
		wantSetDef  bool
		wantPos     []string
		wantInstall string
		wantErr     bool
	}{
		{
			name:    "positional only",
			args:    []string{"steam", "/path/ffxi"},
			wantPos: []string{"steam", "/path/ffxi"},
		},
		{
			name:       "flag after positional",
			args:       []string{"steam", "/path/ffxi", "--set-default"},
			wantSetDef: true,
			wantPos:    []string{"steam", "/path/ffxi"},
		},
		{
			name:       "flag before positional",
			args:       []string{"--set-default", "steam", "/path/ffxi"},
			wantSetDef: true,
			wantPos:    []string{"steam", "/path/ffxi"},
		},
		{
			name:        "install flag",
			args:        []string{"a1b2c3d4", "Squatched", "--install", "lutris"},
			wantPos:     []string{"a1b2c3d4", "Squatched"},
			wantInstall: "lutris",
		},
		{
			name:    "install flag missing value",
			args:    []string{"--install"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			install = ""
			setDef, pos, err := parseConfigFlags(tc.args, stringFlags)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if setDef != tc.wantSetDef {
				t.Errorf("setDefault = %v, want %v", setDef, tc.wantSetDef)
			}
			if install != tc.wantInstall {
				t.Errorf("install = %q, want %q", install, tc.wantInstall)
			}
			if len(pos) != len(tc.wantPos) {
				t.Fatalf("positional = %v, want %v", pos, tc.wantPos)
			}
			for i := range tc.wantPos {
				if pos[i] != tc.wantPos[i] {
					t.Errorf("pos[%d] = %q, want %q", i, pos[i], tc.wantPos[i])
				}
			}
		})
	}
}
