package tcp

import (
	"fmt"
	"log"
	"net"
	"time"

	utils "github.com/iskrapw/utils/src"
)

type StreamClient struct {
	host      string
	port      int
	onReceive func([]byte)
	socket    net.Conn
	operate   bool
}

func (c *StreamClient) Initialize(host string, port int) {
	c.host = host
	c.port = port
}

func (c *StreamClient) Connect() error {
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

func (c *StreamClient) Disconnect() error {
	c.operate = false
	return c.socket.Close()
}

func (c *StreamClient) OnReceive(callback func([]byte)) {
	c.onReceive = callback
}

func (c *StreamClient) Send(data []byte) error {
	_, err := c.socket.Write(data)
	return err
}

func (c *StreamClient) readLoop() {
	buf := make([]byte, _ReadBufferSize)
	for c.operate {
		c.socket.SetReadDeadline(time.Now().Add(_ReadDeadline))
		n, err := c.socket.Read(buf)
		if IsIOTimeoutError(err) {
			return
		} else if err != nil {
			log.Println("Disconnecting from server due to a read error", err)
			return
		} else if (n > 0) && (c.onReceive != nil) {
			c.onReceive(buf[0:n])
		}
	}
}
