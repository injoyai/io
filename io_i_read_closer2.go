package io

import (
	"context"
	"fmt"
	"sync/atomic"
)

func NewIReaderCloser(readCloser MessageReadCloser) *IReaderCloser {
	return NewMReadCloserWithContext(context.Background(), readCloser)
}

func NewMReadCloserWithContext(ctx context.Context, readCloser MessageReadCloser) *IReaderCloser {
	if c, ok := readCloser.(*IReaderCloser); ok && c != nil {
		return c
	}
	entity := &IReaderCloser{
		IPrinter: NewIPrinter(""),
		ICloser:  NewICloserWithContext(ctx, readCloser),
		running:  0,
	}
	entity.SetCloseFunc(func(ctx context.Context, msg Message) {
		entity.cancel()
	})
	return entity
}

type IReaderCloser struct {
	*IPrinter
	*ICloser
	*IReader
	dealFunc DealFunc //处理数据函数
	running  uint32   //是否在运行
}

//================================Nature================================

func (this *IReaderCloser) SetKey(key string) *IReaderCloser {
	this.IPrinter.SetKey(key)
	this.ICloser.SetKey(key)
	return this
}

func (this *IReaderCloser) SetPrintFunc(fn PrintFunc) *IReaderCloser {
	this.IPrinter.SetPrintFunc(fn)
	//错误信息按ASCII编码
	return this
}

func (this *IReaderCloser) Debug(b ...bool) *IReaderCloser {
	this.IPrinter.Debug(b...)
	this.ICloser.Debug(b...)
	return this
}

//================================DealFunc================================

// SetDealFunc 设置数据处理函数
func (this *IReaderCloser) SetDealFunc(fn func(msg Message)) {
	this.dealFunc = fn
}

// SetDealWithNil 不设置数据处理函数
func (this *IReaderCloser) SetDealWithNil() {
	this.SetDealFunc(nil)
}

// SetDealWithWriter 设置数据处理到io.Writer
func (this *IReaderCloser) SetDealWithWriter(writer Writer) {
	this.SetDealFunc(func(msg Message) {
		writer.Write(msg)
	})
}

// SetDealWithChan 设置数据处理到chan
func (this *IReaderCloser) SetDealWithChan(c chan Message) {
	this.SetDealFunc(func(msg Message) {
		c <- msg
	})
}

//================================RunTime================================

// Running 是否在运行
func (this *IReaderCloser) Running() bool {
	return this.running == 1
}

func (this *IReaderCloser) Run() error {
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
