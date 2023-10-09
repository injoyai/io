package p2p

import (
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"testing"
)

func TestNewPeer(t *testing.T) {
	remoteAddr := "127.0.0.1:20001"
	p, err := NewPeer(20000)
	if err != nil {
		t.Log(err)
		return
	}
	go p.Run()
	//logs.Debug("开始")
	t.Log(p.Ping(remoteAddr))
	logs.Debug("结束")
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
