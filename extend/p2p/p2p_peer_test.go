package p2p

import (
	"net"
	"testing"
)

func TestNewPeer(t *testing.T) {
	//remoteAddr := "127.0.0.1:20000"
	//p, err := NewPeer(20000)
	//if err != nil {
	//	t.Log(err)
	//	return
	//}

	//c, err := p.dial(remoteAddr)
	//if err != nil {
	//	t.Log(err)
	//	return
	//}
	//_ = p
	////p.c.WriteTo([]byte("6666"),remoteAddr)
	//t.Log(p.peer.LocalAddr())
	//t.Log(p.peer.RemoteAddr())
	//_, err = p.peer.WriteToUDP([]byte("6666"))
	//t.Log(err)
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
