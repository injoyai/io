package io

import (
	"context"
	"sync"
	"time"
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
	queue chan []byte //写入队列
	once  sync.Once   //队列只执行一次
}

// SetKey 设置唯一标识
func (this *IWriteCloser) SetKey(key string) *IWriteCloser {
	this.IWriter.SetKey(key)
	this.ICloser.SetKey(key)
	return this
}

// Debug 调试模式
// 实现Debugger接口
func (this *IWriteCloser) Debug(b ...bool) {
	this.IWriter.Logger.Debug(b...)
	this.ICloser.Logger.Debug(b...)

}

// WriteQueue 写入队列
func (this *IWriteCloser) WriteQueue(p []byte) *IWriteCloser {
	this.runQueue()
	this.queue <- p
	return this
}

// WriteQueueTimeout 写入队列,超时
func (this *IWriteCloser) WriteQueueTimeout(p []byte, timeout time.Duration) (int, error) {
	this.runQueue()
	select {
	case this.queue <- p:
		return len(p), nil
	case <-time.After(timeout):
		return 0, ErrWithWriteTimeout
	}
}

// WriteQueueTry 尝试写入队列
func (this *IWriteCloser) WriteQueueTry(p []byte) *IWriteCloser {
	this.runQueue()
	select {
	case this.queue <- p:
	default:
	}
	return this
}

func (this *IWriteCloser) runQueue() {
	this.once.Do(func() {
		this.queue = this.NewWriteQueue(this.Ctx())
		go this.For(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case p := <-this.queue:
				_, err := this.Write(p)
				return err
			}
		})
	})
}
