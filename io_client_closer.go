package io

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

// Redial 重新链接,重试,因为指针复用,所以需要根据上下文来处理(例如关闭)
func (this *Client) Redial(options ...OptionClient) *Client {
	this.SetCloseFunc(func(ctx context.Context, c *Client, err error) {
		//等待1秒之后开始重连,防止无限制连接断开
		<-time.After(time.Second)
		if err := this.MustDial(this.ctxParent, func(c *Client) { c.Redial(options...) }); err != nil {
			this.Errorf("[%s] 重连错误,%v\n", this.GetKey(), err)
			return
		}
	})
	this.SetOptions(options...)
	//新建客户端时已经能确定连接成功,为了让用户控制是否输出,所以在Run的时候打印
	//this.Logger.Infof("[%s] 连接服务端成功...\n", this.GetKey())
	go this.Run()
	return this
}

// SetRedialMaxTime 设置退避重试时间,默认32秒,需要连接成功的后续重连生效
func (this *Client) SetRedialMaxTime(t time.Duration) *Client {
	if t > 0 {
		this.redialMaxTime = t
	}
	return this
}

// SetRedialMaxNum 设置最大重连次数
func (this *Client) SetRedialMaxNum(n int) *Client {
	this.redialMaxNum = n
	return this
}

//================================DialFunc================================

// SetDialFunc 设置连接函数
func (this *Client) SetDialFunc(fn DialFunc) *Client {
	this.dialFunc = fn
	return this
}

// SetDialWithNil 设置连接函数为nil
func (this *Client) SetDialWithNil() *Client {
	this.SetDialFunc(nil)
	return this
}

//================================Timer================================

// After 延迟执行函数
func (this *Client) After(after time.Duration, fn func()) {
	select {
	case <-this.Done():
		return
	case <-time.After(after):
		fn()
	}
}

// Timer 定时器执行函数,直到错误
func (this *Client) Timer(interval time.Duration, fn func() error) {
	this.timer(this.Ctx(), this.CloseWithErr, interval, fn)
}

