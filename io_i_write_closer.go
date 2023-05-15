package io

import (
	"context"
	"sync/atomic"
)

func NewIWriteCloser(writeCloser WriteCloser) *IWriteCloser {
	return NewIWriteCloserWithContext(context.Background(), writeCloser)
}

func NewIWriteCloserWithContext(ctx context.Context, writeCloser WriteCloser) *IWriteCloser {
	return &IWriteCloser{
		IWriter: NewIWriter(writeCloser),
		ICloser: NewICloserWithContext(ctx, writeCloser),
	}
}

type IWriteCloser struct {
	*IWriter
	*ICloser
	queue   chan []byte //写入队列
	running uint32      //是否在运行
}

func (this *IWriteCloser) GetKey() string {
	return this.IWriter.GetKey()
}

// SetKey 设置唯一标识
func (this *IWriteCloser) SetKey(key string) *IWriteCloser {
	this.IWriter.SetKey(key)
	this.ICloser.SetKey(key)
	return this
}

// SetPrintFunc 设置打印函数
func (this *IWriteCloser) SetPrintFunc(fn PrintFunc) *IWriteCloser {
	this.IWriter.SetPrintFunc(fn)
	this.ICloser.SetPrintFunc(fn) //错误信息按ASCII编码?
	return this
}

func (this *IWriteCloser) Debug(b ...bool) *IWriteCloser {
	this.IWriter.Debug(b...)
	this.ICloser.Debug(b...)
	return this
}

// WriteQueue 写入队列
func (this *IWriteCloser) WriteQueue(p []byte) *IWriteCloser {
	this.runQueue()
	this.queue <- p
	return this
}

// TryWriteQueue 尝试写入队列
func (this *IWriteCloser) TryWriteQueue(p []byte) *IWriteCloser {
	this.runQueue()
	select {
	case this.queue <- p:
	default:
	}
	return this
}

func (this *IWriteCloser) runQueue() {
	if this.queue == nil {
		this.queue = this.NewWriteQueue(this.Ctx())
	}
	if atomic.SwapUint32(&this.running, 1) == 0 {
		go this.For(func() error {
			_, err := this.Write(<-this.queue)
			return err
		})
	}
}
