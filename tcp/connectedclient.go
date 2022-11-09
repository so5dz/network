package tcp

import "net"

type ConntectedClient struct {
	connected bool
	socket    net.Conn
}

func (cc *ConntectedClient) Address() string {
	return cc.socket.RemoteAddr().String()
}

func (cc *ConntectedClient) Disconnect() error {
	cc.connected = false
	return cc.socket.Close()
}

func (cc *ConntectedClient) Send(data []byte) error {
	_, err := cc.socket.Write(data)
	return err
}
