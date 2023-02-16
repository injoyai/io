package testdata

import (
	"context"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"time"
)

func NewClient(addr string) error {
	c := io.Redial(dial.TCPFunc(addr), func(ctx context.Context, c *io.Client) {
		logs.Debug("重连...")
		c.Debug()
	})
	time.Sleep(time.Second * 5)
	logs.Debug("主动关闭")
	c.Close()
	select {}
	return nil
}

func NewServer(port int) error {
	s, err := io.NewServer(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.Debug()
	return s.Run()
}

func NewTestMustDialBug(port int) error {

	s, err := io.NewServer(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.Debug()
	go func() {
		logs.Err(s.Run())
	}()

	c := io.Redial(dial.TCPFunc(fmt.Sprintf(":%d", port)), func(ctx context.Context, c *io.Client) {
		c.Debug()
	})

	<-time.After(time.Second * 1)
	logs.Debug("关闭服务端")
	s.Close()

	<-time.After(time.Second * 5)
	logs.Debug("关闭客户端")
	c.Close()

	<-time.After(time.Second * 10)
	logs.Debug("关闭客户端2")
	c.CloseAll()

	select {}
	return nil
}