package p2p

import (
	"net"
)

const (
	UDP = "udp"
)

type Peer interface {
}

func NewPeer(localPort int, remoteAddr string) (Peer, error) {
	localAddr := &net.UDPAddr{Port: localPort}
	s, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}
	raddr, err := net.ResolveUDPAddr(UDP, remoteAddr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP(UDP, localAddr, raddr)
	return &peer{p: localPort, s: s, c: c}, err
}

type peer struct {
	p    int          //占用的端口
	s    *net.UDPConn //监听的服务
	c    *net.UDPConn //发起的请求
	stun net.Conn     //代理的服务器
}

func (this *peer) GetIPPort() (net.IP, int, error) {

	return nil, 0, nil
}
