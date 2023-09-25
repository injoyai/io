package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/extend/p2p"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	remoteAddr := "127.0.0.1:20001"
	p, err := p2p.NewPeer(20000, func(s *io.Server) {
		s.SetPrintWithASCII()
	})
	if err != nil {
		logs.Err(err)
		return
	}
	p.SetDealFunc(func(msg *io.IMessage) {
		logs.Debug(msg.String())
	})
	go p.Run()
	//logs.Debug("开始")
	for {
		logs.Debug(p.Ping(remoteAddr))
		logs.Debug("结束")
		logs.Debug(p.WriteTo(remoteAddr, []byte("666")))
		<-time.After(time.Second * 5)
	}
	select {}
}
