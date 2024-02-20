package main

import (
	"github.com/injoyai/io/dial"
	"time"
)

func main() {
	c, _ := dial.NewTCP(":10086")
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
