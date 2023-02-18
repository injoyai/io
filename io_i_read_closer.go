package io

import (
	"context"
	"fmt"
	"sync/atomic"
)

func NewIReadCloser(readCloser ReadCloser) *IReadCloser {
	return NewIReadCloserWithContext(context.Background(), readCloser)
}

func NewIReadCloserWithContext(ctx context.Context, readCloser ReadCloser) *IReadCloser {
	if c, ok := readCloser.(*IReadCloser); ok && c != nil {
		return c
	}
	return &IReadCloser{
		IReader: NewIReader(readCloser),
		ICloser: NewICloserWithContext(ctx, readCloser),
		running: 0,
	}
}

type IReadCloser struct {
	*IReader
	*ICloser
	dealFunc DealFunc //处理数据函数
	running  uint32   //是否在运行
}

//================================Nature================================

func (this *IReadCloser) SetKey(key string) *IReadCloser {
	this.IPrinter.SetKey(key)
	this.ICloser.SetKey(key)
	return this
}

func (this *IReadCloser) SetPrintFunc(fn PrintFunc) *IReadCloser {
	this.IPrinter.SetPrintFunc(fn)
	//错误信息按ASCII编码
	return this
}

func (this *IReadCloser) Debug(b ...bool) *IReadCloser {
	this.IPrinter.Debug(b...)
	this.ICloser.Debug(b...)
	return this
}

//================================DealFunc================================

// SetDealFunc 设置数据处理函数
func (this *IReadCloser) SetDealFunc(fn func(msg Message)) {
	this.dealFunc = fn
}

// SetDealWithNil 不设置数据处理函数
func (this *IReadCloser) SetDealWithNil() {
	this.SetDealFunc(nil)
}

// SetDealWithWriter 设置数据处理到io.Writer
func (this *IReadCloser) SetDealWithWriter(writer Writer) {
	this.SetDealFunc(func(msg Message) {
		writer.Write(msg)
	})
}

// SetDealWithChan 设置数据处理到chan
func (this *IReadCloser) SetDealWithChan(c chan Message) {
	this.SetDealFunc(func(msg Message) {
		c <- msg
	})
}

//================================RunTime================================

// Running 是否在运行
func (this *IReadCloser) Running() bool {
	return this.running == 1
}

func (this *IReadCloser) Run() error {
	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}
	for {
		select {
		case <-this.Done():
			return this.Err()
		default:
			_ = this.CloseWithErr(func() (err error) {
				defer func() {
					if e := recover(); e != nil {
						err = fmt.Errorf("%v", e)
					}
				}()
				bytes, err := this.ReadMessage()
				if err != nil || len(bytes) == 0 {
					return err
				}
				//打印日志
				this.IPrinter.Print(bytes, TagRead, this.GetKey())
				//处理数据
				if this.dealFunc != nil {
					this.dealFunc(bytes)
				}
				return
			}())
		}
	}
}
