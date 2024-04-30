package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	bs := make([]byte, io.DefaultBufferSize)
	go listen.RunTCPServer(10099, func(s *io.Server) {
		s.Debug(false)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			s.SetReadWith1KB()
			if msg.Len() == io.DefaultBufferSize {
				logs.Debug(msg.GetLast())
			}
		})
	})
	<-dial.RedialTCP("127.0.0.1:10099", func(c *io.Client) {
		c.Debug(false)
		n := byte(0)
		c.GoTimerWriter(time.Second, func(w *io.Client) (int, error) {
			n++
			bs[io.DefaultBufferSize-1] = n
			return w.Write(bs)
		})
	}).DoneAll()
}
