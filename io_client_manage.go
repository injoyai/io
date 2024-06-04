package io

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/maps/timeout"
	"github.com/injoyai/conv"
	"sync"
	"time"
)

func NewClientManage(key string, log *logger) *ClientManage {
	e := &ClientManage{
		Key:    Key(key),
		Logger: log,
		mKey:   make(map[string]*Client),
		mu:     sync.RWMutex{},
		Keep:   timeout.New(),
		options: []OptionClient{func(c *Client) {
			c.SetConnectFunc(func(c *Client) error {
				log.Infof("[%s] 新的客户端连接...\n", key)
				return nil
			})
			c.SetReadWith1KB()
			c.SetCloseWithLog()
		}},
	}

	//超时机制
	e.SetTimeout(DefaultTimeout)
	e.SetTimeoutInterval(DefaultTimeoutInterval)
	e.Keep.SetDealFunc(func(key interface{}) error {
		return e.CloseClientWithErr(conv.String(key), ErrWithReadTimeout)
	})

	return e
}

/*
ClientManage
客户端统一管理
例如串口,需要统一
*/
type ClientManage struct {
	Key
	Logger       *logger          //日志
	Keep         *timeout.Timeout //超时机制
	mID          sync.Map
	mKey         map[string]*Client
	mu           sync.RWMutex
	maxClientNum int            //限制最大客户端数
	options      []OptionClient //客户端Option
}

func (this *ClientManage) SetOptions(option ...Option) {
	this.options = append(this.options, option...)
}

func (this *ClientManage) SetConnectFunc(f func(c *Client) error) {
	this.SetOptions(func(c *Client) { c.SetConnectFunc(f) })
}

func (this *ClientManage) SetReadFunc(f func(buf *bufio.Reader) ([]byte, error)) {
	this.SetOptions(func(c *Client) { c.SetReadFunc(f) })
}

func (this *ClientManage) SetDealFunc(f func(c *Client, msg Message)) {
	this.SetOptions(func(c *Client) { c.SetDealFunc(f) })
}

func (this *ClientManage) SetWriteFunc(f func(p []byte) ([]byte, error)) {
	this.SetOptions(func(c *Client) { c.SetWriteFunc(f) })
}

func (this *ClientManage) OnMessage(f func(c *Client, msg Message)) {
	this.SetOptions(func(c *Client) { c.SetDealFunc(f) })
}

// SetMaxClient 设置最大连接数,超过最大连接数的连接会直接断开
func (this *ClientManage) SetMaxClient(max int) {
	this.maxClientNum = max
}

// SetTimeout 设置超时时间,还有time/3的时间误差
func (this *ClientManage) SetTimeout(t time.Duration) {
	this.Keep.SetTimeout(t)
}

// SetTimeoutInterval 设置超时检测间隔,至少需要1秒
func (this *ClientManage) SetTimeoutInterval(ti time.Duration) {
	this.Keep.SetInterval(conv.SelectDuration(ti > time.Second, ti, time.Second))
}

// Close 关闭,实现io.Closer接口
func (this *ClientManage) Close() error {
	this.CloseClientAll()
	return nil
}

// DialClient 连接客户端
func (this *ClientManage) DialClient(dialFunc DialFunc) (*Client, error) {
	c, err := NewDial(dialFunc)
	if err != nil {
		return nil, err
	}
	this.SetClient(c)
	return c, nil
}

// RedialClient 连接客户端直到成功
func (this *ClientManage) RedialClient(dial DialFunc, options ...OptionClient) *Client {
	c := Redial(dial, options...)
	this.SetClient(c)
	return c
}

