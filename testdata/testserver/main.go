package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {

	listen.RunTCPServer(10086, func(s *io.Server) {
		go func() {
			for {
				<-time.After(time.Second * 5)
				logs.Debug(s.GetClientLen())
			}
		}()
	})

}
