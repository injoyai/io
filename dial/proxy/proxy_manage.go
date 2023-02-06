package proxy

import (
	"encoding/json"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"sync"
)

// Manage 连接任意客户端
type Manage struct {
	m  map[string]*io.Client
	mu sync.RWMutex
}

func NewProxy() *Proxy {
	return &Proxy{
		conn: maps.NewSafe(),
		//redirect: newRedirect(),
	}
}

type Proxy struct {
	key  string     //唯已标识
	conn *maps.Safe //存储连接
	//connectFunc func(key, addr string) (io.ReadWriteCloser, error) //代理连接建立函数
	//redirect    IRedirect                                          //重定向接口
}

// SetKey 设置唯一标识
func (this *Proxy) SetKey(key string) *Proxy {
	this.key = key
	return this
}

// GetKey 获取唯一标识
func (this *Proxy) GetKey() string {
	return this.key
}

// SetConn 添加记录,存在则关闭并覆盖
func (this *Proxy) SetConn(key string, i io.ReadWriteCloser) {
	old := this.conn.GetAndSet(key, i)
	if val, ok := old.(io.Closer); ok {
		val.Close()
	}
}

// GetOrSet 获取或者设置,尝试获取数据,不存在则设置
func (this *Proxy) GetOrSet(key string, i io.ReadWriteCloser) io.ReadWriteCloser {
	old := this.GetConn(key)
	if old != nil {
		return old
	}
	this.conn.Set(key, i)
	return nil
}

// GetConn 获取io,不存在或者类型错误则返回nil
func (this *Proxy) GetConn(key string) io.ReadWriteCloser {
	i, _ := this.conn.Get(key)
	if i == nil {
		return nil
	}
	//类型判断是否是需要的类型,是则返回
	if val, ok := i.(io.ReadWriteCloser); ok {
		return val
	}
	//如果记录存在,当类型错误,则删除记录
	this.DelConn(key)
	return nil
}

// DelConn 删除记录
func (this *Proxy) DelConn(key string) {
	this.conn.Del(key)
}

// CloseConn 关闭io,删除记录据
func (this *Proxy) CloseConn(key string) {
	i := this.GetConn(key)
	if i != nil {
		i.Close()
	}
	this.DelConn(key)
}

// CloseConnAll 关闭全部io
func (this *Proxy) CloseConnAll() {
	this.conn.Range(func(key, value interface{}) bool {
		if val, ok := value.(io.Closer); ok {
			val.Close()
		}
		return true
	})
	this.conn = maps.NewSafe()
}

// Switch 处理获取到的消息
func (this *Proxy) Switch(msg *Message) (err error) {

	// 获取连接
	i := this.GetConn(msg.Key)

	if i == nil && (msg.OperateType == Connect || msg.OperateType == Write) {
		//如果连接不存在,则新建连接,并存储
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
		}
		this.SetConn(msg.Key, i)
	}

	if i == nil {
		return
	}

	switch msg.OperateType {
	case Connect:
		//收到建立连接信息

	case Write:
		//收到写数据信息
		_, err = i.Write([]byte(msg.Data))
	case Close:
		//收到关闭连接信息
		err = i.Close()
	}

	return
}
