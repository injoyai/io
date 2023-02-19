package io

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
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
	dealFunc DealFunc      //处理数据函数
	running  uint32        //是否在运行
	timeout  time.Duration //超时时间
}

//================================Nature================================

// GetKey 获取唯一标识
func (this *IReadCloser) GetKey() string {
	return this.IReader.GetKey()
}

// SetKey 设置唯一标识
func (this *IReadCloser) SetKey(key string) *IReadCloser {
	this.IReader.SetKey(key)
	this.ICloser.SetKey(key)
	return this
}

// SetPrintFunc 设置打印函数
func (this *IReadCloser) SetPrintFunc(fn PrintFunc) *IReadCloser {
	this.IReader.SetPrintFunc(fn)
	//错误信息按ASCII编码
	return this
}

// Debug debug模式
func (this *IReadCloser) Debug(b ...bool) *IReadCloser {
	this.IReader.Debug(b...)
	this.ICloser.Debug(b...)
	return this
}

// SetTimeout 设置超时时间
func (this *IReadCloser) SetTimeout(timeout time.Duration) *IReadCloser {
	this.timeout = timeout
	return this
}

//================================DealFunc================================

// SetDealFunc 设置数据处理函数
func (this *IReadCloser) SetDealFunc(fn func(msg Message)) *IReadCloser {
	this.dealFunc = fn
	return this
}

// SetDealWithNil 不设置数据处理函数
func (this *IReadCloser) SetDealWithNil() *IReadCloser {
	return this.SetDealFunc(nil)
}

// SetDealWithWriter 设置数据处理到io.Writer
func (this *IReadCloser) SetDealWithWriter(writer Writer) *IReadCloser {
	return this.SetDealFunc(func(msg Message) {
		writer.Write(msg)
	})
}

// SetDealWithChan 设置数据处理到chan
func (this *IReadCloser) SetDealWithChan(c chan Message) *IReadCloser {
	return this.SetDealFunc(func(msg Message) {
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

	timer := time.NewTimer(0)
	<-timer.C
	for {
		if this.timeout <= 0 {
			select {
			case <-this.Done():
				return this.Err()
			default:
				_ = this.CloseWithErr(this.run())
			}
		} else {
			select {
			case <-this.Done():
				return this.Err()
			case <-timer.C:
				return ErrWithReadTimeout
			default:
				_ = this.CloseWithErr(this.run())
			}
		}
	}
}

func (this *IReadCloser) run() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	//读取数据
	bytes, err := this.ReadMessage()
	if err != nil || len(bytes) == 0 {
		return err
	}
	//处理数据
	if this.dealFunc != nil {
		this.dealFunc(bytes)
	}
	return
}
