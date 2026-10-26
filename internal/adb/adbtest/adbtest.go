// Package adbtest provides an in-process fake adb server for tests.
package adbtest

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
)

// Server answers host:version, host:devices-l, host:transport:<serial>, and
// exec: package installs and Activity starts for the serial "emulator-5554".
type Server struct {
	ln net.Listener

	mu        sync.Mutex
	Devices   string
	Reply     string
	installed []byte
	commands  []string
}

const Serial = "emulator-5554"

func New() (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Server{ln: ln, Devices: Serial + "          device product:sdk model:sdk_gphone_x86_64 device:generic transport_id:1\n", Reply: "Success\n"}
	go s.serve()
	return s, nil
}

func (s *Server) Addr() string { return s.ln.Addr().String() }
func (s *Server) Close() error { return s.ln.Close() }

// SetReply sets the package manager's answer to the next installs.
func (s *Server) SetReply(reply string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Reply = reply
}

func (s *Server) Installed() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.installed
}

func (s *Server) Commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.commands...)
}

func readRequest(conn net.Conn) (string, error) {
	size := make([]byte, 4)
	if _, err := io.ReadFull(conn, size); err != nil {
		return "", err
	}
	n, err := strconv.ParseUint(string(size), 16, 16)
	if err != nil {
		return "", err
	}
	b := make([]byte, n)
	_, err = io.ReadFull(conn, b)
	return string(b), err
}

func block(s string) string { return fmt.Sprintf("%04x%s", len(s), s) }

func (s *Server) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	transport := ""
	for {
		req, err := readRequest(conn)
		if err != nil {
			return
		}
		s.mu.Lock()
		s.commands = append(s.commands, req)
		devices, reply := s.Devices, s.Reply
		s.mu.Unlock()
		const install = "exec:cmd package install -r -S "
		switch {
		case req == "host:version":
			io.WriteString(conn, "OKAY"+block("0029"))
			return
		case req == "host:devices-l":
			io.WriteString(conn, "OKAY"+block(devices))
			return
		case strings.HasPrefix(req, "host:transport:"):
			transport = strings.TrimPrefix(req, "host:transport:")
			if transport != Serial {
				io.WriteString(conn, "FAIL"+block("device '"+transport+"' not found"))
				return
			}
			io.WriteString(conn, "OKAY")
		case transport != "" && strings.HasPrefix(req, install):
			n, err := strconv.Atoi(strings.TrimPrefix(req, install))
			if err != nil || n < 0 || n > 512<<20 {
				io.WriteString(conn, "FAIL"+block("bad size"))
				return
			}
			io.WriteString(conn, "OKAY")
			b := make([]byte, n)
			if _, err = io.ReadFull(conn, b); err != nil {
				return
			}
			s.mu.Lock()
			s.installed = b
			s.mu.Unlock()
			io.WriteString(conn, reply)
			return
		case transport != "" && strings.HasPrefix(req, "exec:am start -W -n "):
			io.WriteString(conn, "OKAY")
			if strings.HasSuffix(req, "/com.koru.runtime.MainActivity") {
				io.WriteString(conn, "Starting: Intent { cmp=x }\nStatus: ok\nLaunchState: COLD\n")
			} else {
				io.WriteString(conn, "Error type 3\nError: Activity class does not exist.\n")
			}
			return
		default:
			io.WriteString(conn, "FAIL"+block("unknown service "+req))
			return
		}
	}
}
