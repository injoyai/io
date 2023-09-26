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
	p.SetDealFunc(func(msg *io.IMessage) {
		msg.Write(msg.Bytes())
	})
	logs.Err(p.Run())
}
