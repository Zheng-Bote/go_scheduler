package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// StatusEvent represents the JSON payload sent by jobs
type StatusEvent struct {
	RunID    int    `json:"run_id"`
	Type     string `json:"type"` // "status" (default) or "audit"
	Status   string `json:"status"`
	Message  string `json:"message"`
	Progress int    `json:"progress"`
}

// Server listens for StatusEvents on a Unix Domain Socket
type Server struct {
	SocketPath string
	OnEvent    func(event StatusEvent)
}

// Start runs the Unix Domain Socket server
func (s *Server) Start() error {
	if _, err := os.Stat(s.SocketPath); err == nil {
		_ = os.Remove(s.SocketPath)
	}

	l, err := net.Listen("unix", s.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	// Set permissions for the socket
	_ = os.Chmod(s.SocketPath, 0666)

	go func() {
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go s.handleConnection(conn)
		}
	}()

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var event StatusEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			if s.OnEvent != nil {
				s.OnEvent(event)
			}
		}
	}
}

// Client is used by jobs to send events to the scheduler
type Client struct {
	SocketPath string
}

// SendEvent sends a StatusEvent to the scheduler socket
func (c *Client) SendEvent(event StatusEvent) error {
	conn, err := net.Dial("unix", c.SocketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = conn.Write(append(data, '\n'))
	return err
}
