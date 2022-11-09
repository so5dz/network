package tcp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	utils "github.com/iskrapw/utils/src"
)

type MessageClient struct {
	host      string
	port      int
	onReceive func([]byte)
	socket    net.Conn
	operate   bool
	toRead    int
}

func (c *MessageClient) Initialize(host string, port int) {
	c.host = host
	c.port = port
	c.operate = false
	c.toRead = 0
}

func (c *MessageClient) Connect() error {
	connectPath := fmt.Sprintf("%s:%d", c.host, c.port)

	log.Println("Connecting to", connectPath)

	var err error
	c.socket, err = net.Dial(_TCPNetworkType, connectPath)
	if err != nil {
		return utils.WrapError(_DialError, err)
	}

	log.Println("Connected")

	c.operate = true
	go c.readLoop()
	return nil
}

func (c *MessageClient) Disconnect() error {
	c.operate = false
	return c.socket.Close()
}

func (c *MessageClient) OnReceive(callback func([]byte)) {
	c.onReceive = callback
}

func (c *MessageClient) Send(data []byte) error {
	err := c.sendHeader(len(data))
	if err != nil {
		return err
	}
	_, err = c.socket.Write(data)
	return err
}

func (c *MessageClient) readLoop() error {
	for c.operate {
		dataLength, err := c.readHeader()
		if err != nil {
			return err
		}

		data, err := c.readExactNumberOfBytes(dataLength)
		if err != nil {
			return err
		}

		if c.onReceive != nil {
			c.onReceive(data)
		}
	}

	return utils.NewError(_OperationInterrupted)
}

func (c *MessageClient) sendHeader(dataLength int) error {
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(dataLength))
	_, err := c.socket.Write(header[:4])
	return err
}

func (c *MessageClient) readHeader() (int, error) {
	headerBuffer, err := c.readExactNumberOfBytes(_MessageHeaderSize)
	if err != nil {
		return 0, err
	}
	return int(binary.BigEndian.Uint32(headerBuffer[:])), nil
}

func (c *MessageClient) readExactNumberOfBytes(bytesToRead int) ([]byte, error) {
	buffer := make([]byte, 0, bytesToRead)
	totalBytesRead := 0

	for totalBytesRead < bytesToRead {
		c.socket.SetReadDeadline(time.Now().Add(_ReadDeadline))
		bytesRead, err := c.socket.Read(buffer[:])
		if err != nil && !IsIOTimeoutError(err) {
			return buffer, err
		} else if (bytesRead > 0) && (c.onReceive != nil) {
			totalBytesRead += bytesRead
		}
		if !c.operate {
			return buffer, utils.NewError(_OperationInterrupted)
		}
	}

	return buffer, nil
}
