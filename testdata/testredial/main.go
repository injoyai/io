package main

import (
	"context"
	"github.com/injoyai/io"
	"net"
)

func main() {
	c := io.Redial(func(ctx context.Context) (io.ReadWriteCloser, string, error) {
		addr := ":10086"
		c, err := net.Dial(io.TCP, addr)
		return c, addr, err
	}, func(c *io.Client) {
		c.SetKey("999")
	})
	<-c.DoneAll()
}
