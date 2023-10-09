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
func TCP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial(io.TCP, addr)
}

// WithTCP 连接函数
func WithTCP(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return TCP(addr) }
}

// NewTCP 新建TCP连接
func NewTCP(addr string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithTCP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

// RedialTCP 一直连接TCP服务端,并重连
func RedialTCP(addr string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithTCP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

//================================UDPDial================================

// UDP 连接
func UDP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("udp", addr)
}

// WithUDP 连接函数
func WithUDP(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return UDP(addr) }
}

func NewUDP(addr string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithUDP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

// RedialUDP 一直连接UDP服务端,并重连
func RedialUDP(addr string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithUDP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

var udpMap *maps.Safe

func WriteUDP(addr string, p []byte) error {
	if udpMap == nil {
		udpMap = maps.NewSafe()
	}
	v := udpMap.GetInterface(addr)
	if v == nil {
		c, err := net.Dial("udp", addr)
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
func File(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}

// WithFile 打开文件函数
func WithFile(path string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return os.Open(path)
	}
}

func NewFile(path string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithFile(path), func(c *io.Client) {
		c.SetKey(path)
		c.SetOptions(options...)
	})
}

//================================MemoryDial================================

// Memory 内存
func Memory(key string) (io.ReadWriteCloser, error) {
	s := common.MemoryServerManage.MustGet(key)
	if s == nil {
		return nil, errors.New("服务不存在")
	}
	return s.(*common.MemoryServer).Connect()
}

func WithMemory(key string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return Memory(key)
	}
}

func NewMemory(key string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithMemory(key), func(c *io.Client) {
		c.SetKey(key)
		c.SetOptions(options...)
	})
}

//================================RabbitmqDial================================

//================================OtherDial================================
