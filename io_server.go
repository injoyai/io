package io

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"sync"
	"sync/atomic"
	"time"
)

func NewServer(newListen ListenFunc, fn ...func(s *Server)) (*Server, error) {
	return NewServerWithContext(context.Background(), newListen, fn...)
}

func NewServerWithContext(ctx context.Context, newListen func() (Listener, error), fn ...func(s *Server)) (*Server, error) {
	listener, err := newListen()
	if err != nil {
		return nil, err
	}
	s := &Server{
		IPrinter:   NewIPrinter(fmt.Sprintf("%p", listener)),
		listener:   listener,
		clientMap:  make(map[string]*Client),
		timeout:    DefaultTimeout * 3,
		readFunc:   buf.ReadWithAll,
		dealFunc:   nil,
		dealQueue:  chans.NewEntity(1, 1000),
		closeFunc:  nil,
		beforeFunc: nil,
		writeFunc:  nil,
		printFunc:  PrintWithASCII,
	}
	s.SetBeforeFunc(s._beforeFunc)
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.dealQueue.SetHandler(func(no, num int, data interface{}) {
		if s.dealFunc != nil {
			s.dealFunc(data.(*IMessage))
		}
	})
	go s.timeoutFunc()
	for _, v := range fn {
		v(s)
	}
	return s, nil
}

// Server todo 整体待优化
type Server struct {
	*IPrinter
	listener   Listener
	clientMap  map[string]*Client  //链接集合,远程地址为key
	clientMu   sync.RWMutex        //锁
	ctx        context.Context     //上下文
	cancel     context.CancelFunc  //上下文
	beforeFunc func(*Client) error //连接前置事件
	dealFunc   func(msg *IMessage) //数据处理方法
	dealQueue  *chans.Entity       //数据处理队列
	closed     uint32              //是否关闭
	closeErr   error               //错误信息

	readFunc  buf.ReadFunc          //数据读取方法
	closeFunc func(msg *IMessage)   //断开连接事件
	writeFunc func(p []byte) []byte //数据发送函数,包装下原始数据
	printFunc PrintFunc             //打印数据方法
	running   uint32                //是否在运行
	timeout   time.Duration         //超时时间,0是永久有效
}

// Ctx 上下文
func (this *Server) Ctx() context.Context {
	return this.ctx
}

// Done ctx.Done
func (this *Server) Done() <-chan struct{} {
	return this.Ctx().Done()
}

// Close 关闭
func (this *Server) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭
func (this *Server) CloseWithErr(err error) error {
	select {
	case <-this.Done():
	default:
		if err != nil {
			this.closeErr = err
			if this.cancel != nil {
				this.cancel()
			}
			this.listener.Close()
			this.CloseClientAll()
		}
	}
	return nil
}

// SetDealQueueNum 设置数据处理队列协程数量
func (this *Server) SetDealQueueNum(num int) *Server {
	this.dealQueue.SetNum(num)
	return this
}

// SetBeforeFunc 设置连接前置方法,连接数据还未监听
func (this *Server) SetBeforeFunc(fn func(c *Client) error) *Server {
	this.beforeFunc = fn
	return this
}

// SetCloseFunc 设置断开连接事件
func (this *Server) SetCloseFunc(fn func(msg *IMessage)) *Server {
	this.closeFunc = fn
	return this
}

// SetDealFunc 设置处理数据方法
func (this *Server) SetDealFunc(fn func(msg *IMessage)) *Server {
	this.dealFunc = fn
	return this
}

// SetDealWithWriter 读取到的数据写入到writer
func (this *Server) SetDealWithWriter(writer Writer) *Server {
	this.SetDealFunc(func(msg *IMessage) {
		writer.Write(msg.Bytes())
	})
	return this
}

// SetReadFunc 设置数据读取
func (this *Server) SetReadFunc(fn func(buf *bufio.Reader) (bytes []byte, err error)) *Server {
	this.readFunc = fn
	return this
}

// SetReadWithAll 设置读取函数:读取全部
func (this *Server) SetReadWithAll() {
	this.SetReadFunc(buf.ReadWithAll)
}

// SetWriteFunc 设置数据发送函数
func (this *Server) SetWriteFunc(fn func([]byte) []byte) *Server {
	this.writeFunc = fn
	return this
}

// SetPrintFunc 设置打印方式
func (this *Server) SetPrintFunc(fn PrintFunc) *Server {
	this.IPrinter.SetPrintFunc(fn)
	this.printFunc = fn
	return this
}

// SetPrintWithHEX 设置打印方式HEX
func (this *Server) SetPrintWithHEX() *Server {
	this.printFunc = PrintWithHEX
	return this
}

// SetPrintWithASCII 设置打印方式ASCII
func (this *Server) SetPrintWithASCII() *Server {
	this.printFunc = PrintWithASCII
	return this
}

// SetTimeout 设置超时时间,还有time/3的时间误差
func (this *Server) SetTimeout(t time.Duration) *Server {
	this.timeout = t
	return this
}

// GetClient 获取一个客户端
func (this *Server) GetClient(key string) *Client {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	return this.clientMap[key]
}

// GetClientAny 获取任意一个连接
func (this *Server) GetClientAny() *Client {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	for _, v := range this.clientMap {
		return v
	}
	return nil
}

// GetClientMap 获取所有连接
func (this *Server) GetClientMap() map[string]*Client {
	m := make(map[string]*Client)
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	for i, v := range this.clientMap {
		m[i] = v
	}
	return m
}

// GetClientCount 获取客户端数量
func (this *Server) GetClientCount() int {
	return len(this.clientMap)
}

// Read todo
func (this *Server) Read(p []byte) (int, error) {
	return 0, nil
}

