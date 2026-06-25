package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"os"

	"github.com/spf13/cobra"
)

type agentRequest struct {
	Args   []string `json:"args"`
	Output string   `json:"output"`
}

type agentResponse struct {
	Code   int    `json:"code"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error,omitempty"`
}

func newAgentCmd() *cobra.Command {
	var portFile string
	var listenAddr string

	cmd := &cobra.Command{
		Use:    "agent",
		Hidden: true,
		Short:  "Run a local TCP agent for the Windower addon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgent(listenAddr, portFile)
		},
	}
	cmd.Flags().StringVar(&portFile, "portfile", "", "write bound address host:port to this file")
	cmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:0", "listen address")
	return cmd
}

func runAgent(listenAddr, portFile string) error {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	addr := ln.Addr().String()
	if portFile != "" {
		if err := os.WriteFile(portFile, []byte(addr+"\n"), 0o600); err != nil {
			return err
		}
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go handleAgentConn(conn)
	}
}

func handleAgentConn(conn net.Conn) {
	defer conn.Close()

	dec := json.NewDecoder(bufio.NewReader(conn))
	var req agentRequest
	if err := dec.Decode(&req); err != nil {
		writeAgentResponse(conn, agentResponse{Code: 1, Error: err.Error()})
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
