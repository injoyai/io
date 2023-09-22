package p2p

import (
	"bufio"
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

func NewPeer(port int) (*peer, error) {
	localAddr := &net.UDPAddr{Port: port}
	c, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}
	listen.NewUDPServer(port)
	if err != nil {
		return nil, err
	}
	return &peer{
		port:      port,
		localAddr: localAddr,
		peer:      c,
		clients:   maps.NewSafe(),
	}, nil
}

type peer struct {
	port      int //占用的端口
	localAddr *net.UDPAddr
	peer      *net.UDPConn //监听的服务
	clients   *maps.Safe
}

func (this *peer) dial(addr string) (*net.UDPConn, error) {
	v, err := this.clients.GetOrSetByHandler(addr, func() (interface{}, error) {
		raddr, err := net.ResolveUDPAddr(io.UDP, addr)
		if err != nil {
			return nil, err
		}
		return net.DialUDP(io.UDP, this.localAddr, raddr)
	})
	if err != nil {
		return nil, err
	}
	return v.(*net.UDPConn), err
}

func (this *peer) WriteTo(addr string, p []byte) (int, error) {
	raddr, err := net.ResolveUDPAddr(io.UDP, addr)
	if err != nil {
		return 0, err
	}
	return this.peer.WriteToUDP(p, raddr)
}

func (this *peer) Ping(addr string) error {
	if _, err := this.WriteTo(addr, io.NewPkgPing()); err != nil {

	}

	raddr, err := net.ResolveUDPAddr(io.UDP, addr)
	if err != nil {
		return err
	}

	c, err := net.DialUDP(io.UDP, this.localAddr, raddr)
	if err != nil {
		return err
	}
	if _, err := c.Write(io.NewPkgPing()); err != nil {
		return err
	}
	if err := c.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}
	p, err := io.ReadPkg(bufio.NewReader(c))
	if err != nil {
		return err
	}
	if !p.IsPong() {
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
