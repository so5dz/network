package tcp

import "net"

type RemoteClient struct {
	connected bool
	socket    net.Conn
}

func (rc *RemoteClient) Address() string {
	return rc.socket.RemoteAddr().String()
}

func (rc *RemoteClient) Disconnect() error {
	rc.connected = false
	return rc.socket.Close()
}

func (rc *RemoteClient) Send(data []byte) error {
	_, err := rc.socket.Write(data)
	return err
}
