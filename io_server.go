package io

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/maps"
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
	//连接listener
	listener, err := newListen()
	if err != nil {
		return nil, err
	}
	//新建实例
	s := &Server{
		printer:    newPrinter(fmt.Sprintf("%p", listener)),
		ICloser:    NewICloserWithContext(ctx, listener),
		Tag:        maps.NewSafe(),
		listener:   listener,
		clientMap:  make(map[string]*Client),
		timeout:    DefaultTimeout * 3,
		readFunc:   buf.ReadWithAll,
		dealFunc:   nil,
		dealQueue:  chans.NewEntity(1, 1000),
		readChan:   make(chan *IMessage),
		closeFunc:  nil,
		clientMax:  0,
		beforeFunc: nil,
		writeFunc:  nil,
		printFunc:  PrintWithASCII,
	}
	//设置关闭函数
	s.ICloser.SetCloseFunc(func(ctx context.Context, msg Message) {
		//关闭listener
		s.listener.Close()
		//关闭已连接的客户端,关闭listener后,客户端还能正常通讯
		s.CloseClientAll()
	})
	//设置前置函数
	s.SetBeforeFunc(func(c *Client) error {
		//默认连接打印信息
		s.Print(NewMessage("新的客户端连接..."), TagInfo, c.GetKey())
		return nil
	})
	//设置默认处理数据函数
	s.SetDealFunc(func(msg *IMessage) { s.readChan <- msg })
	//设置队列处理数据函数
	s.dealQueue.SetHandler(func(no, num int, data interface{}) {
		if s.dealFunc != nil {
			s.dealFunc(data.(*IMessage))
		}
	})
	//预设服务处理
	for _, v := range fn {
		v(s)
	}
	return s, nil
}

// Server 服务端
type Server struct {
	*printer
	*ICloser

	Tag        *maps.Safe          //
	listener   Listener            //listener
	clientMap  map[string]*Client  //链接集合,远程地址为key
	clientMu   sync.RWMutex        //锁
	clientMax  int                 //最大连接数
	beforeFunc func(*Client) error //连接前置事件
	dealFunc   func(msg *IMessage) //数据处理方法
	dealQueue  *chans.Entity       //数据处理队列
	readChan   chan *IMessage      //数据通道,dealFunc二选一

	readFunc        buf.ReadFunc        //数据读取方法
	closeFunc       func(msg *IMessage) //断开连接事件
	writeFunc       WriteFunc           //数据发送函数,包装下原始数据
	printFunc       PrintFunc           //打印数据方法
	running         uint32              //是否在运行
	timeout         time.Duration       //超时时间,小于0是不超时
	timeoutInterval time.Duration       //超时检测间隔
}

//================================SetFunc================================

// SetMaxClient 设置最大连接数,超过最大连接数的连接会直接断开
func (this *Server) SetMaxClient(max int) *Server {
	this.clientMax = max
	return this
}

// SetBeforeFunc 设置连接前置方法
// 如果返回错误则关闭该连接,需要主动读取数据
func (this *Server) SetBeforeFunc(fn func(c *Client) error) *Server {
	this.beforeFunc = fn
	return this
}

// SetCloseFunc 设置断开连接事件
func (this *Server) SetCloseFunc(fn func(msg *IMessage)) *Server {
	this.closeFunc = fn
	return this
}

// SetDealQueueNum 设置数据处理队列协程数量,并发处理
// 例如处理方式是进行数据转发(或处理较慢)时,会出现阻塞,后续数据等待过长
func (this *Server) SetDealQueueNum(n int) *Server {
	this.dealQueue.SetNum(n)
	return this
}

// SetDealFunc 设置处理数据方法
func (this *Server) SetDealFunc(fn func(msg *IMessage)) *Server {
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
	return this
}

// SetDealWithWriter 读取到的数据写入到writer
func (this *Server) SetDealWithWriter(writer Writer) *Server {
	return this.SetDealFunc(func(msg *IMessage) {
		writer.Write(msg.Bytes())
	})
}

// SetReadFunc 设置数据读取
func (this *Server) SetReadFunc(fn func(buf *bufio.Reader) (bytes []byte, err error)) *Server {
	this.readFunc = fn
	return this
}

// SetReadWithPkg 读取方式
func (this *Server) SetReadWithPkg() *Server {
	return this.SetReadFunc(ReadWithPkg)
}

// SetReadWithAll 设置读取函数:读取全部
func (this *Server) SetReadWithAll() *Server {
	return this.SetReadFunc(buf.ReadWithAll)
}

