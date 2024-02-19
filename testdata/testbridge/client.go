package main

import (
	"github.com/injoyai/io/extend/bridge"
	"github.com/injoyai/logs"
)

func main() {
	<-bridge.RedialClient("127.0.0.1:20088", func(c *bridge.Client) {
		c.SetPrintWithHEX()
		go func() {
			logs.PrintErr(c.Subscribe("tcp", "10086"))
		}()
	}).DoneAll()
}
