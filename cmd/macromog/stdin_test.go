package main

import "testing"

func TestReadStdinLine_Multiple(t *testing.T) {
	setInteractiveStdin(t, "\nother\n")

	l1, ok1 := readStdinLine()
	if !ok1 || l1 != "" {
		t.Fatalf("line1 = %q, ok=%v, want empty", l1, ok1)
	}
	l2, ok2 := readStdinLine()
	if !ok2 || l2 != "other" {
		t.Fatalf("line2 = %q, ok=%v, want other", l2, ok2)
	}
}
