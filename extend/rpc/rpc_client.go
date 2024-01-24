package rpc

import (
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"time"
)

type Client struct {
	*io.Pool
	wait *wait.Entity
	bind *maps.Safe
}

func (this *Client) Bind(Type string, handler Handler) {
	this.bind.Set(Type, handler)
}

func (this *Client) Do(Type string, data interface{}) (interface{}, error) {
	return do(this.Pool, this.wait, Type, data)
}

func NewClient(addr string, waitTimeout time.Duration, option ...io.OptionClient) *Client {
	cli := &Client{
		wait: wait.New(waitTimeout),
	}
	cli.Pool = io.NewPool(dial.WithTCP(addr), func(c *io.Client) {
		c.SetOptions(option...)
		c.SetReadWriteWithPkg()
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			dealFunc(cli.bind, cli.wait, c, msg)
		})
	})
	go cli.Pool.PutNew(3)
	return cli
}
