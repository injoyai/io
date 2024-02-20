package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
)

func main() {
	go listen.RunTCPServer(13000)

	go dial.NewTCP("127.0.0.1:13000", func(c *io.Client) {
		c.Write([]byte{0, 1, 2})
		//<-c.Done()
	})
	select {}
}
