package io

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

func NewICloser(closer Closer) *ICloser {
	return NewICloserWithContext(context.Background(), closer)
}

func NewICloserWithContext(ctx context.Context, closer Closer) *ICloser {
	ctx2, cancel2 := context.WithCancel(ctx)
	return &ICloser{
		ClientPrinter: NewClientPrint(),
		ClientKey:     NewClientKey(""),
		closer:        closer,
		redialFunc:    nil,
		closeFunc:     nil,
		closeErr:      nil,
		closed:        0,
		ctx:           ctx2,
		cancel:        cancel2,
		ctxParent:     ctx,
	}
}

type ICloser struct {
	*ClientPrinter                    //打印
	*ClientKey                        //标识
	closer         Closer             //实例
	redialFunc     DialFunc           //重连函数
	closeFunc      CloseFunc          //关闭函数
	mu             sync.Mutex         //锁
	closeErr       error              //错误信息
	closed         uint32             //是否关闭(不公开,做原子操作),0是未关闭,1是已关闭
	ctx            context.Context    //上下文
	cancel         context.CancelFunc //上下文
	ctxParent      context.Context    //上下文
}

// SetCloseFunc 设置关闭函数
func (this *ICloser) SetCloseFunc(fn func(ctx context.Context, msg Message)) {
	this.closeFunc = fn
}

// SetCloseWithNil 设置无关闭函数
func (this *ICloser) SetCloseWithNil() {
	this.SetCloseFunc(nil)
}

// SetRedialFunc 设置重连函数
func (this *ICloser) SetRedialFunc(fn DialFunc) *ICloser {
	this.redialFunc = fn
	return this
}

// SetRedialWithNil 设置重连函数为nil
func (this *ICloser) SetRedialWithNil() *ICloser {
	this.SetRedialFunc(nil)
	return this
}

// Redial 无限重连,返回nil,或者成功数据
func (this *ICloser) Redial(ctx context.Context) ReadWriteCloser {
	t := time.Second
	timer := time.NewTimer(t)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			if this.redialFunc == nil {
				return nil
			}
			readWriteCloser, err := this.redialFunc()
			if err == nil {
				if readWriteCloser != nil {
					this.ClientPrinter.Print(NewMessageFormat("连接服务端成功..."), TagDial, this.GetKey())
				}
				return readWriteCloser
			}
			this.ClientPrinter.Print(NewMessageFormat("%v,等待%d秒重试", dealErr(err), t/time.Second), TagErr, this.GetKey())
			if t < time.Second*32 {
				t = 2 * t
				timer.Reset(t)
			}
		}
	}
}

// Ctx 上下文
func (this *ICloser) Ctx() context.Context {
	return this.ctx
}

// Done 结束,关闭信号,一定有错误
func (this *ICloser) Done() <-chan struct{} {
	return this.ctx.Done()
}

// Err 错误信息
func (this *ICloser) Err() error {
	return this.closeErr
}

// Closed 是否已关闭
func (this *ICloser) Closed() bool {
	select {
	case <-this.ctx.Done():
		return true
	default:
		return false
	}
	//当这个为true时,错误有可能还没赋值,所以使用ctx.Done
	//return this.closed == 1
}

// CloseAll 主动关闭,不会重试
func (this *ICloser) CloseAll() error {
	this.SetCloseWithNil()
	return this.CloseWithErr(ErrHandClose)
}

// Close 主动关闭,不会重试
func (this *ICloser) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭
func (this *ICloser) CloseWithErr(closeErr error) (err error) {
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
			defer this.closeFunc(this.ctxParent, msg)
		}
		this.ClientPrinter.Print(msg, TagClose, this.GetKey())
	}
	this.mu.Unlock()
	return
}
