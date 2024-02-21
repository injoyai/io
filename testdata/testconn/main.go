package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {

	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	logs.SetShowColor(false)

	listen.RunTCPServer(12000, func(s *io.Server) {
		s.Debug(false)
		go func() {
			for {
				<-time.After(time.Second * 5)
				logs.Debug(s.GetClientLen())
			}
		}()
	})
}
