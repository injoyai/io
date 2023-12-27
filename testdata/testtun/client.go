package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
)

func main() {
	s, err := listen.NewTCPServer(10086, func(s *io.Server) {
		s.Debug(true)
		s.Logger.SetPrintWithUTF8()
	})
	if err != nil {
		logs.Error(err)
		return
	}
	listen.NewTunnelClient(s, dial.WithTCP(":10088"), "aiot.qianlangtech.com:8200", func(c *io.Client) {
		c.Debug(false)
		c.SetPrintWithUTF8()
	})
	logs.Err(s.Run())
}
