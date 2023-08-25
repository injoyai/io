package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
)

func main() {
	s, err := dial.NewTCPServer(20088, func(s *io.Server) {
		s.Debug(true)
		s.SetPrintWithASCII()
	})
	if err != nil {
		logs.Err(err)
		return
	}
	dial.NewTunnelServer(s)
	logs.Err(s.Run())
}
