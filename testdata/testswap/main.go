package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
)

func main() {

	s, err := dial.NewTCPSwapServer(22, "192.168.10.26:22", func(s *io.Server) {
		s.Debug(true)
		s.SetPrintWithBase()
	})
	if err != nil {
		logs.Err(err)
		return
	}
	logs.PrintErr(s.Run())
}
