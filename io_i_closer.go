package io

import (
	"context"
	"sync/atomic"
	"time"
)

func NewICloser(closer Closer) *ICloser {
	return NewICloserWithContext(context.Background(), closer)
}

func NewICloserWithContext(ctx context.Context, closer Closer) *ICloser {
	ctxParent, cancelParent := context.WithCancel(ctx)
	ctxChild, cancelChild := context.WithCancel(ctxParent)
	return &ICloser{
		IPrinter:     NewIPrinter(""),
		closer:       closer,
		redialFunc:   nil,
		closeFunc:    nil,
		closeErr:     nil,
		closed:       0,
		ctx:          ctxChild,
		cancel:       cancelChild,
		ctxParent:    ctxParent,
		cancelParent: cancelParent,
	}
}

type ICloser struct {
	*IPrinter                       //打印
	closer       Closer             //实例
	redialFunc   DialFunc           //重连函数
	closeFunc    CloseFunc          //关闭函数
	closeErr     error              //错误信息
	closed       uint32             //是否关闭(不公开,做原子操作),0是未关闭,1是已关闭
	ctx          context.Context    //子级上下文
	cancel       context.CancelFunc //子级上下文
	ctxParent    context.Context    //父级上下文,主动关闭时,用于关闭redial
	cancelParent context.CancelFunc //父级上下文,主动关闭时,用于关闭redial
}

//================================CloseFunc================================

// SetCloseFunc 设置关闭函数
func (this *ICloser) SetCloseFunc(fn func(ctx context.Context, msg Message)) {
	this.closeFunc = fn
}

// SetCloseWithNil 设置无关闭函数
func (this *ICloser) SetCloseWithNil() {
	this.SetCloseFunc(nil)
}

//================================RedialFunc================================

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

//================================RunTime================================

// ParentCtx 父级上下文
func (this *ICloser) ParentCtx() context.Context {
	return this.ctxParent
}

// Ctx 子级上下文
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
	case <-this.Ctx().Done():
		//确保错误信息closeErr已经赋值,不用this.closed==1
		return true
	default:
		return false
	}
}

// CloseAll 主动关闭,不会重试
func (this *ICloser) CloseAll() error {
	this.SetCloseWithNil()
	this.cancelParent()
	return this.CloseWithErr(ErrHandClose)
}

// Close 主动关闭,会重试
func (this *ICloser) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭
func (this *ICloser) CloseWithErr(closeErr error) (err error) {
	if closeErr != nil {
		if atomic.SwapUint32(&this.closed, 1) == 1 {
			return
		}
		//先赋值错误,再赋值关闭,确保关闭后一定有错误信息
		this.closeErr = dealErr(closeErr)
		//关闭子级上下文
		this.cancel()
		//关闭实例
		err = this.closer.Close()
		msg := NewMessage(this.closeErr.Error())
		this.IPrinter.Print(msg, TagErr, this.GetKey())
		if this.closeFunc != nil {
			//需要最后执行,防止后续操作无法执行
			defer this.closeFunc(this.ParentCtx(), msg)
		}
	}
	return
}

// Redial 无限重连,返回nil,或者成功数据
func (this *ICloser) Redial(ctx context.Context) ReadWriteCloser {
	t := time.Second
	timer := time.NewTimer(t)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			if this.redialFunc == nil {
				//未设置重连函数
				return nil
			}
			readWriteCloser, err := this.redialFunc()
			if err == nil {
				//上下文关闭
				return readWriteCloser
			}
			this.Print(NewMessageFormat("%v,等待%d秒重试", dealErr(err), t/time.Second), TagErr, this.GetKey())
			if t < time.Second*32 {
				t = 2 * t
				timer.Reset(t)
			}
		}
	}
}
