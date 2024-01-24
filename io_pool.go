package io

import (
	"github.com/injoyai/conv"
	"sync"
	"sync/atomic"
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
	getNum  uint64
	putNum  uint64
}

func (this *Pool) Len() int {
	return len(this.pool)
}

func (this *Pool) GetNum() uint64 {
	return this.getNum
}

func (this *Pool) PutNum() uint64 {
	return this.putNum
}

func (this *Pool) new() (*Client, error) {
	return NewDial(this.dial, this.options...)
}

// Get 从连接池获取一个客户端
func (this *Pool) Get() (c *Client, _ error) {
	defer func() {
		atomic.AddUint64(&this.getNum, 1)
		if c != nil && !c.Running() {
			go c.Run()
		}
	}()
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
		atomic.AddUint64(&this.putNum, 1)
		this.mu.Lock()
		defer this.mu.Unlock()
		this.pool[c.Pointer()] = c
	}
}

func (this *Pool) PutNew(num int) error {
	for i := 0; i < num; i++ {
		c, err := this.new()
		if err != nil {
			return err
		}
		this.Put(c)
	}
	return nil
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

// WriteAny 实现io.AnyWriter接口
func (this *Pool) WriteAny(any interface{}) (int, error) {
	return this.Write(conv.Bytes(any))
}

// Close 实现io.Closer接口
func (this *Pool) Close() error {
	for _, v := range this.pool {
		v.CloseAll()
	}
	this.pool = make(map[string]*Client)
	return nil
}

func (this *Pool) Closed() bool {
	return false
}
