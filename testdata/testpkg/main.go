package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	s, err := listen.NewTCPServer(10089)
	logs.PanicErr(err)
	s.SetPrintWithHEX()
	s.SetReadWriteWithPkg()
	s.SetDealFunc(func(c *io.Client, msg io.Message) {
		//logs.Debug(msg.String())
	})
	go s.Run()

	<-dial.RedialTCP("127.0.0.1:10089", func(c *io.Client) {
		c.SetPrintWithHEX()
		c.SetReadWriteWithPkg()
		c.GoTimerWriteString(time.Second*5, "666")
	}).DoneAll()
}
