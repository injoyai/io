package main

import (
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/extend/rpc"
	"github.com/injoyai/logs"
	"time"
)

func main() {
	s, err := rpc.NewServer(8080, time.Second, func(s *io.Server) {
		s.Debug()
	})
	if err != nil {
		logs.Err(err)
		return
	}
	s.Bind("/test", func(ctx context.Context, c *io.Client, msg *io.Model) (interface{}, error) {
		<-time.After(time.Millisecond * 100)
		return msg.Data, nil
	})
	go s.Run()
	c := rpc.NewClient(&rpc.ClientConfig{
		Address: "127.0.0.1:8080",
	}, func(c *io.Client) {
		c.Debug(false)
		//c.SetPrintWithErr()
	})
	<-time.After(time.Second)
	for i := 0; i < 10; i++ {
		go func(i int) {
			logs.Debug(c.Do("/test", i))
		}(i)
	}
	<-time.After(time.Second * 10)

}
