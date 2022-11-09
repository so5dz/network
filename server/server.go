package server

type Remote interface {
	Address() string
}

type Server interface {
	Initialize(port int)
	Start() error
	Stop() error
	Broadcast(data []byte) error
	OnReceive(callback func(client Remote, data []byte))
}
