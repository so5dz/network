package common

import (
	"strings"
	"time"
)

type TCPConnectionMode int

const (
	TCPConnectionMode_Stream  TCPConnectionMode = 0
	TCPConnectionMode_Message TCPConnectionMode = 1
)

const TCPNetworkType = "tcp"
const MessageHeaderSize = 4
const ReadBufferSize = 512
const ReadDeadline = time.Second / 2
const AcceptDeadline = time.Second / 2

func IsIOTimeoutError(e error) bool {
	return (e != nil) && strings.HasSuffix(e.Error(), "i/o timeout")
}

const DialError = "unable to dial TCP server"
const ListenError = "unable to start TCP listening"
const OperationInterrupted = "operation interrupted"
