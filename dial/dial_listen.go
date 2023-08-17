package dial

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"net"
	"sync"
)

//================================TCPListen================================

func TCPListener(port int) (io.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &_tcpServer{Listener: listener}, nil
}

func TCPListenFunc(port int) io.ListenFunc {
	return func() (io.Listener, error) { return TCPListener(port) }
}

func NewTCPServer(port int, options ...io.OptionServer) (*io.Server, error) {
	s, err := io.NewServer(TCPListenFunc(port), options...)
	if err == nil {
		s.ClientManage.SetKey(fmt.Sprintf(":%d", port))
	}
	return s, err
}

func RunTCPServer(port int, options ...io.OptionServer) error {
	s, err := NewTCPServer(port, options...)
	if err != nil {
		return err
	}
	return s.Run()
}

func NewTCPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewTCPServer(port, func(s *io.Server) {
		s.SetOptions(options...)
		s.SetReadWithKB(4)
		s.SetBeforeFunc(func(client *io.Client) error {
			_, err := NewTCP(addr, func(c *io.Client) {
				c.Debug(false)
				c.SetReadWithKB(4)
				c.SetDealWithWriter(client)
				c.SetCloseFunc(func(ctx context.Context, msg *io.IMessage) {
					client.CloseWithErr(msg)
				})
				go c.Run()
				client.SetReadWithWriter(c)
			})
			return err
		})
	})
}

func RunTCPProxyServer(port int, addr string, options ...io.OptionServer) error {
	s, err := NewTCPProxyServer(port, addr, options...)
	if err != nil {
		return err
	}
	return s.Run()
}

type _tcpServer struct {
	net.Listener
}

func (this *_tcpServer) Accept() (io.ReadWriteCloser, string, error) {
	c, err := this.Listener.Accept()
	if err != nil {
		return nil, "", err
	}
	return c, c.RemoteAddr().String(), nil
}

func (this *_tcpServer) Addr() string {
	return this.Listener.Addr().String()
}

//================================UDPListen================================

func UDPListener(port int) (io.Listener, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}
	return &_udpServer{UDPConn: listener, m: make(map[string]*_udp)}, nil
}

func UDPListenFunc(port int) io.ListenFunc {
	return func() (io.Listener, error) {
		return UDPListener(port)
	}
}

func NewUDPServer(port int, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(UDPListenFunc(port), options...)
}

type _udp struct {
	s    *_udpServer
	addr *net.UDPAddr
	buff chan []byte
}

func (this *_udp) Read(p []byte) (int, error) {
	return 0, nil
}

func (this *_udp) ReadMessage() ([]byte, error) {
	return <-this.buff, nil
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
	mu sync.RWMutex
}

func (this *_udpServer) Accept() (io.ReadWriteCloser, string, error) {
	for {
		buff := make([]byte, 2<<10)
		n, addr, err := this.UDPConn.ReadFromUDP(buff)
		if err != nil {
			return nil, "", err
		}

		this.mu.RLock()
		val, ok := this.m[addr.String()]
		this.mu.RUnlock()
		if ok {
			select {
			case val.buff <- buff[:n]:
			default:
			}
			continue
		}

		u := &_udp{
			s:    this,
			addr: addr,
			buff: make(chan []byte, 100),
		}

		this.mu.Lock()
		this.m[addr.String()] = u
		this.mu.Unlock()

		u.buff <- buff[:n]

		return u, u.addr.String(), nil
	}
}

func (this *_udpServer) Addr() string {
	return this.UDPConn.RemoteAddr().String()
}
