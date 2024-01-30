package main

import (
	"github.com/injoyai/io/bridge"
	"github.com/injoyai/logs"
)

func main() {
	s, err := bridge.NewServer(20088)
	if err != nil {
		logs.Err(err)
		return
	}

	_, err = s.Listen("tcp", "10086")
	logs.PrintErr(err)
	s.Run()
}
