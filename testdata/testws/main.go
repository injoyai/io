package main

import (
	"bufio"
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"os"
)

func main() {
	buf := bufio.NewReader(os.Stdin)
	<-dial.RedialWebsocket("ws://192.168.10.15:10001/api/ssh/ws", nil, func(c *io.Client) {
		c.Debug()
		go c.For(func(ctx context.Context) error {
			msg, _, err := buf.ReadLine()
			if err != nil {
				return err
			}
			if string(msg) == "666" {
				msg = []byte{0x03}
			}
			_, err = c.Write(msg)
			return err
		})
	}).DoneAll()
}
