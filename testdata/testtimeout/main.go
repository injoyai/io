package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	go listen.RunTCPServer(13000, func(s *io.Server) {
		s.SetLevelInfo()
		s.Keep.SetTimeout(time.Second * 10)
		s.Keep.SetInterval(time.Second)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			c.SetKey(msg.String())
			//s.ClientManage.SetClientKey(c, msg.String())
		})
		go func() {
			for {
				<-time.After(time.Second * 3)

				c := s.GetClient("2")
				if c == nil {
					logs.Debug("c==nil")
				} else {
					c.WriteString("2")
				}
			}
		}()
	})

	go dial.NewTCP("127.0.0.1:13000", func(c *io.Client) {
		//无发送无接收数据
		c.WriteString("1")
	})
	go dial.NewTCP("127.0.0.1:13000", func(c *io.Client) {
		c.Debug(false)
		//接收数据
		c.WriteString("2")
		//go c.Run()
	})
	go dial.NewTCP("127.0.0.1:13000", func(c *io.Client) {
		c.Debug(false)
		//持续发送数据
		c.GoTimerWriteString(time.Second, "3")
	})
	select {}
}
