package config

import "testing"

func TestOpenPath_AlwaysMapsHomeUnderWine(t *testing.T) {
	// Simulate the Windows/Wine branch without a Windows runner.
	canonical := "/home/squatched/.config/macromog/config.yml"
	got, err := resolveForWine(canonical)
	if err != nil {
		t.Fatal(err)
	}
	want := `Z:\home\squatched\.config\macromog\config.yml`
	if got != want {
		t.Errorf("resolveForWine() = %q, want %q", got, want)
	}
}