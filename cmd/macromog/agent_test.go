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

func TestValidateListenAddrRejectsWildcard(t *testing.T) {
	if err := validateListenAddr("0.0.0.0:0"); err == nil {
		t.Fatal("expected wildcard bind to be rejected")
	}
	if err := validateListenAddr("127.0.0.1:0"); err != nil {
		t.Fatalf("loopback bind should be allowed: %v", err)
	}
}

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
	go func() { _ = runAgent("127.0.0.1:0", portFile, time.Minute) }()

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

func TestAgentShutdownRemovesPortFile(t *testing.T) {
	portFile := filepath.Join(t.TempDir(), "agent.port")
	go func() { _ = runAgent("127.0.0.1:0", portFile, time.Minute) }()

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
	req, _ := json.Marshal(agentRequest{Shutdown: true})
	if _, err := conn.Write(append(req, '\n')); err != nil {
		t.Fatalf("write shutdown: %v", err)
	}
	var resp agentResponse
	_ = json.NewDecoder(conn).Decode(&resp)
	conn.Close()

	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(portFile); os.IsNotExist(err) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("port file still present after shutdown")
}
