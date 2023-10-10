package p2p

import (
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"testing"
	"time"
)

func TestNewPeer(t *testing.T) {
	remoteAddr1 := "39.107.120.124:20001"
	remoteAddr2 := "39.107.120.124:20002"
	p, err := NewPeer(20000)
	if err != nil {
		t.Log(err)
		return
	}
	go p.Run()
	for {
		<-time.After(time.Second * 5)
		p.WriteTo(remoteAddr1, []byte("666"))
		p.WriteTo(remoteAddr2, []byte("666"))
	}
	select {}
}

func TestNewPeer2(t *testing.T) {
	p, err := NewPeer(20001)
	if err != nil {
		t.Log(err)
		return
	}
	p.SetBeforeFunc(func(c *io.Client) error {
		logs.Debug(c.GetKey())
		return nil
	})
	p.SetDealFunc(func(c *io.Client, msg io.Message) {
		logs.Debug(msg.String())
	})
	t.Log(p.Run())
}

func TestNewPeer3(t *testing.T) {
	remoteAddr1 := "39.107.120.124:20001"
	remoteAddr2 := "39.107.120.124:20002"

	p, err := NewPeer(20000)
	if err != nil {
		t.Log(err)
		return
	}
	go p.Run()
	for {
		<-time.After(time.Second * 5)
		p.WriteTo(remoteAddr1, []byte("666"))
		p.WriteTo(remoteAddr2, []byte("666"))
	}
	select {}
}
