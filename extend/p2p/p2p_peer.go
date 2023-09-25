package p2p

import (
	"errors"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"net"
	"time"
)

const (
	UDP = "udp"
)

type Peer interface {
	LocalAddr() net.Addr               //本地地址
	Ping(addr string) error            //ping下地址,如果协议一直,则有消息返回
	Base(addr string) (MsgBase, error) //获取基本信息
	Find(addr string) (MsgFind, error) //
	Connect(addr string) error         //建立连接
}

func NewPeer(port int, options ...io.OptionServer) (*peer, error) {
	s, err := listen.NewUDPServer(port, func(s *io.Server) {
		s.SetReadWriteWithPkg()
		s.SetOptions(options...)
	})
	if err != nil {
		return nil, err
	}
	return &peer{
		port:      port,
		localAddr: &net.UDPAddr{Port: port},
		Server:    s,
		clients:   maps.NewSafe(),
	}, nil
}

type peer struct {
	port      int //占用的端口
	localAddr *net.UDPAddr
	*io.Server
	clients *maps.Safe
}

func (this *peer) WriteTo(addr string, p []byte) (int, error) {
	c, err := this.GetClientOrDial(addr, func() (io.ReadWriteCloser, error) {
		return this.Listener().(*listen.UDPServer).NewUDPClient(addr)
	})
	if err != nil {
		return 0, err
	}
	return c.Write(p)
}

func (this *peer) Ping(addr string) error {
	c, err := this.GetClientOrDial(addr, func() (io.ReadWriteCloser, error) {
		return this.Listener().(*listen.UDPServer).NewUDPClient(addr)
	})
	if err != nil {
		return err
	}
	if _, err := c.WriteString(io.Ping); err != nil {
		return err
	}
	resp, err := c.ReadLast(time.Second)
	if err != nil {
		return err
	}
	if string(resp) != io.Pong {
		return errors.New("响应失败")
	}
	return nil
}

func (this *peer) Base(addr string) (MsgBase, error) {
	return MsgBase{}, nil
}

func (this *peer) Find(addr string) (MsgFind, error) {
	//TODO implement me
	panic("implement me")
}

func (this *peer) Connect(addr string) error {
	//TODO implement me
	panic("implement me")
}

func (this *peer) LocalAddr() net.Addr {
	return this.localAddr
}
