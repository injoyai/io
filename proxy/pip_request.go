package proxy

import (
	"context"
	"github.com/injoyai/io"
)

type Request struct {
	SN   string //请求的通道
	key  string //请求标识
	Addr string //请求地址
	*io.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func (this *Request) Key() string {
	return this.key
}

func (this *Request) SetSN(sn string) *Request {
	this.SN = sn
	return this
}

func (this *Request) SetAddr(addr string) *Request {
	this.Addr = addr
	return this
}

func (this *Request) Close() error {
	defer recover()
	if this.cancel != nil {
		this.cancel()
	}
	if this.Client != nil {
		return this.Client.Close()
	}
	return nil
}

func newRequest(ctx context.Context, sn, key, addr string, c *io.Client) *Request {
	r := &Request{
		SN:     sn,
		key:    key,
		Addr:   addr,
		Client: c,
	}
	r.ctx, r.cancel = context.WithCancel(ctx)
	return r
}
