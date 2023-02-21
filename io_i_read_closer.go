package io

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// NewIReadCloser 新建IReader,默认读取函数ReadAll
func NewIReadCloser(readCloser ReadCloser) *IReadCloser {
	return NewIReadCloserWithContext(context.Background(), readCloser)
}

// NewIReadCloserWithContext 新建IReader,默认读取函数ReadAll
func NewIReadCloserWithContext(ctx context.Context, readCloser ReadCloser) *IReadCloser {
	if c, ok := readCloser.(*IReadCloser); ok && c != nil {
		return c
	}
	return &IReadCloser{
		IReader:  NewIReader(readCloser),
		ICloser:  NewICloserWithContext(ctx, readCloser),
		running:  0,
		timeout:  0,
		readChan: make(chan struct{}),
	}
}

type IReadCloser struct {
	*IReader
	*ICloser
	dealFunc DealFunc      //处理数据函数
	running  uint32        //是否在运行
	timeout  time.Duration //超时时间
	readChan chan struct{} //读取到数据信号,配合超时机制使用
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
	this.ICloser.SetPrintFunc(fn) //错误信息按ASCII编码?
	return this
}

// Debug debug模式
func (this *IReadCloser) Debug(b ...bool) *IReadCloser {
	this.IReader.Debug(b...)
	this.ICloser.Debug(b...)
	return this
}

// SetTimeout 设置超时时间,需要在Run之前设置
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

	//todo is a good idea ?
	if this.timeout > 0 {
		go func() {
			timer := time.NewTimer(this.timeout)
			defer timer.Stop()
			for {
				timer.Reset(this.timeout)
				select {
				case <-timer.C:
					_ = this.CloseWithErr(ErrWithReadTimeout)
				case <-this.readChan:
				}
			}
		}()
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
				//读取数据
				bytes, err := this.ReadMessage()
				if err != nil || len(bytes) == 0 {
					return err
				}
				select {
				case this.readChan <- struct{}{}:
				default:
				}
				//处理数据
				if this.dealFunc != nil {
					this.dealFunc(bytes)
				}
				return
			}())
		}
	}
}
