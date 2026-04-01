package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// StdioTransport implements stdio transport for MCP
type StdioTransport struct {
	config   *ServerConfig
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.Reader
	mu       sync.Mutex
	closed   bool
	readMu   sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(config *ServerConfig) *StdioTransport {
	return &StdioTransport{
		config: config,
	}
}

// Name returns the transport name
func (t *StdioTransport) Name() string {
	return "stdio"
}

// Connect establishes the stdio connection
func (t *StdioTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	// Prepare command
	cmd := exec.Command(t.config.Command, t.config.Args...)

	// Set environment
	if t.config.Env != nil {
		for k, v := range t.config.Env {
			os.Setenv(k, v)
		}
		cmd.Env = os.Environ()
	}

	// Connect stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	t.stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	t.stdout = stdout

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	t.cmd = cmd
	return nil
}

// Send sends a JSON-RPC message
func (t *StdioTransport) Send(ctx context.Context, msg *JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("transport is closed")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write as JSON line
	_, err = t.stdin.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Receive receives a JSON-RPC message
func (t *StdioTransport) Receive(ctx context.Context) (*JSONRPCMessage, error) {
	t.readMu.Lock()
	defer t.readMu.Unlock()

	if t.closed {
		return nil, fmt.Errorf("transport is closed")
	}

	// Read one line (JSON object)
	reader := bufio.NewReader(t.stdout)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil, err
			}
			return nil, fmt.Errorf("failed to read message: %w", err)
		}

		// Skip empty lines
		line = []byte(strings.TrimSpace(string(line)))
		if len(line) == 0 {
			continue
		}

		// Parse JSON-RPC message
		var msg JSONRPCMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			// Skip invalid messages
			continue
		}

		return &msg, nil
	}
}

// Close closes the connection
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true

	if t.stdin != nil {
		t.stdin.Close()
	}

	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}

	return nil
}

// IsConnected returns true if the transport is connected
func (t *StdioTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.closed && t.cmd != nil && t.cmd.ProcessState == nil
}

func init() {
	// Register stdio transport
	RegisterTransport(TransportStdio, func(config *ServerConfig) Transport {
		return NewStdioTransport(config)
	})
}
