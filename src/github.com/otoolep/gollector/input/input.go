package input

import (
	"bufio"
	"net"
	"time"

	log "code.google.com/p/log4go"
)

const (
	newlineTimeout = time.Duration(1000 * time.Millisecond)
)

// A TcpServer binds to the supplied interface and receives Syslog messages.
type TcpServer struct {
	iface string
}

// NewTcpServer returns a TCP server.
func NewTcpServer(iface string) *TcpServer {
	s := &TcpServer{iface}
	return s
}

// Start instructs the TcpServer to bind to the interface and accept connections.
func (s *TcpServer) Start(f func() chan<- string) error {
	ln, err := net.Listen("tcp", s.iface)
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Error("failed to accept connection", err)
				continue
			}
			log.Info("accepted new connection from %s", conn.RemoteAddr().String())
			go s.handleConnection(conn, f)
		}
	}()
	return nil
}

func (s *TcpServer) handleConnection(conn net.Conn, f func() chan<- string) {
	defer conn.Close()
	delimiter := NewDelimiter(256)
	reader := bufio.NewReader(conn)
	var event string
	var match bool

	for {
		conn.SetReadDeadline(time.Now().Add(newlineTimeout))
		b, err := reader.ReadByte()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				event, match = delimiter.Vestige()
			} else {
				log.Info("Error from connection:", err)
				return
			}
		} else {
			event, match = delimiter.Push(b)
		}
		if match {
			f() <- event
		}
	}
}
