package rpc

import (
	"fmt"
	"github.com/injoyai/base/g"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"time"
)

type Client struct {
	cfg  *ClientConfig
	pool *io.Pool
	wait *wait.Entity
	bind *maps.Safe
}

func (this *Client) Bind(Type string, handler Handler) {
	this.bind.Set(Type, handler)
}

func (this *Client) Do(Type string, data interface{}) (interface{}, error) {
	return do(this.pool, this.wait, Type, data)
}

func NewClient(cfg *ClientConfig, option ...io.OptionClient) *Client {
	cfg.init()
	cli := &Client{
		cfg:  cfg,
		wait: wait.New(cfg.ResponseTimeout),
	}
	cli.pool = io.NewPool(dial.WithTCP(cfg.Address), func(c *io.Client) {
		c.SetOptions(option...)
		c.SetReadWriteWithPkg()
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			dealFunc(cli.bind, cli.wait, c, msg)
		})
		c.WriteAny(&io.Model{
			Type: io.Register,
			UID:  g.UUID(),
			Data: cfg,
		})
	})
	go cli.pool.PutNew(3)
	return cli
}

type ClientConfig struct {
	Address         string        `json:"-"`       //请求地址
	Name            string        `json:"name"`    //客户端名称
	Memo            string        `json:"memo"`    //客户端备注
	Version         string        `json:"version"` //客户端版本
	ConnectTimeout  time.Duration `json:"-"`       //连接超时时间
	ResponseTimeout time.Duration `json:"-"`       //响应超时时间
}

func (this *ClientConfig) init() {
	if len(this.Name) == 0 {
		this.Name = fmt.Sprintf("%p", this)
	}
	if len(this.Version) == 0 {
		this.Version = "v0.0.0"
	}
	if this.ConnectTimeout <= 0 {
		this.ConnectTimeout = io.DefaultConnectTimeout
	}
	if this.ResponseTimeout <= 0 {
		this.ResponseTimeout = io.DefaultResponseTimeout
	}
}
