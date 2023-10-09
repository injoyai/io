package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
)

func main() {

	s, err := listen.NewTCPProxyServer(22, "192.168.10.26:22", func(s *io.Server) {
		s.Debug(true)
		s.Logger.SetLevel(io.LevelInfo)
	})
	if err != nil {
		logs.Err(err)
		return
	}
	logs.PrintErr(s.Run())
}
