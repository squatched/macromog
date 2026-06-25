package main

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCapturedConfigPath(t *testing.T) {
	stdout, stderr, code := runCaptured([]string{"config", "path"}, "text")
	if code != 0 {
		t.Fatalf("code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "config") && stdout == "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
}

func TestAgentRoundTrip(t *testing.T) {
	portFile := filepath.Join(t.TempDir(), "agent.port")
	done := make(chan error, 1)
	go func() { done <- runAgent("127.0.0.1:0", portFile) }()

	var addr string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		data, err := os.ReadFile(portFile)
		if err == nil && len(data) > 0 {
			addr = strings.TrimSpace(string(data))
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if addr == "" {
		t.Fatal("agent did not write port file")
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial agent: %v", err)
	}
	defer conn.Close()

	req, _ := json.Marshal(agentRequest{Args: []string{"config", "path"}, Output: "text"})
	if _, err := conn.Write(append(req, '\n')); err != nil {
		t.Fatalf("write request: %v", err)
	}

	var resp agentResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("code = %d stderr = %q error = %q", resp.Code, resp.Stderr, resp.Error)
	}
	if resp.Stdout == "" {
		t.Fatal("expected stdout from agent")
	}
}