// SetClient 添加客户端
func (this *ClientManage) SetClient(c *Client) {
	if c == nil {
		return
	}

	//判断是否到达最大连接数,禁止新连接
	if this.maxClientNum > 0 && this.GetClientLen() >= this.maxClientNum {
		c.WriteString(fmt.Sprintf("超过最大连接数(%d)", this.GetClientLen()))
		c.CloseAll()
		return
	}

	//this.SetClientOptions(c)
	c.SetOptions(this.options...)
	c.SetCloseFunc(this.closeFunc)

	// 协程执行,等待连接的后续数据,来决定后续操作
	go func(c *Client) {

		////前置操作,例如等待注册数据,不符合的返回错误则关闭连接
		//if this.beforeFunc != nil {
		//	if err := this.beforeFunc(c); err != nil {
		//		this.Logger.Errorf("[%s] %v\n", c.GetKey(), err)
		//		//丢弃连接,防止重复连接断开
		//		//_ = c.CloseAll()
		//		return
		//	}
		//}

		//TODO 怎么判断是客户端还是服务端的连接实例
		//前置操作,例如等待注册数据,不符合的返回错误则关闭连接
		for _, f := range c.connectFunc {
			if err := f(c); err != nil {
				this.Logger.Errorf("[%s] %v\n", c.GetKey(), err)
				//丢弃连接,防止重复连接断开
				_ = c.CloseAll()
				return
			}
		}

		//注册成功,验证通过
		//判断是否存在老连接,存在则关闭老连接(被挤下线)
		this.mu.RLock()
		old, ok := this.mKey[c.GetKey()]
		this.mu.RUnlock()
		if ok && old.Pointer() != c.Pointer() {
			old.CloseAll()
		}

		//加入map 进行管理
		this.mu.Lock()
		this.mKey[c.GetKey()] = c
		this.mu.Unlock()
		this.mID.Store(c.Pointer(), c)
		c.Run()

	}(c)

}

func (this *ClientManage) GetClientByID(id string) *Client {
	v, ok := this.mID.Load(id)
	if ok {
		return v.(*Client)
	}
	return nil
}

// GetClient 获取客户端
func (this *ClientManage) GetClient(key string) *Client {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.mKey[key]
}

// GetClientOrDial 获取客户端或者(不存在)重新链接
func (this *ClientManage) GetClientOrDial(key string, dialFunc DialFunc) (*Client, error) {
	this.mu.RLock()
	c, ok := this.mKey[key]
	this.mu.RUnlock()
	if !ok {
		var err error
		c, err = this.DialClient(dialFunc)
		if err != nil {
			return nil, err
		}
		c.SetKey(key)
		this.SetClient(c)
	}
	return c, nil
}

// GetClientAny 获取任意一个客户端
func (this *ClientManage) GetClientAny() *Client {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for _, v := range this.mKey {
		return v
	}
	return nil
}

// GetClientDo 获取客户端并执行
func (this *ClientManage) GetClientDo(key string, fn func(c *Client) error) (bool, error) {
	c := this.GetClient(key)
	if c != nil {
		return true, fn(c)
	}
	return false, nil
}

// GetClientLen 获取客户端数量
func (this *ClientManage) GetClientLen() int {
	return len(this.mKey)
}

// Publish 实现接口,io.Publisher
func (this *ClientManage) Publish(topic string, data interface{}) error {
	c := this.GetClient(topic)
	if c != nil {
		_, err := c.WriteAny(data)
		return err
	}
	return errors.New("client not found")
}

// GetClientMap 获取客户端map,元数据,注意安全
//func (this *ClientManage) GetClientMap() map[string]*Client {
//	return this.m
//}

// CopyClientMap 复制所有客户端数据
func (this *ClientManage) CopyClientMap() map[string]*Client {
	m := make(map[string]*Client)
	this.RangeClient(func(key string, c *Client) bool {
		m[key] = c
		return true
	})
	return m
}

// RangeClient 遍历客户端
func (this *ClientManage) RangeClient(fn func(key string, c *Client) bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for i, v := range this.mKey {
		if !fn(i, v) {
			break
		}
	}
}

// Read 无效,使用ReadMessage
func (this *ClientManage) Read(p []byte) (int, error) {
	return 0, ErrUseReadMessage
}

//// ReadMessage 读取数据,当未设置DealFunc时生效
//func (this *ClientManage) ReadMessage() ([]byte, error) {
//	msg := <-this.readChan
//	return msg, nil
//}

// WriteClient 写入客户端数据
func (this *ClientManage) WriteClient(key string, p []byte) (bool, error) {
	return this.GetClientDo(key, func(c *Client) error {
		_, err := c.Write(p)
		return err
	})
}

// WriteClientAll 广播,发送数据给所有连接,加入到连接的队列
func (this *ClientManage) WriteClientAll(p []byte) {
	this.RangeClient(func(key string, c *Client) bool {
		//写入到队列,避免阻塞
		c.WriteQueue(p)
		return true
	})
}

