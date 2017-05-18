package input

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"

	metrics "github.com/rcrowley/go-metrics"
)

const (
	newlineTimeout = time.Duration(1000 * time.Millisecond)
	msgBufSize     = 256
)

// A server captures attributes common to all servers.
type server struct {
	iface    string
	registry metrics.Registry
	eventsRx metrics.Counter
	bytesRx  metrics.Counter
}

// Statistics returns an object storing statistics, which supports JSON
// marshalling.
func (s *server) Statistics() (metrics.Registry, error) {
	return s.registry, nil
}

// A TcpServer binds to the supplied interface and receives Syslog messages.
type TcpServer struct {
	server
	connectionsActive metrics.Counter
}

// NewTcpServer returns a TCP server.
func NewTcpServer(iface string) *TcpServer {
	s := &TcpServer{}
	s.iface = iface

	s.registry = metrics.NewRegistry()
	s.eventsRx = metrics.NewCounter()
	s.bytesRx = metrics.NewCounter()
	s.connectionsActive = metrics.NewCounter()
	s.registry.Register("events.received", s.eventsRx)
	s.registry.Register("events.bytes.received", s.bytesRx)
	s.registry.Register("connections.Active", s.connectionsActive)

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
				log.Println("failed to accept connection", err)
				continue
			}
			log.Println("accepted new connection from", conn.RemoteAddr().String())
			go s.handleConnection(conn, f)
		}
	}()
	return nil
}

func (s *TcpServer) handleConnection(conn net.Conn, f func() chan<- string) {
	s.connectionsActive.Inc(1)
	defer conn.Close()
	defer s.connectionsActive.Dec(1)

	delimiter := NewDelimiter(msgBufSize)
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
				log.Println("Error from connection:", err)
				return
			}
		} else {
			event, match = delimiter.Push(b)
		}
		if match {
			s.eventsRx.Inc(1)
			s.bytesRx.Inc(int64(len(event)))
			f() <- event
		}
	}
}

// A UdpServer listens to the supplied interface and receives Syslog messages.
type UdpServer struct {
	server
	udpAddr *net.UDPAddr
}

// NewUdpServer returns a UDP server.
func NewUdpServer(iface string) *UdpServer {
	addr, err := net.ResolveUDPAddr("udp", iface)
	if err != nil {
		return nil
	}

	s := &UdpServer{}
	s.iface = iface
	s.udpAddr = addr

	s.registry = metrics.NewRegistry()
	s.eventsRx = metrics.NewCounter()
	s.bytesRx = metrics.NewCounter()
	s.registry.Register("events.received", s.eventsRx)
	s.registry.Register("events.bytes.received", s.bytesRx)

	return s
}

// Start instructs the UdpServer to start reading packets from the interface.
func (s *UdpServer) Start(f func() chan<- string) error {
	conn, err := net.ListenUDP("udp", s.udpAddr)
	if err != nil {
		log.Println("failed to start UDP server", err)
		return err
	}

	go func() {
		buf := make([]byte, msgBufSize)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("failed to read UDP", err)
			}
			s.eventsRx.Inc(1)
			s.bytesRx.Inc(int64(len(buf)))
			f() <- strings.Trim(string(buf[:n]), "\r\n")
		}
	}()
	return nil
}
