package io

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"sync"
	"time"
)

func NewClientManage(ctx context.Context, key string) *ClientManage {
	e := &ClientManage{
		Key:             &Key{key},
		Logger:          NewLog(),
		m:               make(map[string]*Client),
		mu:              sync.RWMutex{},
		ctx:             ctx,
		dealQueue:       chans.NewEntityWithContext(ctx, 1, 1000),
		readChan:        make(chan *IMessage, 100),
		readFunc:        buf.ReadWithAll,
		writeFunc:       nil,
		timeout:         DefaultKeepAlive * 3,
		timeoutInterval: DefaultTimeoutInterval,
	}
	e.dealQueue.SetHandler(func(ctx context.Context, no, count int, data interface{}) {
		if e.dealFunc != nil {
			e.dealFunc(data.(*IMessage))
		}
	})
	//设置默认处理数据函数
	e.SetDealFunc(func(msg *IMessage) { e.readChan <- msg })
	//执行超时机制
	go func() {
		for {
			interval := conv.SelectDuration(e.timeoutInterval > time.Second, e.timeoutInterval, time.Second)
			<-time.After(interval)
			select {
			case <-e.ctx.Done():
			default:
				now := time.Now()
				for _, v := range e.GetClientMap() {
					if e.timeout > 0 && now.Sub(v.IReadCloser.LastTime()) > e.timeout {
						_ = v.CloseWithErr(ErrWithReadTimeout)
					}
				}
			}
		}
	}()

	return e
}

/*
ClientManage
客户端统一管理
例如串口,需要统一
*/
type ClientManage struct {
	*Key
	Logger
	m          map[string]*Client
	mu         sync.RWMutex
	max        int                 //最大数
	ctx        context.Context     //ctx
	beforeFunc func(*Client) error //连接前置事件
	closeFunc  func(msg *IMessage) //断开连接事件
	dealFunc   func(msg *IMessage) //数据处理方法
	dealQueue  *chans.Entity       //数据处理队列
	readChan   chan *IMessage      //数据通道,dealFunc二选一

	readFunc  buf.ReadFunc //数据读取方法
	writeFunc WriteFunc    //数据发送函数,包装下原始数据

	timeout         time.Duration //超时时间,小于0是不超时
	timeoutInterval time.Duration //超时检测间隔
}

// SetReadFunc 设置数据读取
func (this *ClientManage) SetReadFunc(fn func(buf *bufio.Reader) (bytes []byte, err error)) {
	this.readFunc = fn
}

// SetReadWithKB 设置读取固定字节长度
func (this *ClientManage) SetReadWithKB(n uint) {
	this.SetReadFunc(buf.NewReadWithKB(n))
}

// SetReadWithMB 设置读取固定字节长度
func (this *ClientManage) SetReadWithMB(n uint) {
	this.SetReadFunc(buf.NewReadWithMB(n))
}

// SetReadWithAll 设置客户端读取函数:读取全部
func (this *ClientManage) SetReadWithAll() {
	this.SetReadFunc(buf.ReadWithAll)
}

// SetWriteFunc 设置客户端的数据发送函数
func (this *ClientManage) SetWriteFunc(fn func([]byte) ([]byte, error)) {
	this.writeFunc = fn
}

// SetReadWriteWithStartEnd 设置读取方式为起始结束帧
func (this *ClientManage) SetReadWriteWithStartEnd(start, end []byte) {
	this.SetReadFunc(buf.NewReadWithStartEnd(start, end))
	this.SetWriteFunc(buf.NewWriteWithStartEnd(start, end))
}

// SetReadWriteWithPkg 设置读写为默认方式
func (this *ClientManage) SetReadWriteWithPkg() {
	this.SetReadFunc(ReadWithPkg)
	this.SetWriteFunc(WriteWithPkg)
}

// SetDealQueueNum 设置数据处理队列协程数量,并发处理
// 例如处理方式是进行数据转发(或处理较慢)时,会出现阻塞,后续数据等待过长
func (this *ClientManage) SetDealQueueNum(n int) {
	this.dealQueue.SetNum(n)
}

// SetDealFunc 设置处理数据方法
func (this *ClientManage) SetDealFunc(fn func(msg *IMessage)) {
	this.dealFunc = fn
	for {
		//清除阻塞数据
		select {
		case <-this.readChan:
			continue
		default:
		}
		break
	}
}

// SetDealWithWriter 读取到的数据写入到writer
func (this *ClientManage) SetDealWithWriter(w Writer) {
	this.SetDealFunc(func(msg *IMessage) {
		w.Write(msg.Bytes())
	})
}

// SetBeforeFunc 设置连接前置方法
// 如果返回错误则关闭该连接,需要主动读取数据
func (this *ClientManage) SetBeforeFunc(fn func(c *Client) error) {
	this.beforeFunc = fn
}

// SetCloseFunc 设置断开连接事件
func (this *ClientManage) SetCloseFunc(fn func(msg *IMessage)) {
	this.closeFunc = fn
}

// SetMaxClient 设置最大连接数,超过最大连接数的连接会直接断开
func (this *ClientManage) SetMaxClient(max int) {
	this.max = max
}