// WriteClient 给一个客户端发送数据
func (this *Server) WriteClient(key string, msg []byte) (exist bool, err error) {
	c := this.GetClient(key)
	if exist = c != nil; exist {
		_, err = c.Write(msg)
	}
	return
}

// Write 给所有客户端发送数据,实现io.Writer接口
func (this *Server) Write(p []byte) (int, error) {
	this.WriteClientAll(p)
	return len(p), nil
}

// WriteClientAll 广播,发送数据给所有连接
func (this *Server) WriteClientAll(msg []byte) {
	for _, c := range this.GetClientMap() {
		c.Write(msg)
	}
}

// CloseClient 关闭一个连接
func (this *Server) CloseClient(key string) {
	if c := this.GetClient(key); c != nil {
		c.Close()
	}
}

// CloseClientAll 关闭所有连接
func (this *Server) CloseClientAll() {
	for _, c := range this.GetClientMap() {
		c.Close()
	}
}

// SetClientKey 重命名key
func (this *Server) SetClientKey(newClient *Client, newKey string) {
	//判断这个标识符的客户端是否存在,存在则关闭
	if oldClient := this.GetClient(newKey); oldClient != nil {
		//判断指针地址是否一致,不一致则关闭
		if oldClient.Pointer() != newClient.Pointer() {
			oldClient.Close()
		}
	}
	//更新新的客户端
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	delete(this.clientMap, newClient.GetKey())
	this.clientMap[newKey] = newClient.SetKey(newKey)
}

// GoFor 协程循环
func (this *Server) GoFor(interval time.Duration, do func(s *Server)) {
	if interval > 0 {
		go func() {
			timer := time.NewTimer(interval)
			defer timer.Stop()
			for {
				timer.Reset(interval)
				select {
				case <-this.Done():
					return
				case <-timer.C:
					do(this)
				}
			}
		}()
	}
}

// Swap 和一个IO交换数据
func (this *Server) Swap(i ReadWriteCloser) *Server {
	this.SwapWithReadFunc(i, buf.ReadWithAll)
	return this
}

// SwapWithReadFunc 根据读取规则俩进行IO数据交换
func (this *Server) SwapWithReadFunc(i ReadWriteCloser, readFunc buf.ReadFunc) {
	c := NewClient(i)
	c.SetReadFunc(readFunc)
	this.SwapClient(c)
}

// SwapClient 和一个客户端交换数据
func (this *Server) SwapClient(c *Client) {
	this.SetDealWithWriter(c)
	c.SetDealWithWriter(this)
	go c.Run()
}

// SwapServer 和另一个服务交换数据
func (this *Server) SwapServer(s *Server) {
	this.SetDealWithWriter(s)
	s.SetDealWithWriter(this)
}

// Running 是否在运行
func (this *Server) Running() bool {
	return this.running == 1
}

// Run 运行(监听)
func (this *Server) Run() error {

	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}

	this.Print(NewMessage("开启服务成功..."), TagInfo, this.GetKey())

	for {
		select {
		case <-this.ctx.Done():
			return this.closeErr
		default:
		}

		c, key, err := this.listener.Accept()
		if err != nil {
			this.CloseWithErr(err)
			return this.closeErr
		}

		//新建客户端,并配置
		x := NewClientWithContext(this.ctx, c)
		x.SetKey(key)                   //设置唯一标识符
		x.Debug(this.GetDebug())        //调试模式
		x.SetReadFunc(this.readFunc)    //读取数据方法
		x.SetDealFunc(this._dealFunc)   //数据处理方法
		x.SetCloseFunc(this._closeFunc) //连接关闭方法
		x.SetTimeout(0)                 //设置超时时间
		x.SetPrintFunc(this.printFunc)  //设置打印函数
		x.SetWriteFunc(this.writeFunc)  //设置发送函数

		// 协程执行,等待连接的后续数据,来决定后续操作
		go func(x *Client) {
			//前置操作,例如等待注册数据,不符合的返回错误则关闭连接
			if this.beforeFunc != nil {
				if err := this.beforeFunc(x); err != nil {
					_ = c.Close()
					return
				}
			}
			//加入map 进行管理
			this.clientMu.Lock()
			this.clientMap[x.GetKey()] = x
			this.clientMu.Unlock()
			x.Run()
		}(x)

	}
}

/*

inside

*/

// beforeFunc 默认前置函数
func (this *Server) _beforeFunc(c *Client) error {
	this.Print(NewMessage("新的客户端连接..."), TagInfo, c.GetKey())
	return nil
}

// delConn 删除连接
func (this *Server) _closeFunc(ctx context.Context, msg *IMessage) {
	if this.closeFunc != nil {
		defer this.closeFunc(msg)
	}
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	oldConn := this.clientMap[msg.GetKey()]
	if oldConn == nil || oldConn.Pointer() != msg.Pointer() {
		//存在新连接上来被关闭的情况,判断是否是老的连接
		return
	}
	delete(this.clientMap, msg.GetKey())
}

// _dealFunc 处理数据
func (this *Server) _dealFunc(msg *IMessage) {
	select {
	case <-this.ctx.Done():
	default:
		this.dealQueue.Do(msg)
	}
}

// timeoutFunc 服务端超时机制,(客户端突然断电,服务端检测不出来)
func (this *Server) timeoutFunc() {
	for {
		interval := conv.SelectDuration(this.timeout/3 > time.Second, this.timeout/3, time.Minute)
		<-time.After(interval)
		now := time.Now()
		for _, v := range this.GetClientMap() {
			if this.timeout > 0 && now.Sub(v.IReadCloser.LastTime()) > this.timeout {
				_ = v.CloseWithErr(ErrWithReadTimeout)
			}
		}
	}
}
