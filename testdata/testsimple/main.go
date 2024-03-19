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
	s.SetReadFunc(io.ReadWithSimple)
	s.SetWriteFunc(io.WriteWithSimple)
	s.SetDealFunc(func(c *io.Client, msg io.Message) {
		//logs.Debug(msg.String())
	})
	go s.Run()

	<-dial.RedialTCP("127.0.0.1:10089", func(c *io.Client) {
		c.GoTimerWriter(time.Second*5, func(w *io.Client) error {
			p := &io.Simple{
				Control: io.SimpleControl{
					IsResponse: false,
					IsErr:      false,
					Type:       io.OprRead,
				},
				Data: io.SimpleData{
					"测试": []byte{1, 2, 3, 4, 5},
				},
			}
			_, err := w.Write(p.Bytes())
			return err
		})
	}).DoneAll()

}
