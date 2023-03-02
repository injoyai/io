package io

import (
	"context"
	"fmt"
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
		printer:       newPrinter(""),
		closer:        closer,
		redialFunc:    nil,
		redialMaxTime: time.Second * 32,
		closeFunc:     nil,
		closeErr:      nil,
		closed:        0,
		ctx:           ctxChild,
		cancel:        cancelChild,
		ctxParent:     ctxParent,
		cancelParent:  cancelParent,
	}
}

type ICloser struct {
	*printer                         //打印
	closer        Closer             //实例
	redialFunc    DialFunc           //重连函数
	redialMaxTime time.Duration      //最大尝试退避重连时间
	closeFunc     CloseFunc          //关闭函数
	closeErr      error              //错误信息
	closed        uint32             //是否关闭(不公开,做原子操作),0是未关闭,1是已关闭
	ctx           context.Context    //子级上下文
	cancel        context.CancelFunc //子级上下文
	ctxParent     context.Context    //父级上下文,主动关闭时,用于关闭redial
	cancelParent  context.CancelFunc //父级上下文,主动关闭时,用于关闭redial
}

//================================CloseFunc================================

// SetCloseFunc 设置关闭函数
func (this *ICloser) SetCloseFunc(fn func(ctx context.Context, msg Message)) *ICloser {
	this.closeFunc = fn
	return this
}

// SetCloseWithNil 设置无关闭函数
func (this *ICloser) SetCloseWithNil() *ICloser {
	this.SetCloseFunc(nil)
	return this
}

// SetRedialMaxTime 设置退避重试时间,默认32秒
func (this *ICloser) SetRedialMaxTime(t time.Duration) *ICloser {
	if t <= 0 {
		t = time.Second * 32
	}
	this.redialMaxTime = t
	return this
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

//================================GoFor================================

// GoTimerParent 协程,定时器执行函数,生命周期(客户端关闭,CloseAll或上下文关闭)
func (this *ICloser) GoTimerParent(interval time.Duration, fn func() error) {
	go this.TimerParent(interval, fn)
}

// TimerParent 协程,定时器执行函数,生命周期(客户端关闭,CloseAll或上下文关闭)
func (this *ICloser) TimerParent(interval time.Duration, fn func() error) {
	this.timer(this.ParentCtx(), func(err error) error { return this.CloseAll() }, interval, fn)
}

// GoTimer 协程,定时器执行函数,生命周期(一次链接,单次连接断开)
func (this *ICloser) GoTimer(interval time.Duration, fn func() error) {
	go this.Timer(interval, fn)
}

// Timer 定时器执行函数,直到错误
func (this *ICloser) Timer(interval time.Duration, fn func() error) {
	this.timer(this.Ctx(), this.CloseWithErr, interval, fn)
}

// timer 定时器
func (this *ICloser) timer(ctx context.Context, dealErr func(error) error, interval time.Duration, fn func() error) {
	if interval > 0 {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for {
			timer.Reset(interval)
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				if err := fn(); err != nil {
					_ = dealErr(err)
					return
				}
			}
		}
	}
}

// For 循环执行
func (this *ICloser) For(fn func() error) (err error) {
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
				return fn()
			}())
		}
	}
}

//================================RunTime================================

// ParentCtx 父级上下文(客户端)
func (this *ICloser) ParentCtx() context.Context {
	return this.ctxParent
}

func (this *ICloser) DoneAll() <-chan struct{} {
	return this.ParentCtx().Done()
}

// Ctx 子级上下文,生命周期(单次连接)
func (this *ICloser) Ctx() context.Context {
	return this.ctx
}

// Done 结束,关闭信号,一定有错误
func (this *ICloser) Done() <-chan struct{} {
	return this.Ctx().Done()
}

// Err 错误信息
func (this *ICloser) Err() error {
	return this.closeErr
}

// Closed 是否已关闭
func (this *ICloser) Closed() bool {
	select {
	case <-this.Done():
		//确保错误信息closeErr已经赋值,不用this.closed==1
		return true
	default:
		return false
	}
}

// CloseAll 主动关闭,不会重试
func (this *ICloser) CloseAll() error {
	//关闭重试函数
	this.SetCloseWithNil()
	//关闭父级上下文
	this.cancelParent()
	//关闭子级
	return this.CloseWithErr(ErrHandClose)
}

// Close 主动关闭,会重试(如果设置了重连)
func (this *ICloser) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭,会重试(如果设置了重连)
func (this *ICloser) CloseWithErr(closeErr error) (err error) {
	if closeErr != nil {
		//原子判断是否执行过
		if atomic.SwapUint32(&this.closed, 1) == 1 {
			return
		}
		//先赋值错误,再赋值关闭,确保关闭后一定有错误信息
		this.closeErr = dealErr(closeErr)
		//关闭子级上下文
		this.cancel()
		//关闭实例
		err = this.closer.Close()
		//生成错误信息
		msg := NewMessage(this.closeErr.Error())
		//打印错误信息
		this.printer.Print(msg, TagErr, this.GetKey())
		//执行用户设置的错误函数
		if this.closeFunc != nil {
			//需要最后执行,防止后续操作无法执行,如果设置了重连不会执行到下一步
			this.closeFunc(this.ctxParent, msg)
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
			t *= 2
			if t > this.redialMaxTime {
				t = this.redialMaxTime
			}
			timer.Reset(t)
		}
	}
}
