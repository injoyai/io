package dial

import (
	"errors"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/internal/common"
	"net"
	"os"
)

//================================TCPDial================================

// TCP 连接
func TCP(addr string) (io.ReadWriteCloser, string, error) {
	r, err := net.Dial(io.TCP, addr)
	return r, addr, err
}

// WithTCP 连接函数
func WithTCP(addr string) io.DialFunc {
	return func() (io.ReadWriteCloser, string, error) { return TCP(addr) }
}

// NewTCP 新建TCP连接
func NewTCP(addr string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithTCP(addr), options...)
}

// RedialTCP 一直连接TCP服务端,并重连
func RedialTCP(addr string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithTCP(addr), options...)
}

//================================UDPDial================================

// UDP 连接
func UDP(addr string) (io.ReadWriteCloser, string, error) {
	c, err := net.Dial(io.UDP, addr)
	return c, addr, err
}

// WithUDP 连接函数
func WithUDP(addr string) io.DialFunc {
	return func() (io.ReadWriteCloser, string, error) { return UDP(addr) }
}

func NewUDP(addr string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithUDP(addr), options...)
}

// RedialUDP 一直连接UDP服务端,并重连
func RedialUDP(addr string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithUDP(addr), options...)
}

var udpMap *maps.Safe

func WriteUDP(addr string, p []byte, selfPort ...int) error {
	if udpMap == nil {
		udpMap = maps.NewSafe()
	}
	v := udpMap.GetInterface(addr)
	if v == nil {
		c, err := net.Dial(io.UDP, addr)
		if err != nil {
			return err
		}
		udpMap.Set(addr, c)
		v = c
	}
	c := v.(net.Conn)
	_, err := c.Write(p)
	return err
}

//================================FileDial================================

// File 打开文件
func File(path string) (io.ReadWriteCloser, string, error) {
	c, err := os.Open(path)
	return c, path, err
}

// WithFile 打开文件函数
func WithFile(path string) io.DialFunc {
	return func() (io.ReadWriteCloser, string, error) { return File(path) }
}

func NewFile(path string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithFile(path), options...)
}

//================================MemoryDial================================

// Memory 内存
func Memory(key string) (io.ReadWriteCloser, string, error) {
	s := common.MemoryServerManage.MustGet(key)
	if s == nil {
		return nil, "", errors.New("服务不存在")
	}
	c, err := s.(*common.MemoryServer).Connect()
	return c, key, err
}

func WithMemory(key string) io.DialFunc {
	return func() (io.ReadWriteCloser, string, error) { return Memory(key) }
}

func NewMemory(key string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithMemory(key), options...)
}

func RedialMemory(key string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithMemory(key), options...)
}

//================================Rabbitmq================================

//================================Other================================