// timer 定时器
func (this *Client) timer(ctx context.Context, dealErr func(error) error, interval time.Duration, fn func() error) {
	if interval > 0 {
		timer := time.NewTimer(interval)
		defer timer.Stop()
		for i := 0; ; i++ {
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
func (this *Client) For(fn func(ctx context.Context) error) (err error) {
	for {
		select {
		//case <-this.DoneAll():
		//1. 调用了CloseAll方法进行关闭
		//2. 通过外部的上下文Cancel进行关闭,需要进行断开连接操作
		//这里执行不到,没有数据接收会一直阻塞
		//logs.Debug("DoneAll")
		//this.CloseAll()
		//return this.Err()
		case <-this.Done():
			//1. 调用了Close方法
			//2. 连接报错触发了Close
			return this.Err()
		default:
			_ = this.CloseWithErr(func() (err error) {
				defer func() {
					if e := recover(); e != nil {
						err = fmt.Errorf("%v", e)
					}
				}()
				return fn(this.Ctx())
			}())
		}
	}
}

//================================RunTime================================

// CtxAll 父级上下文,生命周期(客户端)
func (this *Client) CtxAll() context.Context {
	return this.ctxParent
}

// DoneAll 全部结束,关闭信号,一定有错误,只能手动或者上下文
func (this *Client) DoneAll() <-chan struct{} {
	return this.CtxAll().Done()
}

// Ctx 子级上下文,生命周期(单次连接)
func (this *Client) Ctx() context.Context {
	return this.ctx
}

// Done 结束,关闭信号,一定有错误
func (this *Client) Done() <-chan struct{} {
	return this.Ctx().Done()
}

// Err 错误信息
func (this *Client) Err() error {
	return this.closeErr
}

// Closed 是否已关闭
func (this *Client) Closed() bool {
	//方便业务逻辑 xxx==nil || xxx.Closed
	//对象是nil,也能调用对象的方法,不能调用对象的字段
	if this == nil {
		return true
	}
	select {
	case <-this.Done():
		//确保错误信息closeErr已经赋值,不用this.closed==1
		return true
	default:
		return false
	}
}

// CloseAll 主动关闭,不会重试
func (this *Client) CloseAll() error {
	return this.CloseAllWithErr(ErrHandClose)
}

func (this *Client) CloseAllWithErr(err error) error {
	if err == nil {
		return nil
	}
	//关闭重试函数
	this.SetDialWithNil()
	//关闭父级上下文
	this.cancelParent()
	//关闭子级
	return this.CloseWithErr(ErrHandClose)
}

// Close 主动关闭,会重试(如果设置了重连)
func (this *Client) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// TryCloseWithDeadline 尝试使用Deadline关闭,例如net.Conn
func (this *Client) TryCloseWithDeadline() error {
	return this.closeWithErr(ErrHandClose, func(closer Closer) error {
		switch c := closer.(type) {
		case interface{ SetDeadline(t time.Time) error }:
			//例如net.Conn,表示没有数据读或者写,则关闭连接
			return c.SetDeadline(time.Time{})
		default:
			return closer.Close()
		}
	})
}

// CloseWithErr 根据错误关闭,会重试(如果设置了重连)
func (this *Client) CloseWithErr(closeErr error) (err error) {
	return this.closeWithErr(closeErr)
}

// closeWithErr 根据错误关闭,会重试(如果设置了重连),自定义关闭函数
func (this *Client) closeWithErr(closeErr error, fn ...func(Closer) error) (err error) {
	if closeErr != nil {
		//原子判断是否执行过
		if atomic.SwapUint32(&this.closed, 1) == 1 {
			return
		}
		//先赋值错误,再赋值关闭,确保关闭后一定有错误信息
		this.closeErr = dealErr(closeErr)
		//关闭子级上下文
		this.cancel()
		//关闭写队列
		if this.writeQueue != nil {
			close(this.writeQueue)
		}
		//关闭实例,可自定义关闭方式,例如设置超时
		if len(fn) == 0 && this.i != nil {
			err = this.i.Close()
		} else {
			for _, v := range fn {
				err = v(this.i)
			}
		}
		//生成错误信息
		//msg := Message(this.closeErr.Error())
		////打印错误信息
		//this.logger.Errorf("[%s] %s\n", this.GetKey(), msg.String())
		this.logger.Errorf("[%s] 断开连接: %v\n", this.GetKey(), this.closeErr)

		//执行用户设置的错误函数,需要最后执行,防止后续操作无法执行,如果设置了重连不会执行到下一步
		if this.closeFunc != nil {
			this.closeFunc(this.CtxAll(), this, this.closeErr)
		}

		////执行用户设置的错误函数
		//if this.closeFunc != nil {
		//	//需要最后执行,防止后续操作无法执行,如果设置了重连不会执行到下一步
		//	this.closeFunc(this.CtxAll(), this, msg)
		//}
		//如果执行到了这里,则说明没有重试,则关闭父级上下文
		//重复关闭会怎么样? 测试可以重复关闭
		//logs.Debug("cancelParent")
		//this.cancelParent()
	}
	return
}

// MustDial 无限重连,返回错误信息
func (this *Client) MustDial(ctx context.Context, options ...OptionClient) error {
	t := time.Second
	timer := time.NewTimer(0)
	defer timer.Stop()
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return errors.New("上下文关闭")

		case <-timer.C:
			if this.dialFunc == nil {
				//未设置重连函数
				//this.Errorf("[%s] 连接断开(%v),未设置重连函数\n", this.GetKey(), this.Err())
				return errors.New("未设置重连函数")
			}
			err := this.Dial(options...)
			if err == nil {
				return nil
			}
			//判断是否超出重试次数
			if this.redialMaxNum > 0 && i >= this.redialMaxNum {
				return err
			}
			t *= 2
			if t > this.redialMaxTime {
				t = this.redialMaxTime
			}

			this.Logger.Errorf("[%s] %v,等待%d秒重试\n", this.GetKey(), dealErr(err), t/time.Second)
			timer.Reset(t)
		}
	}
}

// Dial 连接,成功后执行Option
func (this *Client) Dial(options ...OptionClient) error {

	if this.dialFunc == nil {
		//未设置重连函数
		return errors.New("无效连接函数")
	}

	select {
	case <-this.ctxParent.Done():
		//外部上下文关闭,关闭并取消重试
		_ = this.CloseAll()
		return errors.New("上下文关闭")

	default:

		//尝试进行连接,返回ReadWriteCloser和唯一标识key
		i, key, err := this.dialFunc(this.ctx)
		if err != nil {
			if len(key) > 0 {
				//尝试设置key,如果错误也返回key的话
				this.SetKey(key)
			}
			return err
		}

		//数据初始化操作,声明内存等操作
		this.reset(i, key, options...)

		//判断初始化操作是否出现错误,出现错误则返回错误
		//例如开始连接成功,后来失败了,或者option中关闭了连接
		if this.Err() != nil {
			return this.Err()
		}

		//连接成功事件
		for _, f := range this.connectFunc {
			if err := this.CloseWithErr(f(this)); err != nil {
				return err
			}
		}

		return nil

		//转移至连接事件
		//if !this.Closed() {
		//	//如果在option执行了关闭方法的话,会出现先打印关闭,再打印连接成功,
		//	//所以判断下连接是否还在,再决定是否打印连接成功的信息
		//	//至于为什么放option后面,是为了让用户能控制日志的打印
		//	this.Infof("[%s] 连接服务端成功...\n", this.GetKey())
		//}

	}
}
