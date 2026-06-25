package config

import (
	"strings"
	"testing"
)

func TestHostpath_ForwardSlashes(t *testing.T) {
	got := hostpath("/home", "squatched", ".config", "macromog")
	want := "/home/squatched/.config/macromog"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if strings.Contains(got, `\`) {
		t.Errorf("hostpath contains backslashes: %q", got)
	}
}

func TestNormalizeHostPath_MixedSlashes(t *testing.T) {
	got := normalizeHostPath(`/home\squatched/.config/macromog/config.yml`)
	want := "/home/squatched/.config/macromog/config.yml"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLinuxHomeFromPath_NoBackslashes(t *testing.T) {
	cases := []string{
		"/home/squatched/Games/final-fantasy-xi-online",
		`Z:\home\squatched\Games\final-fantasy-xi-online`,
	}
	for _, in := range cases {
		got, ok := linuxHomeFromPath(in)
		if !ok {
			t.Fatalf("linuxHomeFromPath(%q) not ok", in)
		}
		if got != "/home/squatched" {
			t.Errorf("linuxHomeFromPath(%q) = %q, want /home/squatched", in, got)
		}
		if strings.Contains(got, `\`) {
			t.Errorf("linuxHomeFromPath(%q) contains backslashes: %q", in, got)
		}
	}
}
