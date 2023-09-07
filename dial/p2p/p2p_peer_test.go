package p2p

import (
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