// TryWriteClientAll 广播,发送数据给所有连接,尝试加入到连接的队列
func (this *ClientManage) TryWriteClientAll(p []byte) {
	this.RangeClient(func(key string, c *Client) bool {
		//写入到队列,避免阻塞,加入不了则丢弃数据
		c.WriteQueueTry(p)
		return true
	})
}

// WriteClientAny 写入任意一个客户端数据
func (this *ClientManage) WriteClientAny(p []byte) (int, error) {
	if c := this.GetClientAny(); c != nil {
		return c.Write(p)
	}
	return len(p), nil
}

// WriteReadClient 写入客户端并等待结果返回
func (this *ClientManage) WriteReadClient(key string, p []byte) ([]byte, bool, error) {
	var res []byte
	var err error
	has, err := this.GetClientDo(key, func(c *Client) error {
		res, err = c.WriteRead(p)
		return err
	})
	return res, has, err
}

// Write 给所有客户端发送数据,实现io.Writer接口
func (this *ClientManage) Write(p []byte) (int, error) {
	this.WriteClientAll(p)
	return len(p), nil
}

// CloseClient 关闭客户端,会重试
func (this *ClientManage) CloseClient(key string) error {
	if c := this.GetClient(key); c != nil {
		if err := c.CloseAll(); err != nil {
			return err
		}
		this.mID.Delete(c.Pointer())
	}
	return nil
}

// CloseClientWithErr 关闭客户端,会重试
func (this *ClientManage) CloseClientWithErr(key string, err error) error {
	if c := this.GetClient(key); c != nil && err != nil {
		if err := c.CloseAllWithErr(err); err != nil {
			return err
		}
		this.mID.Delete(c.Pointer())
	}
	return nil
}

// CloseClientAll 关闭所有客户端
func (this *ClientManage) CloseClientAll() {
	for _, v := range this.CopyClientMap() {
		v.CloseAll()
	}
	this.mKey = make(map[string]*Client)
}

// SetClientKey 重命名key
func (this *ClientManage) SetClientKey(newClient *Client, newKey string) {
	//判断这个标识符的客户端是否存在,存在则关闭
	if oldClient := this.GetClient(newKey); oldClient != nil {
		//判断指针地址是否一致,不一致则关闭
		if oldClient.Pointer() != newClient.Pointer() {
			this.mID.Delete(oldClient.Pointer())
			oldClient.CloseAll()
		}
	}
	this.mID.Store(newClient.Pointer(), newClient)
	//更新新的客户端
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.mKey, newClient.GetKey())
	newClient.SetKey(newKey)
	this.mKey[newKey] = newClient
}

/*



 */

//// _dealFunc 处理数据
//func (this *ClientManage) _dealFunc(c *Client, msg Message) {
//	select {
//	case <-this.ctx.Done():
//	default:
//		//尝试加入通道
//		select {
//		case this.readChan <- msg:
//		default:
//		}
//		if this.dealFunc != nil {
//			this.dealFunc(c, msg)
//		}
//	}
//}

//func (this *ClientManage) _closeFunc(ctx context.Context, c *Client, err error) {
//	defer c.CloseAll()
//	if this.closeFunc != nil {
//		defer this.closeFunc(c, err)
//	}
//	this.mu.Lock()
//	defer this.mu.Unlock()
//	//获取老的连接
//	oldConn := this.m[c.GetKey()]
//	//存在新连接上来被关闭的情况,判断是否是老的连接
//	if oldConn == nil || oldConn.Pointer() != c.Pointer() {
//		return
//	}
//	this.mid.Delete(c.Pointer())
//	delete(this.m, c.GetKey())
//}

func (this *ClientManage) closeFunc(ctx context.Context, c *Client, err error) {
	defer c.CloseAll()
	this.mu.Lock()
	defer this.mu.Unlock()
	//获取老的连接
	oldConn := this.mKey[c.GetKey()]
	//存在新连接上来被关闭的情况,判断是否是老的连接
	if oldConn == nil || oldConn.Pointer() != c.Pointer() {
		return
	}
	this.mID.Delete(c.Pointer())
	delete(this.mKey, c.GetKey())
}