// SetTimeout 设置超时时间,还有time/3的时间误差
func (this *ClientManage) SetTimeout(t time.Duration) {
	this.timeout = t
}

// SetTimeoutInterval 设置超时检测间隔,至少需要1秒
func (this *ClientManage) SetTimeoutInterval(ti time.Duration) {
	this.timeoutInterval = conv.SelectDuration(ti > time.Second, ti, time.Second)
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
	if this.max > 0 && this.GetClientLen() >= this.max {
		c.WriteString(fmt.Sprintf("超过最大连接数(%d)", this.GetClientLen()))
		c.CloseAll()
		return
	}

	c.SetDealFunc(this._dealFunc)  //数据处理方法
	c.SetReadFunc(this.readFunc)   //读取数据方法
	c.SetWriteFunc(this.writeFunc) //设置发送函数
	c.SetLogger(this.Logger)       //同步logger配置

	// 协程执行,等待连接的后续数据,来决定后续操作
	go func(c *Client) {

		//前置操作,例如等待注册数据,不符合的返回错误则关闭连接
		if this.beforeFunc != nil {
			if err := this.beforeFunc(c); err != nil {
				this.Logger.Errorf("[%s] %v", c.GetKey(), err)
				_ = c.Close()
				return
			}
		}

		this.mu.RLock()
		old, ok := this.m[c.GetKey()]
		this.mu.RUnlock()
		if ok && old.Pointer() != c.Pointer() {
			old.CloseAll()
		}

		//设置连接关闭事件
		c.SetCloseFunc(this._closeFunc(c.GetCloseFunc()))

		//加入map 进行管理
		this.mu.Lock()
		this.m[c.GetKey()] = c
		this.mu.Unlock()
		c.Run()

	}(c)

}

// GetClient 获取客户端
func (this *ClientManage) GetClient(key string) *Client {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.m[key]
}

// GetClientAny 获取任意一个客户端
func (this *ClientManage) GetClientAny() *Client {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for _, v := range this.m {
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
	return len(this.m)
}

// GetClientMap 获取客户端map,元数据,注意安全
func (this *ClientManage) GetClientMap() map[string]*Client {
	return this.m
}

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
	for i, v := range this.m {
		if !fn(i, v) {
			break
		}
	}
}

// Read 无效,使用ReadMessage
func (this *ClientManage) Read(p []byte) (int, error) {
	return 0, nil
}

// ReadMessage 读取数据,当未设置DealFunc时生效
func (this *ClientManage) ReadMessage() ([]byte, error) {
	m := <-this.readChan
	return m.Message, nil
}

// WriteClient 写入客户端数据
func (this *ClientManage) WriteClient(key string, p []byte) (bool, error) {
	return this.GetClientDo(key, func(c *Client) error {
		_, err := c.Write(p)
		return err
	})
}

// WriteClientAll 广播,发送数据给所有连接,加入到连接的队列
func (this *ClientManage) WriteClientAll(p []byte) {
	for _, c := range this.GetClientMap() {
		//写入到队列,避免阻塞
		c.WriteQueue(p)
	}
}

// TryWriteClientAll 广播,发送数据给所有连接,尝试加入到连接的队列
func (this *ClientManage) TryWriteClientAll(p []byte) {
	for _, c := range this.GetClientMap() {
		//写入到队列,避免阻塞,加入不了则丢弃数据
		c.TryWriteQueue(p)
	}
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
		return c.CloseAll()
	}
	return nil
}

// CloseClientAll 关闭所有客户端
func (this *ClientManage) CloseClientAll() {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, v := range this.m {
		v.CloseAll()
	}
	this.m = make(map[string]*Client)
}

// SetClientKey 重命名key
func (this *ClientManage) SetClientKey(newClient *Client, newKey string) {
	//判断这个标识符的客户端是否存在,存在则关闭
	if oldClient := this.GetClient(newKey); oldClient != nil {
		//判断指针地址是否一致,不一致则关闭
		if oldClient.Pointer() != newClient.Pointer() {
			oldClient.CloseAll()
		}
	}
	//更新新的客户端
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.m, newClient.GetKey())
	this.m[newKey] = newClient.SetKey(newKey)
}

/*



 */

// _dealFunc 处理数据
func (this *ClientManage) _dealFunc(msg *IMessage) {
	select {
	case <-this.ctx.Done():
	default:
		//加入消费队列
		this.dealQueue.Do(msg)
	}
}

func (this *ClientManage) _closeFunc(closeFunc ...CloseFunc) func(ctx context.Context, msg *IMessage) {
	return func(ctx context.Context, msg *IMessage) {
		defer func() {
			for _, f := range closeFunc {
				if f != nil {
					f(ctx, msg.Message)
				}
			}
		}()
		if this.closeFunc != nil {
			defer this.closeFunc(msg)
		}
		this.mu.Lock()
		defer this.mu.Unlock()
		//获取老的连接
		oldConn := this.m[msg.GetKey()]
		//存在新连接上来被关闭的情况,判断是否是老的连接
		if oldConn == nil || oldConn.Pointer() != msg.Pointer() {
			return
		}
		delete(this.m, msg.GetKey())
	}
}
