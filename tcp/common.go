package tcp

import (
	"strings"
	"time"
)

type TCPConnectionMode int

const (
	TCPConnectionMode_Stream  TCPConnectionMode = 0
	TCPConnectionMode_Message TCPConnectionMode = 1
)

const _TCPNetworkType = "tcp"
const _MessageHeaderSize = 4
const _ReadBufferSize = 512
const _ReadDeadline = time.Second / 2
const _AcceptDeadline = time.Second / 2

func IsIOTimeoutError(e error) bool {
	return (e != nil) && strings.HasSuffix(e.Error(), "i/o timeout")
}

const _DialError = "unable to dial TCP server"
const _ListenError = "unable to start TCP listening"
const _OperationInterrupted = "operation interrupted"
