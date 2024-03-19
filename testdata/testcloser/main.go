package main

import (
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	c, err := dial.NewTCP(":10086")
	logs.PanicErr(err)
	c.Redial()
	c.Close()
	go func() {
		<-time.After(time.Second * 5)
		c.CloseAll()
	}()
	<-c.Done()
	<-c.DoneAll()
	<-time.After(time.Second * 5)
}
