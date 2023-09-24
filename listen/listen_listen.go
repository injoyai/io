package listen

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/internal/common"
	"net"
	"net/http"
	"sync"
)

//================================TCPListen================================

func TCP(port int) (io.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &_tcpServer{Listener: listener}, nil
}

func WithTCP(port int) io.ListenFunc {
	return func() (io.Listener, error) { return TCP(port) }
}

func NewTCPServer(port int, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(WithTCP(port), func(s *io.Server) {
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetOptions(options...)
	})
}

func RunTCPServer(port int, options ...io.OptionServer) error {
	return RunServer(NewTCPServer(port, options...))
}

func NewTCPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewProxyServer(WithTCP(port), dial.WithTCP(addr), options...)
}

func RunTCPProxyServer(port int, addr string, options ...io.OptionServer) error {
	return RunProxyServer(WithTCP(port), dial.WithTCP(addr), options...)
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

func UDP(port int) (io.Listener, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		return nil, err
	}
	return &UDPServer{UDPConn: listener, m: make(map[string]*UDPClient)}, nil
}

func WithUDP(port int) io.ListenFunc {
	return func() (io.Listener, error) {
		return UDP(port)
	}
}

func NewUDPServer(port int, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(WithUDP(port), func(s *io.Server) {
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetOptions(options...)
	})
}

func RunUDPServer(port int, options ...io.OptionServer) error {
	return RunServer(NewUDPServer(port, options...))
}

func NewUDPProxyServer(port int, addr string, options ...io.OptionServer) (*io.Server, error) {
	return NewProxyServer(WithUDP(port), dial.WithTCP(addr), options...)
}

func RunUDPProxyServer(port int, addr string, options ...io.OptionServer) error {
	return RunProxyServer(WithUDP(port), dial.WithTCP(addr), options...)
}

type UDPClient struct {
	s    *UDPServer
	addr *net.UDPAddr
	buff chan []byte
}

func (this *UDPClient) Read(p []byte) (int, error) {
	return 0, nil
}

func (this *UDPClient) ReadMessage() ([]byte, error) {
	return <-this.buff, nil
}

func (this *UDPClient) Write(p []byte) (int, error) {
	return this.s.WriteToUDP(p, this.addr)
}

func (this *UDPClient) Close() error {
	this.s.mu.Lock()
	defer this.s.mu.Unlock()
	delete(this.s.m, this.addr.String())
	return nil
}

// UDPServer todo 待优化
type UDPServer struct {
	*net.UDPConn
	m  map[string]*UDPClient
	mu sync.RWMutex
	m2 *maps.Safe
}

func (this *UDPServer) NewUDPClient(addr string) (*UDPClient, error) {
	raddr, err := net.ResolveUDPAddr(io.UDP, addr)
	if err != nil {
		return nil, err
	}
	return this.newUDPClient(raddr), nil
}

func (this *UDPServer) newUDPClient(addr *net.UDPAddr) *UDPClient {
	v, _ := this.m2.GetOrSetByHandler(addr.String(), func() (interface{}, error) {
		return &UDPClient{
			s:    this,
			addr: addr,
			buff: make(chan []byte, 100),
		}, nil
	})
	return v.(*UDPClient)
}

func (this *UDPServer) Accept() (io.ReadWriteCloser, string, error) {
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

		u := this.newUDPClient(addr)

		this.mu.Lock()
		this.m[addr.String()] = u
		this.mu.Unlock()

		u.buff <- buff[:n]

		return u, u.addr.String(), nil
	}
}

func (this *UDPServer) Addr() string {
	return this.UDPConn.RemoteAddr().String()
}

//================================MemoryListen================================

func Memory(key string) (io.Listener, error) {
	return common.NewMemoryServer(key), nil
}

func WithMemory(key string) io.ListenFunc {
	return func() (io.Listener, error) {
		return Memory(key)
	}
}

func NewMemoryServer(key string, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(WithMemory(key), func(s *io.Server) {
		s.SetKey(key)
		s.SetOptions(options...)
	})
}

func RunMemoryServer(key string, options ...io.OptionServer) error {
	return RunServer(NewMemoryServer(key, options...))
}

//================================WebsocketListen================================

func Websocket(port int) (io.Listener, error) {
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

func WithWebsocket(port int) io.ListenFunc {
	return func() (io.Listener, error) { return Websocket(port) }
}

func NewWebsocketServer(port int, options ...io.OptionServer) (*io.Server, error) {
	return io.NewServer(WithWebsocket(port), func(s *io.Server) {
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
