package io

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
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
	//新建实例
	s := &Server{
		Key:          Key(key),
		Logger:       defaultLogger(),
		ICloser:      NewICloserWithContext(ctx, listener),
		ClientManage: NewClientManage(ctx, key),
		tag:          maps.NewSafe(),
		listener:     listener,
	}
	s.SetLogger(s.Logger)
	//开启基础信息打印
	s.Debug()
	//设置关闭函数
	s.ICloser.SetCloseFunc(func(ctx context.Context, msg Message) {
		//关闭listener
		s.listener.Close()
		//关闭已连接的客户端,关闭listener后,客户端还能正常通讯
		s.ClientManage.CloseClientAll()
	})
	//预设服务处理
	s.SetOptions(options...)
	return s, nil
}

// Server 服务端
type Server struct {
	Key
	*ICloser
	*ClientManage

	Logger    *logger
	tag       *maps.Safe //tag
	listener  Listener   //listener
	running   uint32     //是否在运行
	startTime time.Time  //运行时间
	closeTime time.Time  //关闭时间
}

//================================Logger================================

func (this *Server) Debug(b ...bool) {
	this.Logger.Debug(b...)
	this.ICloser.Logger.Debug(b...)
	this.ClientManage.Logger.Debug(b...)
}

func (this *Server) SetLogger(logger Logger) *Server {
	l := newLogger(logger)
	this.Logger = l
	this.ICloser.Logger = l
	this.ClientManage.Logger = l
	return this
}

func (this *Server) SetPrintWithUTF8() *Server {
	this.Logger.SetPrintWithUTF8()
	this.ICloser.Logger.SetPrintWithUTF8()
	this.ClientManage.Logger.SetPrintWithUTF8()
	return this
}

func (this *Server) SetPrintWithHEX() *Server {
	this.Logger.SetPrintWithHEX()
	this.ICloser.Logger.SetPrintWithHEX()
	this.ClientManage.Logger.SetPrintWithHEX()
	return this
}

func (this *Server) SetLevel(level Level) *Server {
	this.Logger.SetLevel(level)
	this.ICloser.Logger.SetLevel(level)
	this.ClientManage.Logger.SetLevel(level)
	return this
}

// SetPrintWithAll 设置打印等级为全部
func (this *Server) SetPrintWithAll() *Server {
	return this.SetLevel(LevelAll)
}

// SetPrintWithBase 设置打印ASCII,基础信息
func (this *Server) SetPrintWithBase() *Server {
	return this.SetLevel(LevelInfo)
}

// SetPrintWithErr 设置打印错误信息
func (this *Server) SetPrintWithErr() *Server {
	return this.SetLevel(LevelError)
}

//================================Nature================================

func (this *Server) Tag() *maps.Safe {
	return this.tag
}

func (this *Server) SetTag(key, value interface{}) *Server {
	this.Tag().Set(key, value)
	return this
}

func (this *Server) GetTag(key interface{}) (interface{}, bool) {
	return this.Tag().Get(key)
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

// Timer 定时执行
func (this *Server) Timer(interval time.Duration, do OptionServer) {
	go this.ICloser.Timer(interval, func() error {
		do(this)
		return nil
	})
}

func (this *Server) Close() error {
	return this.ICloser.Close()
}

func (this *Server) SetCloseFunc(fn func(c *Client, msg Message)) *Server {
	this.ClientManage.SetCloseFunc(fn)
	return this
}

// Swap 和一个IO交换数据
func (this *Server) Swap(i ReadWriteCloser) *Server {
	c := NewClient(i)
	c.SetReadWithAll()
	return this.SwapClient(c)
}

// SwapClient 和一个客户端交换数据
func (this *Server) SwapClient(c *Client) *Server {
	this.SetDealWithWriter(c)
	c.SetDealWithWriter(this)
	go c.Run()
	return this
}

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
	this.Logger.Infof("[%s] 开启服务成功...", this.GetKey())

	//执行监听连接
	for {
		select {
		case <-this.Done():
			return this.Err()
		default:
		}

		c, key, err := this.listener.Accept()
		if err != nil {
			this.CloseWithErr(err)
			return err
		}

		//新建客户端,并配置
		x := NewClientWithContext(this.Ctx(), c).SetKey(key)
		x.Tag().Set("address", key)
		this.Logger.Infof("[%s] 新的客户端连接...", key)
		this.ClientManage.SetClient(x)

	}
}
