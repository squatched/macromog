package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func testCLIState(t *testing.T, format string) (*cliState, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	return &cliState{
		out:     buf,
		format:  format,
		printer: NewPrinter(buf, OutputFormat(format)),
	}, buf
}

func TestDebugEnv_IncludesPath(t *testing.T) {
	t.Setenv("MACROMOG_TEST_MARKER", "present")
	state, buf := testCLIState(t, "text")
	if err := emitDebugEnv(state); err != nil {
		t.Fatalf("emitDebugEnv() = %v", err)
	}
	if !strings.Contains(buf.String(), "MACROMOG_TEST_MARKER=present") {
		t.Fatalf("output missing test env var:\n%s", buf.String())
	}
}

func TestDebugPaths_JSON(t *testing.T) {
	state, buf := testCLIState(t, "json")
	if err := emitDebugPaths(state); err != nil {
		t.Fatalf("emitDebugPaths() = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"goos"`) || !strings.Contains(out, `"config_path"`) {
		t.Fatalf("unexpected json output: %s", out)
	}
}

func TestDebugAll_Text(t *testing.T) {
	state, buf := testCLIState(t, "text")
	cmd := newDebugAllCmd(state)
	if got := execWithState(cmd, nil, state); got != 0 {
		t.Fatalf("debug all = %d, want 0", got)
	}
	out := buf.String()
	if !strings.Contains(out, "environment:") || !strings.Contains(out, "paths:") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestDebugEnv_Sorted(t *testing.T) {
	t.Setenv("ZZZ_MACROMOG_SORT_TEST", "1")
	t.Setenv("AAA_MACROMOG_SORT_TEST", "1")
	state, buf := testCLIState(t, "text")
	if err := emitDebugEnv(state); err != nil {
		t.Fatalf("emitDebugEnv() = %v", err)
	}
	out := buf.String()
	aaa := strings.Index(out, "AAA_MACROMOG_SORT_TEST=1")
	zzz := strings.Index(out, "ZZZ_MACROMOG_SORT_TEST=1")
	if aaa < 0 || zzz < 0 || aaa > zzz {
		t.Fatalf("environment not sorted:\n%s", out)
	}
	_ = os.Unsetenv("ZZZ_MACROMOG_SORT_TEST")
	_ = os.Unsetenv("AAA_MACROMOG_SORT_TEST")
}

func TestRun_DebugEnv(t *testing.T) {
	t.Setenv("MACROMOG_TEST_MARKER", "present")
	if got := run([]string{"macromog", "debug", "env"}); got != 0 {
		t.Fatalf("run(debug env) = %d, want 0", got)
	}
}
