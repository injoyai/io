package io

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func NewClientCloser(closer Closer) *ClientCloser {
	return &ClientCloser{
		ClientPrinter: NewClientPrint(),
		closer:        closer,
		redialFunc:    nil,
		closeFunc:     nil,
		closed:        &atomic.Value{},
	}
}

type ClientCloser struct {
	*ClientPrinter
	closer     Closer
	redialFunc func() (ReadWriteCloser, error)
	closeFunc  func(msg *Message)
	mu         sync.Mutex
	closed     *atomic.Value
}

// SetCloseFunc 设置关闭函数
func (this *ClientCloser) SetCloseFunc(fn func(msg *Message)) {
	this.closeFunc = fn
}

// SetCloseWithNil 设置无关闭函数
func (this *ClientCloser) SetCloseWithNil() {
	this.SetCloseFunc(nil)
}

// SetCloseWithRedial 设置关闭后重连
func (this *ClientCloser) SetCloseWithRedial(fn ...func(closer *ClientCloser)) {
	this.SetCloseFunc(func(msg *Message) {
		if this.redialFunc == nil {
			this.ClientPrinter.Print(TagClose, msg)
			return
		}
		this.ClientPrinter.Print(TagRedial, msg)
		<-time.After(time.Second)
		//todo
		for _, v := range fn {
			v(this)
		}
	})

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
		if t < time.Second*32 {
			t = 2 * t
		}
		this.ClientPrinter.Print(TagErr, NewMessage([]byte(fmt.Sprintf("%v,等待%d秒重试", dealErr(err), t))))
		<-time.After(t)
	}
}

// Err 错误信息
func (this *ClientCloser) Err() error {
	v := this.closed.Load()
	if v != nil {
		return v.(error)
	}
	return nil
}

// Closed 是否已关闭
func (this *ClientCloser) Closed() bool {
	return this.Err() != nil
}

// Close 关闭
func (this *ClientCloser) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭
func (this *ClientCloser) CloseWithErr(closeErr error) (err error) {
	this.mu.Lock()
	if closeErr != nil {
		this.closed.Store(closeErr)
		err = this.closer.Close()
		msg := NewMessage([]byte(closeErr.Error()))
		if this.closeFunc != nil {
			//需要最后执行,防止死锁
			defer this.closeFunc(msg)
		}
		this.ClientPrinter.Print(TagClose, msg)
	}
	this.mu.Unlock()
	return
}
