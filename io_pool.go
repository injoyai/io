package io

import (
	"sync"
)

// NewPool 新建连接池
func NewPool(dial DialFunc, options ...OptionClient) *Pool {
	return &Pool{
		dial:    dial,
		options: options,
		pool:    make(map[string]*Client),
	}
}

// Pool 简单版连接池
type Pool struct {
	dial    DialFunc
	options []OptionClient
	pool    map[string]*Client
	mu      sync.RWMutex
}

func (this *Pool) new() (*Client, error) {
	return NewDial(this.dial, this.options...)
}

// Get 从连接池获取一个客户端
func (this *Pool) Get() (*Client, error) {
	this.mu.RLock()
	for _, v := range this.pool {
		this.mu.RUnlock()
		return v, nil
	}
	this.mu.RUnlock()
	return this.new()
}

// Put 放回连接池
func (this *Pool) Put(c *Client) {
	if c != nil && !c.Closed() {
		this.mu.Lock()
		defer this.mu.Unlock()
		this.pool[c.GetKey()] = c
	}
}

// Write 实现io.Writer接口
func (this *Pool) Write(p []byte) (int, error) {
	c, err := this.Get()
	if err != nil {
		return 0, err
	}
	defer this.Put(c)
	return c.Write(p)
}

// Close 实现io.Closer接口
func (this *Pool) Close() error {
	for _, v := range this.pool {
		v.CloseAll()
	}
	this.pool = make(map[string]*Client)
	return nil
}
