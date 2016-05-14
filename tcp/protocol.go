package tcp

import (
	"net"
)

type Packet interface {
	Serialize() []byte
}

type Middleware interface {
	Load() []byte
}

type Protocol interface {
	ReadPacket(conn *net.TCPConn) (Packet, error)
	Middleware(s *Server)
}
