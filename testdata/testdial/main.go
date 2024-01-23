package main

import (
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"net"
)

func main() {
	dial.New(func(ctx context.Context) (io.ReadWriteCloser, string, error) {
		c, err := net.Dial("tcp", "127.0.0.1:10086")
		return c, c.LocalAddr().String(), err
	})
}
