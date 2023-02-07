package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"github.com/injoyai/io/dial"
	"sync"
	"time"
)

func New() *Entity {
	return &Entity{
		ioMap:       maps.NewSafe(),
		ConnectFunc: DefaultConnectFunc,
		buff:        make(chan byte, 2<<10),
	}
}

type Entity struct {
	key         string                                               //唯已标识
	ioMap       *maps.Safe                                           //存储连接
	ConnectFunc func(msg *Message) (i io.ReadWriteCloser, err error) //连接函数
	buff        chan byte                                            //
	mu          sync.Mutex                                           //
}

// SetKey 设置唯一标识
func (this *Entity) SetKey(key string) *Entity {
	this.key = key
	return this
}

// GetKey 获取唯一标识
func (this *Entity) GetKey() string {
	return this.key
}

func (this *Entity) Proxy(msg *Message) {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, v := range msg.Bytes() {
		this.buff <- v
	}
}

// Read 实现io.Reader
func (this *Entity) Read(p []byte) (n int, err error) {
	for n = 0; ; n++ {
		if len(p) <= n {
			return
		}
		select {
		case b := <-this.buff:
			p[n] = b
		case <-time.After(time.Millisecond):
			return
		}
	}
}

// Write 实现io.Writer 写入数据,解析数据,处理数据
func (this *Entity) Write(p []byte) (int, error) {
	msg, err := DecodeMessage(p)
	if err != nil {
		return 0, err
	}
	return len(p), this.Switch(msg)
}

// Close 实现io.Closer
func (this *Entity) Close() error {
	this.CloseConnAll()
	return nil
}

// SetIO 添加记录,存在则关闭并覆盖
func (this *Entity) SetIO(key string, i io.ReadWriteCloser) {
	old := this.ioMap.GetAndSet(key, i)
	if val, ok := old.(io.Closer); ok {
		val.Close()
	}
}

// GetOrSet 获取或者设置,尝试获取数据,不存在则设置
func (this *Entity) GetOrSet(key string, i io.ReadWriteCloser) io.ReadWriteCloser {
	old := this.GetIO(key)
	if old != nil {
		return old
	}
	this.ioMap.Set(key, i)
	return nil
}

// GetIO 获取io,不存在或者类型错误则返回nil
func (this *Entity) GetIO(key string) io.ReadWriteCloser {
	i, _ := this.ioMap.Get(key)
	if i == nil {
		return nil
	}
	//类型判断是否是需要的类型,是则返回
	if val, ok := i.(io.ReadWriteCloser); ok {
		return val
	}
	//如果记录存在,当类型错误,则删除记录
	this.DelIO(key)
	return nil
}

// DelIO 删除记录
func (this *Entity) DelIO(key string) {
	this.ioMap.Del(key)
}

// CloseIO 关闭io,删除记录据
func (this *Entity) CloseIO(key string) {
	i := this.GetIO(key)
	if i != nil {
		i.Close()
	}
	this.DelIO(key)
}

// CloseConnAll 关闭全部io
func (this *Entity) CloseConnAll() {
	this.ioMap.Range(func(key, value interface{}) bool {
		if val, ok := value.(io.Closer); ok {
			val.Close()
		}
		return true
	})
	this.ioMap = maps.NewSafe()
}

// Switch 处理获取到的消息
func (this *Entity) Switch(msg *Message) (err error) {

	i := this.GetIO(msg.Key)

	if i == nil && (msg.OperateType == Connect || msg.OperateType == Write) {
		if this.ConnectFunc == nil {
			this.ConnectFunc = DefaultConnectFunc
		}
		i, err = this.ConnectFunc(msg)
		if err != nil {
			return err
		}

		c := io.NewClient(i)
		c.SetReadFunc(buf.ReadWithAll)
		c.SetDealFunc(func(msg2 *io.ClientMessage) {
			this.Proxy(NewWriteMessage(msg.Key, msg.Addr, msg2.Bytes()))
		})
		c.SetCloseFunc(func(msg2 *io.ClientMessage) {
			this.DelIO(msg.Key)
			this.Proxy(NewCloseMessage(msg.Key, msg2.String()))
		})
		go c.Run()

		this.SetIO(msg.Key, c)
	}

	if i == nil {
		return
	}

	switch msg.OperateType {
	case Connect:
		//收到建立连接信息
	case Write:
		//收到写数据信息
		_, err = i.Write(msg.Data)
	case Close:
		//收到关闭连接信息
		err = i.Close()
	}

	return
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

func SwapTCPClient(addr string, fn ...func(ctx context.Context, c *io.Client, e *Entity)) error {
	e := New()
	c := io.Redial(dial.TCPFunc(addr), func(ctx context.Context, c *io.Client) {
		c.SetPrefix("P|C")
		c.SetWriteFunc(DefaultWriteFunc)
		for _, v := range fn {
			v(ctx, c, e)
		}
		c.Swap(e)
	})
	go c.Run()
	return nil
}

func SwapTCPServer(port int, fn ...func(s *io.Server)) error {
	s, err := io.NewServer(dial.TCPListenFunc(port))
	if err != nil {
		return err
	}
	s.SetPrefix("P|S")
	s.SetReadFunc(DefaultReadFunc)
	s.Debug()
	for _, v := range fn {
		v(s)
	}
	s.Swap(New())
	go s.Run()
	return nil
}
