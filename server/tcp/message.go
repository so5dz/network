package tcp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/iskrapw/network/common"
	"github.com/iskrapw/network/server"
	"github.com/iskrapw/utils/misc"
)

type MessageServer struct {
	operate                    bool
	port                       int
	onReceive                  func(server.Remote, []byte)
	clients                    []*RemoteClient
	hasUnhandledDisconnections bool
}

func (s *MessageServer) Initialize(port int) {
	s.port = port
}

func (s *MessageServer) Start() error {
	s.hasUnhandledDisconnections = false
	s.clients = make([]*RemoteClient, 0, _ExpectedClients)
	s.operate = true

	listenPath := fmt.Sprintf("0.0.0.0:%d", s.port)
	listener, err := net.Listen("tcp", listenPath)
	if err != nil {
		return misc.WrapError(common.ListenError, err)
	}

	log.Println("Accepting connections on", listenPath)
	go s.start(listener)
	return nil
}

func (s *MessageServer) start(listener net.Listener) {
	for s.operate {
		tcpListener := listener.(*net.TCPListener)
		tcpListener.SetDeadline(time.Now().Add(common.AcceptDeadline))

		socket, err := listener.Accept()
		if common.IsIOTimeoutError(err) {
			continue
		} else if err != nil {
			log.Println("Failed to accept incoming connection on", listener.Addr(), "-", err)
			continue
		}

		client := RemoteClient{
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
	return misc.WrapMultiple(_ProblemsClosingServerError, errors)
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
	return misc.WrapMultiple(_ProblemsBroadcastingError, errors)
}

func (s *MessageServer) OnReceive(callback func(server.Remote, []byte)) {
	s.onReceive = callback
}

func (server *MessageServer) readLoop(client *RemoteClient) error {
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

	return misc.NewError(common.OperationInterrupted)
}

func sendHeader(client *RemoteClient, dataLength int) error {
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(dataLength))
	return client.Send(header[:])
}

func (server *MessageServer) readHeader(client *RemoteClient) (int, error) {
	headerBuffer, err := server.readExactNumberOfBytes(client, common.MessageHeaderSize)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(headerBuffer[0:common.MessageHeaderSize])), nil
}

func (server *MessageServer) readExactNumberOfBytes(client *RemoteClient, bytesToRead int) ([]byte, error) {
	buffer := make([]byte, bytesToRead)
	totalBytesRead := 0

	for totalBytesRead < bytesToRead {
		client.socket.SetReadDeadline(time.Now().Add(common.ReadDeadline))
		bytesRead, err := client.socket.Read(buffer[totalBytesRead:bytesToRead])
		totalBytesRead += bytesRead

		if err != nil && !common.IsIOTimeoutError(err) {
			return buffer, err
		}

		if !server.operate {
			return buffer, misc.NewError(common.OperationInterrupted)
		}
	}

	return buffer, nil
}

func (s *MessageServer) removeDisconnectedClients() {
	connectedClients := make([]*RemoteClient, 0, len(s.clients))
	for _, c := range s.clients {
		if c.connected {
			connectedClients = append(connectedClients, c)
		} else {
			log.Println("Client", c.Address(), "removed")
		}
	}
	s.clients = connectedClients
}
