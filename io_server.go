package io

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
	"sync/atomic"
	"time"
)

func NewServer(newListen ListenFunc, options ...OptionServer) (*Server, error) {
	return NewServerWithContext(context.Background(), newListen, options...)
}

func NewServerWithContext(ctx context.Context, newListen func() (Listener, error), options ...OptionServer) (*Server, error) {
	//连接listener
	listener, err := newListen()
	if err != nil {
		return nil, err
	}
	//新建实例
	s := &Server{
		printer:      newPrinter(fmt.Sprintf("%p", listener)),
		ICloser:      NewICloserWithContext(ctx, listener),
		ClientManage: NewClientManage(ctx),
		Tag:          maps.NewSafe(),
		listener:     listener,
	}
	//设置默认打印函数,打印基础信息
	//s.SetPrintWithBase()
	//开启基础信息打印
	s.Debug()
	//设置关闭函数
	s.ICloser.SetCloseFunc(func(ctx context.Context, msg Message) {
		//关闭listener
		s.listener.Close()
		//关闭已连接的客户端,关闭listener后,客户端还能正常通讯
		s.ClientManage.CloseClientAll()
	})
	//设置前置函数
	s.ClientManage.SetBeforeFunc(func(c *Client) error {
		//默认连接打印信息
		s.Print(NewMessage("新的客户端连接..."), TagInfo, c.GetKey())
		return nil
	})
	//预设服务处理
	s.SetOptions(options...)
	return s, nil
}

// Server 服务端
type Server struct {
	*printer
	*ICloser
	*ClientManage

	Tag      *maps.Safe //tag
	listener Listener   //listener
	running  uint32     //是否在运行
}

//================================SetFunc================================

// SetOptions 设置选项
func (this *Server) SetOptions(options ...OptionServer) *Server {
	for _, v := range options {
		v(this)
	}
	return this
}

// SetPrintFunc 设置打印方式
func (this *Server) SetPrintFunc(fn PrintFunc) *Server {
	this.printer.SetPrintFunc(fn)
	this.printFunc = fn
	return this
}

// SetPrintWithHEX 设置打印方式HEX
func (this *Server) SetPrintWithHEX() *Server {
	return this.SetPrintFunc(PrintWithHEX)
}

// SetPrintWithASCII 设置打印方式ASCII
func (this *Server) SetPrintWithASCII() *Server {
	return this.SetPrintFunc(PrintWithASCII)
}

// SetPrintWithBase 设置打印方式ASCII,打印基础信息
func (this *Server) SetPrintWithBase() *Server {
	return this.SetPrintFunc(PrintWithBase)
}

// SetPrintWithErr 设置打印方式ASCII,打印错误信息
func (this *Server) SetPrintWithErr() *Server {
	return this.SetPrintFunc(PrintWithErr)
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

func (this *Server) SetCloseFunc(fn func(msg *IMessage)) *Server {
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

//// SwapServer 和另一个服务交换数据,客户端都的话存在数据重复发送客户端和速度瓶颈
//func (this *Server) SwapServer(s *Server) *Server {
//	this.SetDealWithWriter(s)
//	s.SetDealWithWriter(this)
//	return this
//}

//================================RunTime================================

// Running 是否在运行
func (this *Server) Running() bool {
	return atomic.LoadUint32(&this.running) == 1
}

// Run 运行(监听)
func (this *Server) Run() error {

	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}

	this.Print(NewMessage("开启服务成功..."), TagInfo, this.GetKey())

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
		x := NewClientWithContext(this.Ctx(), c)
		x.SetKey(key)                               //设置唯一标识符
		x.Debug(this.GetDebug())                    //调试模式
		x.SetPrintFunc(this.printer.GetPrintFunc()) //设置打印函数

		this.ClientManage.SetClient(x)

	}
}

//================================Inside================================
