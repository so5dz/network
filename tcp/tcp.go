package tcp

type Client interface {
	Initialize(host string, port int)
	Connect() error
	Disconnect() error
	Send(data []byte) error
	OnReceive(callback func(data []byte))
}

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

func NewClient(host string, port int, mode TCPConnectionMode) Client {
	var client Client
	if mode == TCPConnectionMode_Stream {
		client = &StreamClient{}
	} else if mode == TCPConnectionMode_Message {
		client = &MessageClient{}
	}
	client.Initialize(host, port)
	return client
}

func NewServer(port int, mode TCPConnectionMode) Server {
	var server Server
	if mode == TCPConnectionMode_Stream {
		server = &StreamServer{}
	} else if mode == TCPConnectionMode_Message {
		server = &MessageServer{}
	}
	server.Initialize(port)
	return server
}
