package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
)

// New 新建代理实例
func New() *Entity {
	return &Entity{
		ioMap:       maps.NewSafe(),
		connectFunc: DefaultConnectFunc,
		buff:        make(chan io.Message, 1000),
		printFunc:   io.PrintWithASCII,
	}
}

// Entity 代理实例,通过数据进行对应的操作(读取,写入,连接,关闭)
type Entity struct {
	ioMap       *maps.Safe                                           //存储连接
	connectFunc func(msg *Message) (i io.ReadWriteCloser, err error) //连接函数
	buff        chan io.Message                                      //
	debug       bool                                                 //
	printFunc   io.PrintFunc                                         //
}

// Debug 调试模式
func (this *Entity) Debug(b ...bool) *Entity {
	this.debug = !(len(b) > 0 && !b[0])
	return this
}

// SetPrintFunc 设置打印函数
func (this *Entity) SetPrintFunc(fn func(msg io.Message, tag ...string)) *Entity {
	this.printFunc = fn
	return this
}

// AddMessage 添加到message缓存,供ReadMessage读取
func (this *Entity) AddMessage(msg *Message) {
	this.buff <- msg.Bytes()
}

// Read 实现io.Reader,无效,使用ReadMessage
func (this *Entity) Read(p []byte) (n int, err error) {
	return 0, nil
}

// ReadMessage 实现接口 io.MessageReader
func (this *Entity) ReadMessage() (bs []byte, _ error) {
	defer func() {
		//logs.Debug(222, string(bs))
	}()
	return <-this.buff, nil
}

// Write 实现io.Writer 写入数据,解析数据,处理数据,代理格式
// 自动解析数据,并根据解析的内容进行效应的操作(例发起请求)
func (this *Entity) Write(p []byte) (int, error) {
	msg, err := DecodeMessage(p)
	if err != nil {
		return 0, err
	}
	return len(p), this.WriteMessage(msg)
}

// WriteMessage 需要并发处理,防止1个连接阻塞,导致后续请求都超时
func (this *Entity) WriteMessage(msg *Message) (err error) {
	return this.writeMessage(msg)
}

// WriteMessage 处理获取到的消息
func (this *Entity) writeMessage(msg *Message) (err error) {

	//查找该请求是否已经代理过
	proxyClient := this.getIO(msg.Key)

	//如果不存在,请求连接或写入,则重新建立连接
	//需要协程执行,避免阻塞后续请求
	if proxyClient == nil && (msg.OperateType == Connect || msg.OperateType == Write) {
		go func() {
			if this.connectFunc == nil {
				//设置默认连接函数
				this.connectFunc = DefaultConnectFunc
			}
			//使用连接建立连接
			i, err := this.connectFunc(msg)
			if err != nil {
				//连接失败,则响应关闭连接
				this.AddMessage(NewCloseMessage(msg.Key, err.Error()))
				return
			}

			proxyClient = io.NewClient(i)
			proxyClient.Debug(this.debug)
			proxyClient.SetPrintFunc(this.printFunc)
			proxyClient.SetKey(msg.Addr)
			proxyClient.SetReadFunc(buf.ReadWithAll)
			proxyClient.SetDealFunc(func(m *io.IMessage) {
				logs.Debug(111, m.String())
				this.AddMessage(msg.Response(m.Bytes()))
			})
			proxyClient.SetCloseFunc(func(ctx context.Context, m *io.IMessage) {
				this.delIO(msg.Key)
				this.AddMessage(NewCloseMessage(msg.Key, m.String()))
			})
			go proxyClient.Run()
			//加入到缓存
			this.setIO(msg.Key, proxyClient)
			if msg.OperateType == Write {
				proxyClient.Write(msg.GetData())
			}
		}()
		return
	}

	//如果不存在则结束
	if proxyClient == nil {
		return
	}

	switch msg.OperateType {
	case Connect:
		//收到建立连接信息
	case Write:
		//收到写数据信息
		_, err = proxyClient.Write(msg.GetData())
	case Close:
		//收到关闭连接信息
		err = proxyClient.TryCloseWithDeadline()
	}

	return
}

// Close 实现 io.Closer
func (this *Entity) Close() error {
	this.closeIOAll()
	return nil
}

