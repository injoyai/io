package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
)

func main() {
	s, err := dial.NewTCPServer(20086, func(s *io.Server) {
		s.Debug(false)
		s.SetPrintWithASCII()
	})
	if err != nil {
		logs.Error(err)
		return
	}
	dial.NewTunnelClient(s, dial.WithTCP(":20088"), func(c *io.Client) {
		c.Debug(true)
		c.SetPrintWithASCII()
	})
	logs.Err(s.Run())
}
