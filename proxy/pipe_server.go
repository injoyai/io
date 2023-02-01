package proxy

import (
	"bufio"
	"context"
	"errors"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"log"
	"strconv"
	"time"
)

// NewServer 新建服务,里面包含了服务端,提供对外写入接口,默认base64加密
func NewServer(newListen func() (io.Listener, error)) (*Server, error) {
	return NewServerWithContext(context.Background(), newListen)
}

func NewServerWithContext(ctx context.Context, newListen func() (io.Listener, error)) (*Server, error) {
	pipeServer, err := io.NewServerWithContext(ctx, newListen)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		Server:   pipeServer,
		dealRead: nil,
		ctx:      ctx,
		cancel:   cancel,
		client:   maps.NewSafe(),
	}

	pipeServer.SetPrintFunc(DefaultPrint)
	pipeServer.SetReadFunc(DefaultRead)
	pipeServer.SetWriteFunc(DefaultWrite)
	pipeServer.SetTimeout(time.Minute * 3)
	pipeServer.SetBeforeFunc(s.beforeFunc)
	pipeServer.SetDealFunc(s.dealFunc)
	pipeServer.SetCloseFunc(s.closeFunc)
	return s, nil
}

type Server struct {
	Server    *io.Server
	dealRead  func(c *io.Client, msg *Message) error                        //处理读到的数据
	dealWrite func(r *Request, listenPort int, data []byte) ([]byte, error) //处理写入的数据,例如修改请求头...
	ctx       context.Context                                               //上下文
	cancel    context.CancelFunc                                            //下下文关闭
	client    *maps.Safe                                                    //请求的客户端连接
	debug     bool                                                          //
}

// 关闭该通道的所有连接
func (this *Server) closeFunc(m *io.ClientMessage) {}

// readFunc 读取客户端数据
func (this *Server) readFunc(r *Request, listenPort int) (err error) {
	this.printProxy("连接", r.Key(), "")
	defer func() {
		this.closeProxy(r, conv.String(err))
		this.client.Del(r.Key())
		r.Close()
		if err != nil {
			this.printProxy("关闭", r.Key(), err.Error())
		}
	}()
	buf := bufio.NewReader(r)
	for {
		select {
		case <-r.ctx.Done():
			return io.EOF
		default:
			data := make([]byte, 4096)
			length, err := buf.Read(data)
			if err != nil {
				return err
			}
			data = data[:length]
			this.printProxy("接收", r.Key(), string(data))
			if _, err = this.writeProxy(r, listenPort, data); err != nil {
				return err
			}
		}
	}
}

func (this *Server) printProxy(tag, srcAddr, msg string) {
	if this.debug {
		if msg == "EOF" {
			msg = ""
		}
		if len(msg) > 50 {
			msg = msg[:50] + "..." + strconv.Itoa(len(msg)) + "..."
		}
		log.Printf("[proxy][%s][%s] %s", tag, srcAddr, msg)
	}
}

// SetClient 新增请求客户端,会覆盖老的,并关闭老的
func (this *Server) SetClient(r *Request, listenPort int) *Server {
	old := this.client.GetAndSet(r.Key(), r)
	if old != nil {
		old.(*Request).Close()
	}
	go this.readFunc(r, listenPort)
	return this
}

// Debug 调试模式
func (this *Server) Debug(b ...bool) *Server {
	this.debug = !(len(b) > 0 && !b[0])
	return this
}

// SetDealWriteFunc 设置写入通道的数据,处理请求的数据,例如修改请求头,过滤非法请求等...
func (this *Server) SetDealWriteFunc(fn func(r *Request, listenPort int, data []byte) ([]byte, error)) *Server {
	this.dealWrite = fn
	return this
}

// SetDealReadFunc 设置读取通道的数据,处理读取到的数据
func (this *Server) SetDealReadFunc(fn func(c *io.Client, msg *Message) error) *Server {
	this.dealRead = fn
	return this
}

// Close 关闭通道服务
func (this *Server) Close() {
	if this.cancel != nil {
		this.cancel()
	}
}

// Run 开始通道服务
func (this *Server) Run() error {
	return this.Server.Run()
}

//===================================Inside===============================

// closeProxy 关闭代理连接
func (this *Server) closeProxy(r *Request, data string) {
	this.pipeWrite(r, NewCloseMsg(r.Key(), data))
}

// writeProxy 向通道客户端发送数据
func (this *Server) writeProxy(r *Request, listenPort int, data []byte) (n int, err error) {

	// HTTP代理
	//var isHTTP bool
	//data, isHTTP = dealHTTP(r, data)
	//if isHTTP {
	//	this.printProxy("发送", r.Key(), Connection)
	//	return 0, nil
	//}

	// TCP代理
	if this.dealWrite != nil {
		data, err = this.dealWrite(r, listenPort, data)
	}

	if err != nil {
		return this.pipeWrite(r, NewCloseMsg(r.Key(), err.Error()))
	}

	return this.pipeWrite(r, NewWriteMsg(r.Key(), r.Addr, data))
}

// pipeWrite 写入通道
func (this *Server) pipeWrite(r *Request, msg *Message) (int, error) {
	pipe := this.Server.GetClient(r.SN)
	if pipe == nil {
		return 0, ErrNoConnected
	}
	return pipe.WriteBytes(msg.Bytes())
}

// BeforeWithServer 通道服务的前置操作,打印客户端连接信息
func (this *Server) beforeFunc(c *io.Client) (err error) {
	defer func() {
		if err != nil {
			DefaultPrint("错误", c.GetKey(), []byte(err.Error()))
		}
	}()

	bytes, err := DefaultRead(c.Buffer())
	if err != nil {
		return err
	}
	m, err := DecodeMsg(bytes)
	if err != nil {
		return err
	}
	if m.Type != Register {
		return errors.New("注册失败")
	}
	this.Server.SetClientKey(c, m.Key)
	DefaultPrint("连接", c.GetKey(), []byte("通道客户端连接成功..."))
	return nil
}

// dealFunc 处理通道客户端过来的数据,解析数据,并向请求客户端写数据
func (this *Server) dealFunc(msg *io.ClientMessage) {
	// 解析通道客户端发送的数据
	m, err := DecodeMsg(msg.Bytes())
	if err != nil {
		return
	}
	if this.dealRead != nil && m.Type != Close {
		if err = this.dealRead(msg.Client, m); err != nil {
			m.Type = Close
			m.Data = err.Error()
		}
	}
	switch m.Type {
	case Register:
		// 通道客户端注册信息,修改连接key
		this.Server.SetClientKey(msg.Client, m.Key)
	case Write:
		// 向请求的客户端写数据
		c := this.client.GetInterface(m.Key)
		if c != nil {
			c.(*Request).Write([]byte(m.Data))
			this.printProxy("发送", m.Key, m.Data)
		}
	case Close:
		// 关闭请求的客户端
		c := this.client.GetInterface(m.Key)
		if c != nil {
			c.(*Request).Close()
		}
	}
}
