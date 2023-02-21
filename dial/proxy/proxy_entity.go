package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"github.com/injoyai/io/dial"
)

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

func (this *Entity) Debug(b ...bool) *Entity {
	this.debug = !(len(b) > 0 && !b[0])
	return this
}

func (this *Entity) SetPrintFunc(fn func(msg io.Message, tag ...string)) *Entity {
	this.printFunc = fn
	return this
}

// Resp 响应数据
func (this *Entity) Resp(msg *Message) {
	this.buff <- msg.Bytes()
}

// Read 实现io.Reader,无效,使用ReadMessage
func (this *Entity) Read(p []byte) (n int, err error) {
	return 0, nil
}

// ReadMessage 实现接口
func (this *Entity) ReadMessage() ([]byte, error) {
	return <-this.buff, nil
}

// Write 实现io.Writer 写入数据,解析数据,处理数据,代理格式
func (this *Entity) Write(p []byte) (int, error) {
	msg, err := DecodeMessage(p)
	if err != nil {
		return 0, err
	}
	return len(p), this.WriteMessage(msg)
}

// Close 实现io.Closer
func (this *Entity) Close() error {
	this.closeIOAll()
	return nil
}

// WriteMessage 处理获取到的消息
func (this *Entity) WriteMessage(msg *Message) (err error) {
	i := this.getIO(msg.Key)

	if i == nil && (msg.OperateType == Connect || msg.OperateType == Write) {
		if this.connectFunc == nil {
			this.connectFunc = DefaultConnectFunc
		}
		i, err = this.connectFunc(msg)
		if err != nil {
			return err
		}

		c := io.NewClient(i)
		c.Debug(this.debug)
		c.SetPrintFunc(this.printFunc)
		c.SetKey(msg.Addr)
		c.SetReadFunc(buf.ReadWithAll)
		c.SetDealFunc(func(m *io.IMessage) {
			this.Resp(NewWriteMessage(msg.Key, msg.Addr, m.Bytes()))
		})
		c.SetCloseFunc(func(ctx context.Context, m *io.IMessage) {
			this.delIO(msg.Key)
			this.Resp(NewCloseMessage(msg.Key, m.String()))
		})
		go c.Run()
		this.setIO(msg.Key, c)
		i = c
	}

	if i == nil {
		return
	}

	switch msg.OperateType {
	case Connect:
		//收到建立连接信息
	case Write:
		//收到写数据信息
		_, err = i.Write(msg.GetData())
	case Close:
		//收到关闭连接信息
		err = i.Close()
	}

	return
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
	case File:
	case MQ:
	case MQTT:
	case HTTP:
	case Websocket:
	default:
		i, err = dial.TCP(msg.Addr)
	}
	return
}

func SwapTCPClient(addr string, fn ...func(ctx context.Context, c *io.Client, e *Entity)) *io.Client {
	e := New()
	return io.Redial(dial.TCPFunc(addr), func(ctx context.Context, c *io.Client) {
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"P|C"}, tag...)...)
		})
		c.SetWriteFunc(DefaultWriteFunc)
		c.SetReadFunc(DefaultReadFunc)
		for _, v := range fn {
			v(ctx, c, e)
		}
		c.Swap(e)
	})
}

func SwapTCPServer(port int, fn ...func(s *io.Server)) error {
	s, err := io.NewServer(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.SetPrintFunc(func(msg io.Message, tag ...string) {
		io.PrintWithASCII(msg, append([]string{"P|S"}, tag...)...)
	})
	s.SetWriteFunc(DefaultWriteFunc)
	s.SetReadFunc(DefaultReadFunc)
	s.Debug()
	for _, v := range fn {
		v(s)
	}
	s.Swap(New())
	go s.Run()
	return nil
}

//func writeFunc(writeLen uint, writeBytes func(p []byte) (int, error), key, addr string, p []byte) (int, error) {
//
//	total := len(p)
//
//	//一次性发送全部数据,数据太多可能会影响到其他数据的时效
//	if writeLen == 0 {
//		msg := newWriteMessage(key, addr, p)
//		return writeBytes(msg.Bytes())
//	}
//
//	//分包发送,避免其他数据不能及时发送
//	for len(p) > 0 {
//		data := []byte(nil)
//		if len(p) > int(writeLen) {
//			data = p[:writeLen]
//			p = p[writeLen:]
//		} else {
//			data = p[:]
//			p = p[:0]
//		}
//		msg := newWriteMessage(key, addr, data)
//		if _, err := writeBytes(msg.Bytes()); err != nil {
//			return 0, err
//		}
//	}
//
//	return total, nil
//}

// setIO 添加记录,存在则关闭并覆盖
func (this *Entity) setIO(key string, i io.ReadWriteCloser) {
	old := this.ioMap.GetAndSet(key, i)
	if val, ok := old.(io.Closer); ok {
		val.Close()
	}
}

// getOrSet 获取或者设置,尝试获取数据,不存在则设置
func (this *Entity) getOrSet(key string, i io.ReadWriteCloser) io.ReadWriteCloser {
	old := this.getIO(key)
	if old != nil {
		return old
	}
	this.ioMap.Set(key, i)
	return nil
}

// getIO 获取io,不存在或者类型错误则返回nil
func (this *Entity) getIO(key string) io.ReadWriteCloser {
	i, _ := this.ioMap.Get(key)
	if i == nil {
		return nil
	}
	//类型判断是否是需要的类型,是则返回
	if val, ok := i.(io.ReadWriteCloser); ok {
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
		if val, ok := value.(io.Closer); ok {
			val.Close()
		}
		return true
	})
	this.ioMap = maps.NewSafe()
}
