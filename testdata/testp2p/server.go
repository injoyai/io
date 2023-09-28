package main

import (
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/g"
	"github.com/injoyai/io"
	"github.com/injoyai/io/extend/p2p"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	remoteAddr := "218.108.149.186:20000"
	p, err := p2p.NewPeer(20001)
	if err != nil {
		logs.Err(err)
		return
	}
	p.SetDealFunc(func(msg *io.IMessage) {
		msg.WriteString("success")
	})
	go func() {
		for {
			<-time.After(time.Second * 10)
			_, err := p.WriteTo(remoteAddr, []byte("666"))
			logs.PrintErr(err)
		}
	}()
	go func() {
		g.Input()
		c := chans.NewWaitLimit(1000)
		for i := 1; i <= 65535; i++ {
			c.Add()
			go func(i int) {
				defer c.Done()
				err := p.Ping(fmt.Sprintf("218.108.149.186:%d", i), time.Millisecond*100)
				if !logs.PrintErr(err) {
					logs.Debug("成功")
					panic(fmt.Sprintf("218.108.149.186:%d", i))
				}
			}(i)
		}
	}()
	logs.Err(p.Run())
}
