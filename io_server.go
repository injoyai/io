package io

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"sync/atomic"
	"time"
)

func RunServer(newListen ListenFunc, options ...OptionServer) error {
	s, err := NewServer(newListen, options...)
	if err != nil {
		return err
	}
	return s.Run()
}

// NewServer 定义Listen,不用忘记运行Run,不要回出现能连接,服务无反应的情况
func NewServer(newListen ListenFunc, options ...OptionServer) (*Server, error) {
	return NewServerWithContext(context.Background(), newListen, options...)
}

func NewServerWithContext(ctx context.Context, newListen func() (Listener, error), options ...OptionServer) (*Server, error) {
	//连接listener
	listener, err := newListen()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%p", listener)
	logger := defaultLogger()
	//新建实例
	s := &Server{
		Key:          Key(key),
		logger:       logger,
		Closer:       safe.NewCloser(),
		ClientManage: NewClientManage(key, logger),
		ctx:          ctx,
		tag:          maps.NewSafe(),
		listener:     listener,
	}
	//开启基础信息打印
	s.Debug()
	//设置关闭函数
	s.Closer.SetCloseFunc(func() error {
		//关闭listener
		s.listener.Close()
		//关闭已连接的客户端,关闭listener后,客户端还能正常通讯
		s.ClientManage.Close()
		return nil
	})
	//预设服务处理
	s.SetOptions(options...)
	//运行超时机制
	go s.Keep.Run(ctx)
	return s, nil
}

// Server 服务端
type Server struct {
	Key
	*ClientManage
	*logger
	*safe.Closer
	ctx       context.Context
	tag       *maps.Safe //tag
	listener  Listener   //listener
	running   uint32     //是否在运行
	startTime time.Time  //运行时间
	closeTime time.Time  //关闭时间
}

//================================Nature================================

func (this *Server) SetLogger(logger Logger) *Server {
	this.logger = NewLogger(logger)
	return this
}

func (this *Server) Tag() *maps.Safe {
	if this.tag == nil {
		this.tag = maps.NewSafe()
	}
	return this.tag
}

func (this *Server) StartTime() time.Time {
	return this.startTime
}

func (this *Server) CloseTime() time.Time {
	return this.closeTime
}

//================================SetFunc================================

// SetOptions 设置选项
func (this *Server) SetOptions(options ...OptionServer) *Server {
	for _, v := range options {
		v(this)
	}
	return this
}

func (this *Server) Listener() Listener {
	return this.listener
}

func (this *Server) Close() error {
	return this.Closer.Close()
}

func (this *Server) SetCloseFunc(fn func(c *Client, err error)) *Server {
	this.ClientManage.SetOptions(func(c *Client) {
		c.SetCloseFunc(func(ctx context.Context, c *Client, err error) {
			fn(c, err)
		})
	})
	return this
}

//// Swap 和一个IO交换数据
//func (this *Server) Swap(i ReadWriteCloser) *Server {
//	c := NewClient(i)
//	c.SetReadWith1KB()
//	return this.SwapClient(c)
//}
//
//// SwapClient 和一个客户端交换数据
//func (this *Server) SwapClient(c *Client) *Server {
//	this.ClientManage.SetOptions(func(c *Client) {
//
//	})
//	this.SetDealWithWriter(c)
//	c.SetDealWithWriter(this)
//	go c.Run()
//	return this
//}

//================================RunTime================================

// Running 是否在运行
func (this *Server) Running() bool {
	return atomic.LoadUint32(&this.running) == 1
}

// Run 运行(监听)
func (this *Server) Run() error {

	//判断是否在运行,防止重复运行
	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}

	//结束执行,修改运行状态和时间
	defer func() {
		atomic.StoreUint32(&this.running, 0)
		this.closeTime = time.Now()
	}()

	this.startTime = time.Now()
	this.Infof("[%s] 开启服务成功...\n", this.GetKey())

	//执行监听连接
	for {
		select {
		case <-this.ctx.Done():
			return this.Err()

		default:
		}

		c, key, err := this.listener.Accept()
		if err != nil {
			this.CloseWithErr(err)
			return this.Err() //使用最初的错误信息,否则会返回"use closed xxx"
			//return err
		}

		//新建客户端,并配置
		x := NewClientWithContext(this.ctx, c)
		x.SetLogger(this.logger)
		x.SetKey(key)
		x.Tag().Set("address", key)
		//this.Infof("[%s] 新的客户端连接...\n", key)
		this.ClientManage.SetClient(x)

	}
}
