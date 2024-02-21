package io

import (
	"context"
	"github.com/injoyai/base/chans"
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
		IReader:      NewIReader(readCloser),
		ICloser:      NewICloserWithContext(ctx, readCloser),
		running:      0,
		timeout:      0,
		timeoutReset: make(chan struct{}),
	}
}

type IReadCloser struct {
	*IReader
	*ICloser
	dealFunc     func(msg Message) //处理数据函数
	running      uint32            //是否在运行
	timeout      time.Duration     //超时时间,读取
	timeoutReset chan struct{}     //超时重置
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

// SetReadIntervalTimeout 设置读取间隔超时时间,需要在Run之前设置
func (this *IReadCloser) SetReadIntervalTimeout(timeout time.Duration) *IReadCloser {
	this.timeout = timeout
	return this
}

//================================Log================================

// Debug debug模式,实现Debugger接口,不用返回值
func (this *IReadCloser) Debug(b ...bool) {
	this.IReader.Logger.Debug(b...)
	this.ICloser.Logger.Debug(b...)
}

func (this *IReadCloser) SetLogger(logger Logger) *IReadCloser {
	l := NewLogger(logger)
	this.IReader.Logger = l
	this.ICloser.Logger = l
	return this
}

func (this *IReadCloser) SetLevel(level Level) *IReadCloser {
	this.IReader.Logger.SetLevel(level)
	this.ICloser.Logger.SetLevel(level)
	return this
}

// SetPrintWithHEX 设置打印HEX
func (this *IReadCloser) SetPrintWithHEX() *IReadCloser {
	this.IReader.Logger.SetPrintWithHEX()
	this.ICloser.Logger.SetPrintWithHEX()
	return this
}

func (this *IReadCloser) SetPrintWithUTF8() *IReadCloser {
	this.IReader.Logger.SetPrintWithUTF8()
	this.ICloser.Logger.SetPrintWithUTF8()
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

// SetDealWithQueue 设置协程队列处理数据
// @num 协程数量
// @fn 处理函数
func (this *IReadCloser) SetDealWithQueue(num int, fn func(msg Message)) *IReadCloser {
	queue := chans.NewEntity(num).SetHandler(func(ctx context.Context, no, count int, data interface{}) {
		fn(data.(Message))
	})
	this.SetDealFunc(func(msg Message) { queue.Do(msg) })
	return this
}

//================================RunTime================================

// Running 是否在运行
func (this *IReadCloser) Running() bool {
	return atomic.LoadUint32(&this.running) == 1
}

// Run 开始运行数据读取
func (this *IReadCloser) Run() error {

	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}

	//todo is a good idea ?
	if this.timeout > 0 {
		go func(ctx context.Context) {
			timer := time.NewTimer(this.timeout)
			defer timer.Stop()
			for {
				timer.Reset(this.timeout)
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					_ = this.CloseWithErr(ErrWithReadTimeout)
					return
				case <-this.timeoutReset:
				}
			}
		}(this.ICloser.Ctx())
	}

	readFunc := func(ctx context.Context) (err error) {
		//读取数据
		bs, err := this.ReadMessage()
		if err != nil || len(bs) == 0 {
			return err
		}
		//尝试加入通道,超时定时器重置
		select {
		case this.timeoutReset <- struct{}{}:
		default:
		}
		//处理数据
		if this.dealFunc != nil {
			this.dealFunc(bs)
		}
		return nil
	}

	return this.For(readFunc)

}