// SetConnectFunc 设置连接函数
func (this *Entity) SetConnectFunc(fn func(msg *Message) (i io.ReadWriteCloser, err error)) *Entity {
	this.connectFunc = fn
	return this
}

// DefaultConnectFunc 默认连接函数
func DefaultConnectFunc(msg *Message) (i io.ReadWriteCloser, err error) {
	err = errors.New("未实现")
	switch msg.ConnectType {
	case TCP:
		i, err = dial.TCP(msg.Addr)
	case UDP:
		i, err = dial.UDP(msg.Addr)
	case Serial:
		cfg := new(dial.SerialConfig)
		err = json.Unmarshal([]byte(msg.Addr), cfg)
		if err != nil {
			return
		}
		i, err = dial.Serial(cfg)
	case File, MQ, MQTT, HTTP, Websocket:
		//待实现
	default:
		i, err = dial.TCP(msg.Addr)
	}
	return
}

func NewTCPClient(addr string, fn ...func(ctx context.Context, c *io.Client, e *Entity)) *io.Client {
	e := New()
	return dial.RedialPipe(addr, func(ctx context.Context, c *io.Client) {
		for _, v := range fn {
			v(ctx, c, e)
		}
		c.Swap(e)
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			//打印函数的处理,不重要
			switch true {
			case msg.String() == io.Ping || msg.String() == io.Pong:
			case len(tag) > 0:
				switch tag[0] {
				case io.TagWrite, io.TagRead:
					m, err := DecodeMessage(msg.Bytes())
					if err != nil {
						logs.Debug(err)
						return
					}
					logs.Debugf("[PI|C][%s] %s\n", tag[0], m.String())
				default:
					logs.Debugf(io.PrintfWithASCII(msg, append([]string{"PI|C"}, tag...)...))
				}
			}
		})
	})
}

// NewSwapTCPServer 和TCP服务端交换数据,带测试
func NewSwapTCPServer(port int, fn ...func(s *io.Server)) error {
	s, err := dial.NewPipeServer(port)
	if err != nil {
		return err
	}
	for _, v := range fn {
		v(s)
	}
	s.Swap(New())
	s.SetPrintFunc(func(msg io.Message, tag ...string) {
		logs.Debug(io.PrintfWithASCII(msg, append([]string{"PR|S"}, tag...)...))
	})
	go s.Run()
	return nil
}

func WithClientDebug(b ...bool) func(ctx context.Context, c *io.Client, e *Entity) {
	return func(ctx context.Context, c *io.Client, e *Entity) {
		c.Debug(b...)
		e.Debug()
	}
}

/*

	Inside

*/

// setIO 添加记录,存在则关闭并覆盖
func (this *Entity) setIO(key string, i *io.Client) {
	old := this.ioMap.GetAndSet(key, i)
	if val, ok := old.(*io.Client); ok {
		val.Close()
	}
}

// getOrSet 获取或者设置,尝试获取数据,不存在则设置
func (this *Entity) getOrSet(key string, i *io.Client) *io.Client {
	old := this.getIO(key)
	if old != nil {
		return old
	}
	this.ioMap.Set(key, i)
	return nil
}

// getIO 获取io,不存在或者类型错误则返回nil
func (this *Entity) getIO(key string) *io.Client {
	i, _ := this.ioMap.Get(key)
	if i == nil {
		return nil
	}
	//类型判断是否是需要的类型,是则返回
	if val, ok := i.(*io.Client); ok {
		return val
	}
	//如果记录存在,当类型错误,则删除记录
	this.delIO(key)
	return nil
}

// delIO 删除记录
func (this *Entity) delIO(key string) {
	this.ioMap.Del(key)
}

// closeIO 关闭io,删除记录据
func (this *Entity) closeIO(key string) {
	i := this.getIO(key)
	if i != nil {
		i.Close()
	}
	this.delIO(key)
}

// CloseIOAll 关闭全部io
func (this *Entity) closeIOAll() {
	this.ioMap.Range(func(key, value interface{}) bool {
		value.(*io.Client).TryCloseWithDeadline()
		return true
	})
	this.ioMap = maps.NewSafe()
}