// SetWriteFunc 设置数据发送函数
func (this *Server) SetWriteFunc(fn func([]byte) ([]byte, error)) *Server {
	this.writeFunc = fn
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

// SetTimeout 设置超时时间,还有time/3的时间误差
func (this *Server) SetTimeout(t time.Duration) *Server {
	this.timeout = t
	return this
}

// SetTimeoutInterval 设置超时检测间隔,至少需要1秒
func (this *Server) SetTimeoutInterval(ti time.Duration) *Server {
	this.timeoutInterval = conv.SelectDuration(ti > time.Second, ti, time.Second)
	return this
}

//================================Client================================

// GetClient 根据key获取一个客户端
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

// GetClientDo 获取客户端并执行函数,返回是否存在和执行错误,先判断是否存在
func (this *Server) GetClientDo(key string, fn func(c *Client) error) (bool, error) {
	c := this.GetClient(key)
	if c != nil {
		return true, fn(c)
	}
	return false, nil
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

// GetClientLen 获取客户端数量
func (this *Server) GetClientLen() int {
	return len(this.clientMap)
}

// Read 无效,使用ReadMessage
func (this *Server) Read(p []byte) (int, error) {
	return 0, nil
}

// ReadMessage 读取数据,当未设置DealFunc时生效
func (this *Server) ReadMessage() ([]byte, error) {
	m := <-this.readChan
	return m.Message, nil
}

// Write 给所有客户端发送数据,实现io.Writer接口
func (this *Server) Write(p []byte) (int, error) {
	this.WriteClientAll(p)
	return len(p), nil
}

// WriteClient 给一个客户端发送数据
func (this *Server) WriteClient(key string, msg []byte) (bool, error) {
	return this.GetClientDo(key, func(c *Client) error {
		_, err := c.Write(msg)
		return err
	})
}

// WriteClientAll 广播,发送数据给所有连接
func (this *Server) WriteClientAll(msg []byte) {
	for _, c := range this.GetClientMap() {
		c.Write(msg)
	}
}

// CloseClient 关闭一个连接
func (this *Server) CloseClient(key string) (bool, error) {
	return this.GetClientDo(key, func(c *Client) error { return c.Close() })
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

// Timer 定时执行
func (this *Server) Timer(interval time.Duration, do func(s *Server)) {
	go this.ICloser.Timer(interval, func() error {
		do(this)
		return nil
	})
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

// SwapServer 和另一个服务交换数据,客户端都的话存在数据重复发送客户端和速度瓶颈
func (this *Server) SwapServer(s *Server) *Server {
	this.SetDealWithWriter(s)
	s.SetDealWithWriter(this)
	return this
}

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

	//执行超时机制
	go func() {
		for {
			interval := conv.SelectDuration(this.timeoutInterval > time.Second, this.timeoutInterval, time.Second)
			<-time.After(interval)
			select {
			case <-this.ctx.Done():
			default:
				now := time.Now()
				for _, v := range this.GetClientMap() {
					if this.timeout > 0 && now.Sub(v.IReadCloser.LastTime()) > this.timeout {
						_ = v.CloseWithErr(ErrWithReadTimeout)
					}
				}
			}
		}
	}()

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

		//判断是否到达最大连接数,禁止新连接
		if this.clientMax > 0 && len(this.clientMap) >= this.clientMax {
			c.Write([]byte(fmt.Sprintf("超过最大连接数(%d)", len(this.clientMap))))
			c.Close()
			continue
		}

		//新建客户端,并配置
		x := NewClientWithContext(this.Ctx(), c)
		x.SetKey(key)                   //设置唯一标识符
		x.Debug(this.GetDebug())        //调试模式
		x.SetReadFunc(this.readFunc)    //读取数据方法
		x.SetDealFunc(this._dealFunc)   //数据处理方法
		x.SetCloseFunc(this._closeFunc) //连接关闭方法
		x.SetPrintFunc(this.printFunc)  //设置打印函数
		x.SetWriteFunc(this.writeFunc)  //设置发送函数

		// 协程执行,等待连接的后续数据,来决定后续操作
		go func(x *Client) {

			//前置操作,例如等待注册数据,不符合的返回错误则关闭连接
			if this.beforeFunc != nil && this.beforeFunc(x) != nil {
				_ = c.Close()
				return
			}

			//加入map 进行管理
			this.clientMu.Lock()
			this.clientMap[x.GetKey()] = x
			this.clientMu.Unlock()
			x.Run()

		}(x)

	}
}

//================================Inside================================

// _dealFunc 处理数据
func (this *Server) _dealFunc(msg *IMessage) {
	select {
	case <-this.Done():
	default:
		//加入消费队列
		this.dealQueue.Do(msg)
	}
}

// _closeFunc 删除连接
func (this *Server) _closeFunc(ctx context.Context, msg *IMessage) {
	if this.closeFunc != nil {
		defer this.closeFunc(msg)
	}
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	//获取老的连接
	oldConn := this.clientMap[msg.GetKey()]
	//存在新连接上来被关闭的情况,判断是否是老的连接
	if oldConn == nil || oldConn.Pointer() != msg.Pointer() {
		return
	}
	delete(this.clientMap, msg.GetKey())
}
