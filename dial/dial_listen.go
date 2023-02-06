package dial

import (
	"bytes"
	"fmt"
	"github.com/injoyai/io"
	"net"
	"sync"
)

func TCPListener(port int) (io.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &_tcpServer{Listener: listener}, nil
}

func TCPListenFunc(port int) io.ListenFunc {
	return func() (io.Listener, error) {
		return TCPListener(port)
	}
}

type _tcpServer struct {
	net.Listener
}

func (this *_tcpServer) Accept() (io.ReadWriteCloser, string, error) {
	c, err := this.Listener.Accept()
	return c, c.RemoteAddr().String(), err
}

func (this *_tcpServer) Addr() string {
	return this.Listener.Addr().String()
}

func UDPListener(port int) (io.Listener, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}
	return &_udpServer{UDPConn: listener}, nil
}

func UDPListenFunc(port int) io.ListenFunc {
	return func() (io.Listener, error) {
		return UDPListener(port)
	}
}

type _udp struct {
	s    *_udpServer
	addr *net.UDPAddr
	buff *bytes.Buffer
}

func (this *_udp) Read(p []byte) (int, error) {
	return this.buff.Read(p)
}

func (this *_udp) Write(p []byte) (int, error) {
	return this.s.WriteToUDP(p, this.addr)
}

func (this *_udp) Close() error {
	this.s.mu.Lock()
	defer this.s.mu.Unlock()
	delete(this.s.m, this.addr.String())
	return nil
}

// todo 待优化
type _udpServer struct {
	*net.UDPConn
	m  map[string]*_udp
	mu sync.Mutex
}

func (this *_udpServer) Accept() (io.ReadWriteCloser, string, error) {
	for {
		buff := make([]byte, 2<<10)
		n, addr, err := this.UDPConn.ReadFromUDP(buff)
		if err != nil {
			return nil, "", err
		}
		val, ok := this.m[addr.String()]
		if ok {
			val.buff.Write(buff[:n])
			continue
		}

		u := &_udp{
			addr: addr,
			buff: bytes.NewBuffer(buff[:n]),
		}

		this.mu.Lock()
		this.m[addr.String()] = u
		this.mu.Unlock()

		return u, u.addr.String(), nil
	}
}

func (this *_udpServer) Addr() string {
	return this.UDPConn.RemoteAddr().String()
}
