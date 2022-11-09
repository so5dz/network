package client

type Client interface {
	Initialize(host string, port int)
	Connect() error
	Disconnect() error
	Send(data []byte) error
	OnReceive(callback func(data []byte))
}
