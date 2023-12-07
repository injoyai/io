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
	s.SetReadFunc(io.ReadWithSimple)
	s.SetWriteFunc(io.WriteWithSimple)
	s.SetDealFunc(func(c *io.Client, msg io.Message) {
		//logs.Debug(msg.String())
	})
	go s.Run()

	<-dial.RedialTCP("127.0.0.1:10089", func(c *io.Client) {
		c.GoTimerWriter(time.Second*5, func(w *io.IWriter) error {
			p := &io.Simple{
				Control: 0,
				Type:    io.SimpleRead,
				Data: []io.SimpleKeyVal{
					{Key: "测试", Val: 6666},
				},
			}
			_, err := w.Write(p.Bytes())
			return err
		})
	}).DoneAll()

}
