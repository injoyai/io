package dial

import (
	"fmt"
	"github.com/injoyai/io"
	"net"
)

func TCPListener(port int) (io.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &_tcp{Listener: listener}, nil
}

type _tcp struct {
	net.Listener
}

func (this *_tcp) Accept() (io.ReadWriteCloser, string, error) {
	c, err := this.Listener.Accept()
	return c, c.RemoteAddr().String(), err
}

func (this *_tcp) Addr() string {
	return this.Listener.Addr().String()
}
