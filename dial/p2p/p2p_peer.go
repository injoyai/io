package p2p

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"net"
)

const (
	UDP = "udp"
)

type Peer interface {
}

func NewPeer(localPort int, remoteAddr string) (*peer, error) {
	localAddr := &net.UDPAddr{Port: localPort}
	//s, err := net.ListenUDP("udp", localAddr)
	s, err := listen.NewUDPServer(localPort, func(s *io.Server) {
		s.Debug()
		s.SetPrintWithASCII()
	})
	if err != nil {
		return nil, err
	}
	raddr, err := net.ResolveUDPAddr(UDP, remoteAddr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP(UDP, localAddr, raddr)
	if err != nil {
		return nil, err
	}
	go s.Run()
	return &peer{p: localPort, s: s, c: c}, nil
}

type peer struct {
	p int        //占用的端口
	s *io.Server //监听的服务
	//s    *net.UDPConn //监听的服务
	c    *net.UDPConn //发起的请求
	stun net.Conn     //代理的服务器
}

func (this *peer) GetIPPort() (net.IP, int, error) {

	return nil, 0, nil
}
