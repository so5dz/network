package tcp

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/iskrapw/utils/misc"
)

const _ExpectedClients = 8
const _ProblemsClosingServerError = "there were problems closing tcp server"
const _ProblemsBroadcastingError = "there were problems broadcasting message"

type StreamServer struct {
	operate                    bool
	port                       int
	onReceive                  func(Remote, []byte)
	clients                    []*ConntectedClient
	hasUnhandledDisconnections bool
}

func (s *StreamServer) Initialize(port int) {
	s.port = port
}

func (s *StreamServer) Start() error {
	s.hasUnhandledDisconnections = false
	s.clients = make([]*ConntectedClient, 0, _ExpectedClients)
	s.operate = true

	listenPath := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen("tcp", listenPath)
	if err != nil {
		return misc.WrapError(_ListenError, err)
	}

	log.Println("Accepting connections on", listenPath)
	go s.start(listener)
	return nil
}

func (s *StreamServer) start(listener net.Listener) {
	for s.operate {
		tcpListener := listener.(*net.TCPListener)
		tcpListener.SetDeadline(time.Now().Add(_AcceptDeadline))

		socket, err := listener.Accept()
		if IsIOTimeoutError(err) {
			continue
		} else if err != nil {
			log.Println("Failed to accept incoming connection on", listener.Addr(), "-", err)
			continue
		}

		client := ConntectedClient{
			connected: true,
			socket:    socket,
		}
		s.clients = append(s.clients, &client)
		go s.readLoop(&client)
	}

	listener.Close()
}

func (s *StreamServer) Stop() error {
	s.operate = false
	errors := make([]error, 0)
	for _, c := range s.clients {
		err := c.Disconnect()
		if err != nil {
			errors = append(errors, err)
		}
	}
	s.clients = s.clients[0:0]
	return misc.WrapMultiple(_ProblemsClosingServerError, errors)
}

func (s *StreamServer) Broadcast(data []byte) error {
	errors := make([]error, 0)
	for _, c := range s.clients {
		_, err := c.socket.Write(data)
		if err != nil {
			log.Println("Disconnecting", c.Address(), "due to a write error", err)
			err := c.Disconnect()
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	s.removeDisconnectedClients()
	return misc.WrapMultiple(_ProblemsBroadcastingError, errors)
}

func (s *StreamServer) OnReceive(callback func(Remote, []byte)) {
	s.onReceive = callback
}

func (s *StreamServer) readLoop(c *ConntectedClient) {
	log.Println(c.Address(), "conntected")

	buf := make([]byte, _ReadBufferSize)
	for s.operate && c.connected {
		c.socket.SetReadDeadline(time.Now().Add(time.Second / 2))
		n, err := c.socket.Read(buf)
		if IsIOTimeoutError(err) {
			continue
		} else if err != nil {
			log.Println("Disconnecting", c.Address(), "due to a read error", err)
			c.Disconnect()
			return
		} else if (n > 0) && (s.onReceive != nil) {
			s.onReceive(c, buf[0:n])
		}
	}
}

func (s *StreamServer) removeDisconnectedClients() {
	connectedClients := make([]*ConntectedClient, 0, len(s.clients))
	for _, c := range s.clients {
		if c.connected {
			connectedClients = append(connectedClients, c)
		} else {
			log.Println("Client", c.Address(), "removed")
		}
	}
	s.clients = connectedClients
}
