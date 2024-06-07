package main

import (
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io"
	"github.com/injoyai/io/extend/p2p"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	remoteAddr := "218.108.149.186:20001"
	p, err := p2p.NewPeer(20000)
	if err != nil {
		logs.Err(err)
		return
	}
	p.SetDealFunc(func(c *io.Client, msg io.Message) {
		c.WriteString("success")
	})

	p.Register(remoteAddr)

	go func() {
		for {
			<-time.After(time.Second * 10)
			_, err := p.WriteTo(remoteAddr, []byte("666"))
			logs.PrintErr(err)
		}
	}()
	go func() {
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
