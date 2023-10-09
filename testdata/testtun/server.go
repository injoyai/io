package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
)

func main() {
	s, err := listen.NewTCPServer(10088, func(s *io.Server) {
		s.Debug(true)
		s.Logger.SetPrintWithASCII()
	})
	if err != nil {
		logs.Err(err)
		return
	}
	listen.NewTunnelServer(s)
	logs.Err(s.Run())
}
