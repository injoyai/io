package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/extend/p2p"
	"github.com/injoyai/logs"
)

func main() {
	p, err := p2p.NewPeer(20001)
	if err != nil {
		logs.Err(err)
		return
	}
	p.SetBeforeFunc(func(c *io.Client) error {
		logs.Debug(c.GetKey())
		return nil
	})
	p.SetDealFunc(func(msg *io.IMessage) {
		logs.Debug(msg.String())
	})
	logs.Err(p.Run())
}
