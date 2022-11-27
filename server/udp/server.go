package udp

import base "github.com/so5dz/network/server"

type Server struct {
}

func (server *Server) Initialize(port int) {}

func (server *Server) Start() error {
	return nil
}

func (server *Server) Stop() error {
	return nil
}

func (server *Server) Broadcast(data []byte) error {
	return nil
}

func (server *Server) OnReceive(callback func(client base.Remote, data []byte)) {

}
