package dial

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"net"
	"net/http"
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
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetOptions(options...)
	})
}

func RunTCPServer(port int, options ...io.OptionServer) error {
	return RunServer(NewTCPServer(port, options...))
}

func NewTCPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewProxyServer(TCPListenFunc(port), WithTCP(addr), options...)
}

func RunTCPProxyServer(port int, addr string, options ...io.OptionServer) error {
	return RunProxyServer(TCPListenFunc(port), WithTCP(addr), options...)
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
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetOptions(options...)
	})
}

func RunUDPServer(port int, options ...io.OptionServer) error {
	return RunServer(NewUDPServer(port, options...))
}

func NewUDPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewProxyServer(UDPListenFunc(port), WithTCP(addr), options...)
}

func RunUDPProxyServer(port int, addr string, options ...io.OptionServer) error {
	return RunProxyServer(UDPListenFunc(port), WithTCP(addr), options...)
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

//================================WebsocketListen================================

func WebsocketListenFunc(port int) io.ListenFunc {
	return func() (io.Listener, error) { return WebsocketListener(port) }
}

func NewWebsocketServer(port int, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(WebsocketListenFunc(port), func(s *io.Server) {
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetOptions(options...)
	})
}

func RunWebsocketServer(port int, options ...io.OptionServer) error {
	return RunServer(NewWebsocketServer(port, options...))
}

type _websocketClient struct {
	*websocket.Conn
}

func (this *_websocketClient) Read(p []byte) (int, error) {
	return 0, errors.New("请使用ReadMessage")
}

func (this *_websocketClient) Write(p []byte) (int, error) {
	err := this.Conn.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}

func (this *_websocketClient) ReadMessage() ([]byte, error) {
	_, data, err := this.Conn.ReadMessage()
	return data, err
}

func WebsocketListener(port int) (io.Listener, error) {
	ch := make(chan *websocket.Conn)
	s := &_websocketServer{
		s: &http.Server{
			Addr: fmt.Sprintf(":%d", port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ws, err := websocket.Upgrade(w, r, r.Header, 4096, 4096)
				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte(err.Error()))
					return
				}
				ch <- ws
			}),
		},
		c:      ch,
		closed: make(chan struct{}),
	}
	go func() {
		s.err = s.s.ListenAndServe()
		close(s.closed)
	}()
	return s, nil
}

type _websocketServer struct {
	s      *http.Server
	c      chan *websocket.Conn
	err    error
	closed chan struct{}
}

func (this *_websocketServer) Accept() (io.ReadWriteCloser, string, error) {
	select {
	case <-this.closed:
		return nil, "", this.err
	case ws := <-this.c:
		return &_websocketClient{ws}, ws.RemoteAddr().String(), nil
	}
}

func (this *_websocketServer) Close() error {
	return this.s.Close()
}

func (this *_websocketServer) Addr() string {
	return this.s.Addr
}
