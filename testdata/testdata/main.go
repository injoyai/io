package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	bs := make([]byte, buf.DefaultBufferSize+1)
	go listen.RunTCPServer(10099, func(s *io.Server) {
		s.Debug(false)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			s.SetReadWithAll()
			if msg.Len() == buf.DefaultBufferSize+1 {
				logs.Debug(msg.GetLast())
			}
		})
	})
	<-dial.RedialTCP("127.0.0.1:10099", func(c *io.Client) {
		c.Debug(false)
		n := byte(0)
		c.GoTimerWriter(time.Second, func(w *io.IWriter) error {
			n++
			bs[buf.DefaultBufferSize] = n
			_, err := w.Write(bs)
			return err
		})
	}).DoneAll()
}
