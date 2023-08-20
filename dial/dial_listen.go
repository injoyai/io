package dial

import (
	"bytes"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"net"
	"sync"
	"time"
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
	return io.NewServer(TCPListenFunc(port), func(s *io.Server) {
		s.SetOptions(options...)
		s.SetKey(fmt.Sprintf(":%d", port))
	})
}

func RunTCPServer(port int, options ...io.OptionServer) error {
	return RunServer(NewTCPServer(port, options...))
}

func NewTCPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewProxyServer(TCPListenFunc(port), TCPFunc(addr), options...)
}

func RunTCPProxyServer(port int, addr string, options ...io.OptionServer) error {
	return RunProxyServer(TCPListenFunc(port), TCPFunc(addr), options...)
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
	return io.NewServer(UDPListenFunc(port), func(s *io.Server) {
		s.SetOptions(options...)
		s.SetKey(fmt.Sprintf(":%d", port))
	})
}

func RunUDPServer(port int, options ...io.OptionServer) error {
	return RunServer(NewUDPServer(port, options...))
}

func NewUDPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewProxyServer(UDPListenFunc(port), TCPFunc(addr), options...)
}

func RunUDPProxyServer(port int, addr string, options ...io.OptionServer) error {
	return RunProxyServer(UDPListenFunc(port), TCPFunc(addr), options...)
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

//================================MemoryListen================================

func MemoryListener(key string) (io.Listener, error) {
	s, _ := memoryServerManage.GetOrSetByHandler(key, func() (interface{}, error) {
		return &_memoryServer{
			key: key,
			ch:  make(chan io.ReadWriteCloser, 1000),
		}, nil
	})
	return s.(*_memoryServer), nil
}

func MemoryListenFunc(key string) io.ListenFunc {
	return func() (io.Listener, error) {
		return MemoryListener(key)
	}
}

func NewMemoryServer(key string, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(MemoryListenFunc(key), options...)
}

func RunMemoryServer(key string, options ...io.OptionServer) error {
	return io.RunServer(MemoryListenFunc(key), options...)
}

var memoryServerManage = maps.NewSafe()

type _memoryClient struct {
	*bytes.Buffer
}

func (this *_memoryClient) Close() error {
	this.Reset()
	return nil
}

// _memoryServer 虚拟服务,为了实现接口
type _memoryServer struct {
	key string
	ch  chan io.ReadWriteCloser
}

func (this *_memoryServer) connect() (io.ReadWriteCloser, error) {
	c := &_memoryClient{Buffer: bytes.NewBuffer(nil)}
	select {
	case this.ch <- c:
	case <-time.After(io.DefaultConnectTimeout):
		return nil, io.ErrWithTimeout
	}
	return c, nil
}

func (this *_memoryServer) Accept() (io.ReadWriteCloser, string, error) {
	c := <-this.ch
	return c, fmt.Sprintf("%p", c), nil
}

func (this *_memoryServer) Close() error {
	memoryServerManage.Del(this.key)
	return nil
}

func (this *_memoryServer) Addr() string {
	return fmt.Sprintf("%p", this)
}
