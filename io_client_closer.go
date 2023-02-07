package io

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

func NewClientCloser(closer Closer) *ClientCloser {
	return NewClientCloserWithContext(context.Background(), closer)
}

func NewClientCloserWithContext(ctx context.Context, closer Closer) *ClientCloser {
	ctx, cancel := context.WithCancel(ctx)
	return &ClientCloser{
		ClientPrinter: NewClientPrint(),
		ClientKey:     NewClientKey(""),
		closer:        closer,
		redialFunc:    nil,
		closeFunc:     nil,
		closeErr:      nil,
		closed:        0,
		ctx:           ctx,
		cancel:        cancel,
	}
}

type ClientCloser struct {
	*ClientPrinter
	*ClientKey
	closer     Closer
	redialFunc DialFunc
	closeFunc  func(msg Message)
	mu         sync.Mutex
	closeErr   error  //错误信息
	closed     uint32 //是否关闭(不公开,做原子操作),0是未关闭,1是已关闭
	ctx        context.Context
	cancel     context.CancelFunc
}

// SetCloseFunc 设置关闭函数
func (this *ClientCloser) SetCloseFunc(fn func(msg Message)) {
	this.closeFunc = fn
}

// SetCloseWithNil 设置无关闭函数
func (this *ClientCloser) SetCloseWithNil() {
	this.SetCloseFunc(nil)
}

// SetRedialFunc 设置重连函数
func (this *ClientCloser) SetRedialFunc(fn func() (ReadWriteCloser, error)) {
	this.redialFunc = fn
}

// Redial 重连
func (this *ClientCloser) Redial() (ReadWriteCloser, error) {
	if this.redialFunc != nil {
		return this.redialFunc()
	}
	return nil, nil
}

// MustDial 无限重连,返回nil,或者成功数据
func (this *ClientCloser) MustDial() ReadWriteCloser {
	t := time.Second
	for {
		readWriteCloser, err := this.Redial()
		if err == nil {
			return readWriteCloser
		}
		this.ClientPrinter.Print(NewMessageFormat("%v,等待%d秒重试", dealErr(err), t/time.Second), TagErr, this.GetKey())
		<-time.After(t)
		if t < time.Second*32 {
			t = 2 * t
		}
	}
}

// Ctx 上下文
func (this *ClientCloser) Ctx() context.Context {
	return this.ctx
}

// Done 结束,关闭信号,一定有错误
func (this *ClientCloser) Done() <-chan struct{} {
	return this.ctx.Done()
}

// Err 错误信息
func (this *ClientCloser) Err() error {
	return this.closeErr
}

// Closed 是否已关闭
func (this *ClientCloser) Closed() bool {
	select {
	case <-this.ctx.Done():
		return true
	default:
		return false
	}
	//当这个为true时,错误有可能还没赋值,所以使用ctx.Done
	//return this.closed == 1
}

// Close 关闭
func (this *ClientCloser) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭
func (this *ClientCloser) CloseWithErr(closeErr error) (err error) {
	this.mu.Lock()
	if closeErr != nil {
		if atomic.SwapUint32(&this.closed, 1) == 1 {
			return
		}
		//先赋值错误,再赋值关闭,确保关闭后一定有错误信息
		this.closeErr = closeErr
		this.cancel()
		err = this.closer.Close()
		msg := NewMessage([]byte(closeErr.Error()))
		if this.closeFunc != nil {
			//需要最后执行,防止死锁
			defer this.closeFunc(msg)
		}
		this.ClientPrinter.Print(msg, TagClose, this.GetKey())
	}
	this.mu.Unlock()
	return
}
