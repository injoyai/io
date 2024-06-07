package tcp

import (
	"io"
	"net"
)

var _ io.ReadWriteCloser = (*TCP)(nil)

type TCP struct {
	net.Conn
}

func New(address string) (*TCP, error) {
	c, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TCP{
		Conn: c,
	}, nil
}
