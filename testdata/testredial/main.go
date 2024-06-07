package main

import (
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"net"
)

func main() {

	go listen.RunTCPServer(10086, func(s *io.Server) {
		s.Debug(false)
		s.SetConnectFunc(func(c *io.Client) error {
			return c.Close()
		})
	})

	c := io.Redial(func(ctx context.Context) (io.ReadWriteCloser, string, error) {
		addr := ":10086"
		c, err := net.Dial(io.TCP, addr)
		return c, addr, err
	}, func(c *io.Client) {
		c.SetKey("999")
	})
	<-c.DoneAll()
}
