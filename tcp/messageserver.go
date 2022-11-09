package tcp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	utils "github.com/iskrapw/utils/src"
)

type MessageServer struct {
	operate                    bool
	port                       int
	onReceive                  func(Remote, []byte)
	clients                    []*ConntectedClient
	hasUnhandledDisconnections bool
}

func (s *MessageServer) Initialize(port int) {
	s.port = port
}

func (s *MessageServer) Start() error {
	s.hasUnhandledDisconnections = false
	s.clients = make([]*ConntectedClient, 0, _ExpectedClients)
	s.operate = true

	listenPath := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen("tcp", listenPath)
	if err != nil {
		return utils.WrapError(_ListenError, err)
	}

	log.Println("Accepting connections on", listenPath)
	go s.start(listener)
	return nil
}

func (s *MessageServer) start(listener net.Listener) {
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

func (s *MessageServer) Stop() error {
	s.operate = false
	errors := make([]error, 0)
	for _, c := range s.clients {
		err := c.Disconnect()
		if err != nil {
			errors = append(errors, err)
		}
	}
	s.clients = s.clients[0:0]
	return utils.WrapMultiple(_ProblemsClosingServerError, errors)
}

func (s *MessageServer) Broadcast(data []byte) error {
	errors := make([]error, 0)
	for _, c := range s.clients {
		errHeader := sendHeader(c, len(data))
		if errHeader != nil {
			errors = append(errors, errHeader)
		}

		errData := c.Send(data)
		if errData != nil {
			errors = append(errors, errData)
		}

		if errHeader != nil || errData != nil {
			log.Println("Disconnecting", c.Address(), "due to a write error")
			err := c.Disconnect()
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	s.removeDisconnectedClients()
	return utils.WrapMultiple(_ProblemsBroadcastingError, errors)
}

func (s *MessageServer) OnReceive(callback func(Remote, []byte)) {
	s.onReceive = callback
}

func (server *MessageServer) readLoop(client *ConntectedClient) error {
	for server.operate && client.connected {
		dataLength, err := server.readHeader(client)
		if err != nil {
			return err
		}

		data, err := server.readExactNumberOfBytes(client, dataLength)
		if err != nil {
			return err
		}

		if server.onReceive != nil {
			server.onReceive(client, data)
		}
	}

	return utils.NewError(_OperationInterrupted)
}

func sendHeader(client *ConntectedClient, dataLength int) error {
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(dataLength))
	return client.Send(header[:])
}

func (server *MessageServer) readHeader(client *ConntectedClient) (int, error) {
	headerBuffer, err := server.readExactNumberOfBytes(client, _MessageHeaderSize)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(headerBuffer[0:_MessageHeaderSize])), nil
}

func (server *MessageServer) readExactNumberOfBytes(client *ConntectedClient, bytesToRead int) ([]byte, error) {
	buffer := make([]byte, bytesToRead)
	totalBytesRead := 0

	for totalBytesRead < bytesToRead {
		client.socket.SetReadDeadline(time.Now().Add(_ReadDeadline))
		bytesRead, err := client.socket.Read(buffer[totalBytesRead:bytesToRead])
		totalBytesRead += bytesRead

		if err != nil && !IsIOTimeoutError(err) {
			return buffer, err
		}

		if !server.operate {
			return buffer, utils.NewError(_OperationInterrupted)
		}
	}

	return buffer, nil
}

func (s *MessageServer) removeDisconnectedClients() {
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
