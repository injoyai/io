package io

import (
	"context"
	"errors"
)

// NewPool 新建连接池
func NewPool(dial DialFunc, num int, options ...func(ctx context.Context, c *Client)) *Pool {
	p := &Pool{
		client: make(map[string]*Client),
	}
	p.IWriter = NewWriter(p)
	p.ICloser = NewICloser(p)
	go func() {
		for i := 0; i < num; i++ {
			c := Redial(dial, options...)
			p.client[c.GetKey()] = c
		}
	}()
	return p
}

// Pool 简单版连接池
type Pool struct {
	*IWriter
	*ICloser
	client     map[string]*Client
	chooseLast *Client
	chooseFunc func(all map[string]*Client) (*Client, error)
}

// SetChooseFunc 设置选择客户端规则函数,默认随机
func (this *Pool) SetChooseFunc(fn func(all map[string]*Client) (*Client, error)) *Pool {
	this.chooseFunc = fn
	return this
}

// 获取一个客户端
func (this *Pool) choose() (*Client, error) {
	if this.chooseFunc == nil {
		this.chooseFunc = func(all map[string]*Client) (*Client, error) {
			for _, v := range all {
				if !v.Closed() {
					return v, nil
				}
			}
			return nil, errors.New("连接失败")
		}
	}
	return this.chooseFunc(this.client)
}

// Write 实现io.Writer接口
func (this *Pool) Write(p []byte) (int, error) {
	c, err := this.choose()
	if err != nil {
		return 0, err
	}
	return c.Write(p)
}

// Close 实现io.Closer接口
func (this *Pool) Close() error {
	for _, v := range this.client {
		v.CloseAll()
	}
	this.client = make(map[string]*Client)
	return nil
}
