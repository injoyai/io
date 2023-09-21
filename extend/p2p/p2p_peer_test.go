package p2p

import (
	"net"
	"testing"
)

func TestNewPeer(t *testing.T) {
	remoteAddr := "124.160.230.36:20000"
	p, err := NewPeer(20000, remoteAddr)
	if err != nil {
		t.Log(err)
		return
	}
	_ = p
	//p.c.WriteTo([]byte("6666"),remoteAddr)
	t.Log(p.c.LocalAddr())
	t.Log(p.c.RemoteAddr())
	_, err = p.c.Write([]byte("6666"))
	t.Log(err)
	select {}
}

func TestNewPeer2(t *testing.T) {
	remoteAddr := "127.0.0.1:20001"
	raddr, err := net.ResolveUDPAddr(UDP, remoteAddr)
	if err != nil {
		t.Log(err)
		return
	}
	c, err := net.DialUDP(UDP, &net.UDPAddr{Port: 20001}, raddr)
	if err != nil {
		t.Log(err)
		return
	}
	c.Write([]byte("66666"))
}
