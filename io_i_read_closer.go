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
		IReader:  NewIReader(readCloser),
		ICloser:  NewICloserWithContext(ctx, readCloser),
		running:  0,
		timeout:  0,
		readSign: make(chan struct{}),
	}
}

type IReadCloser struct {
	*IReader
	*ICloser
	dealFunc func(msg Message) //处理数据函数
	running  uint32            //是否在运行
	timeout  time.Duration     //超时时间
	readSign chan struct{}     //读取到数据信号,配合超时机制使用
	queue    *chans.Entity     //协程队列,可选
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

// Debug debug模式
func (this *IReadCloser) Debug(b ...bool) *IReadCloser {
	this.IReader.Logger.Debug(b...)
	this.ICloser.Logger.Debug(b...)
	return this
}

// SetReadIntervalTimeout 设置读取间隔超时时间,需要在Run之前设置
func (this *IReadCloser) SetReadIntervalTimeout(timeout time.Duration) *IReadCloser {
	this.timeout = timeout
	return this
}

//================================Log================================

func (this *IReadCloser) SetLogger(logger Logger) *IReadCloser {
	this.IReader.Logger = logger
	this.ICloser.Logger = logger
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

func (this *IReadCloser) SetPrintWithASCII() *IReadCloser {
	this.IReader.Logger.SetPrintWithASCII()
	this.ICloser.Logger.SetPrintWithASCII()
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

// SetDealQueueFunc 设置协程队列处理数据
// @num 协程数量
// @no 协程序号
// @count 当前协程执行次数
// @msg 消息内容
func (this *IReadCloser) SetDealQueueFunc(num int, fn func(msg Message)) *IReadCloser {
	if this.queue == nil {
		this.queue = chans.NewEntity(num).SetHandler(func(ctx context.Context, no, count int, data interface{}) {
			fn(data.(Message))
		})
	} else {
		this.queue.SetNum(num)
	}
	this.SetDealFunc(func(msg Message) { this.queue.Do(msg) })
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
		go func() {
			timer := time.NewTimer(this.timeout)
			defer timer.Stop()
			for {
				timer.Reset(this.timeout)
				select {
				case <-timer.C:
					_ = this.CloseWithErr(ErrWithReadTimeout)
				case <-this.readSign:
				}
			}
		}()
	}

	readFunc := func(ctx context.Context) (err error) {
		//读取数据
		bs, err := this.ReadMessage()
		if err != nil || len(bs) == 0 {
			return err
		}
		//尝试加入通道,超时定时器重置
		select {
		case this.readSign <- struct{}{}:
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
