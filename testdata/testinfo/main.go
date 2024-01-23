package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
)

func main() {
	go listen.RunTCPServer(20123)
	dial.NewTCP("127.0.0.1:20123", func(c *io.Client) {
		c.Close()
	})
}
