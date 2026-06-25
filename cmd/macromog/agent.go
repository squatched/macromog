package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type agentRequest struct {
	Args     []string `json:"args"`
	Output   string   `json:"output"`
	Shutdown bool     `json:"shutdown,omitempty"`
}

type agentResponse struct {
	Code   int    `json:"code"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error,omitempty"`
}

type agentServer struct {
	ln          net.Listener
	portFile    string
	idleTimeout time.Duration
	mu          sync.Mutex
	lastActive  time.Time
	stopCh      chan struct{}
	stopOnce    sync.Once
}

func newAgentCmd() *cobra.Command {
	var portFile string
	var listenAddr string
	var idleTimeout time.Duration

	cmd := &cobra.Command{
		Use:    "agent",
		Hidden: true,
		Short:  "Run a local TCP agent for the Windower addon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgent(listenAddr, portFile, idleTimeout)
		},
	}
	cmd.Flags().StringVar(&portFile, "portfile", "", "write bound address host:port to this file")
	cmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:0", "loopback listen address (127.0.0.1 only)")
	cmd.Flags().DurationVar(&idleTimeout, "idle-timeout", 2*time.Minute, "exit after this long without requests")
	return cmd
}

func validateListenAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("listen address: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("agent must bind to loopback only, got %q", host)
	}
	return nil
}

func isLoopbackRemote(addr net.Addr) bool {
	if addr == nil {
		return false
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func runAgent(listenAddr, portFile string, idleTimeout time.Duration) error {
	if err := validateListenAddr(listenAddr); err != nil {
		return err
	}
	if idleTimeout <= 0 {
		idleTimeout = 2 * time.Minute
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	srv := &agentServer{
		ln:          ln,
		portFile:    portFile,
		idleTimeout: idleTimeout,
		lastActive:  time.Now(),
		stopCh:      make(chan struct{}),
	}
	defer srv.cleanup()

	addr := ln.Addr().String()
	if portFile != "" {
		if err := os.WriteFile(portFile, []byte(addr+"\n"), 0o600); err != nil {
			return err
		}
	}

	go srv.watchIdle()

	for {
		if tcp, ok := ln.(*net.TCPListener); ok {
			_ = tcp.SetDeadline(time.Now().Add(time.Second))
		}
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-srv.stopCh:
				return nil
			default:
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				return err
			}
		}
		srv.touch()
		go srv.handleConn(conn)
	}
}

func (s *agentServer) touch() {
	s.mu.Lock()
	s.lastActive = time.Now()
	s.mu.Unlock()
}

func (s *agentServer) idleExpired() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.lastActive) >= s.idleTimeout
}

func (s *agentServer) watchIdle() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if s.idleExpired() {
				s.requestStop()
				return
			}
		}
	}
}

func (s *agentServer) requestStop() {
	s.stopOnce.Do(func() {
		s.cleanup()
		close(s.stopCh)
	})
}

func (s *agentServer) cleanup() {
	if s.portFile != "" {
		_ = os.Remove(s.portFile)
	}
	if s.ln != nil {
		_ = s.ln.Close()
		s.ln = nil
	}
}

func (s *agentServer) handleConn(conn net.Conn) {
	defer conn.Close()

	if !isLoopbackRemote(conn.RemoteAddr()) {
		return
	}

	dec := json.NewDecoder(bufio.NewReader(conn))
	var req agentRequest
	if err := dec.Decode(&req); err != nil {
		writeAgentResponse(conn, agentResponse{Code: 1, Error: err.Error()})
		return
	}

	if req.Shutdown {
		writeAgentResponse(conn, agentResponse{Code: 0})
		s.requestStop()
		return
	}

	stdout, stderr, code := runCaptured(req.Args, req.Output)
	writeAgentResponse(conn, agentResponse{
		Code:   code,
		Stdout: stdout,
		Stderr: stderr,
	})
}

func writeAgentResponse(w io.Writer, resp agentResponse) {
	enc := json.NewEncoder(w)
	_ = enc.Encode(resp)
}

func runCaptured(argv []string, output string) (stdout, stderr string, code int) {
	if output == "" {
		output = "text"
	}

	var outBuf, errBuf bytes.Buffer

	root, state := newRootCmd()
	state.format = output
	state.out = &outBuf
	state.printer = NewPrinter(&outBuf, OutputFormat(output))

	errR, errW, err := os.Pipe()
	if err != nil {
		return "", err.Error(), 1
	}
	oldErr := os.Stderr
	os.Stderr = errW

	errDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&errBuf, errR)
		close(errDone)
	}()

	root.SetArgs(argv)
	if execErr := root.Execute(); execErr != nil {
		if state.code == 0 {
			state.code = 1
		}
	}

	_ = errW.Close()
	<-errDone
	os.Stderr = oldErr

	return outBuf.String(), errBuf.String(), state.code
}
