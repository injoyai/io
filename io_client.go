package io

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"io"
	"net"
	"sync"
	"time"
)

//// Redial 一直连接,直到成功
//func Redial(dial DialFunc, options ...OptionClient) *Client {
//	return RedialWithContext(context.Background(), dial, options...)
//}
//
//// RedialWithContext 一直尝试连接,直到成功,需要输入上下文
//func RedialWithContext(ctx context.Context, dial DialFunc, options ...OptionClient) *Client {
//	x := NewICloserWithContext(ctx, nil)
//	x.Logger.Debug()
//	x.SetRedialFunc(dial)
//	x.SetKey(conv.String(dial))
//	r, key := x.MustDial(ctx)
//
//	return NewClientWithContext(ctx, r, func(c *Client) {
//		c.SetKey(key)
//		c.SetDialFunc(dial)
//		c.Redial(options...)
//		//用户控制输出,需要在SetOptions之后打印
//		c.Logger.Infof("[%s] 连接服务端成功...\n", c.GetKey())
//	})
//}

// Redial 一直连接,直到成功
func Redial(dial DialFunc, options ...OptionClient) *Client {
	return RedialWithContext(context.Background(), dial, options...)
}

// RedialWithContext 一直尝试连接,直到成功,需要输入上下文
func RedialWithContext(ctx context.Context, dial DialFunc, options ...OptionClient) *Client {
	ctxParent, cancelParent := context.WithCancel(ctx)
	c := &Client{
		ctxParent:    ctxParent,
		cancelParent: cancelParent,
	}
	c.SetDialFunc(dial)
	c.MustDial(func(c *Client) {
		c.Redial(options...)
	})
	return c
}

// NewDial 尝试连接,返回*Client和错误
func NewDial(dial DialFunc, options ...OptionClient) (*Client, error) {
	return NewDialWithContext(context.Background(), dial, options...)
}

// NewDialWithContext 尝试连接,返回*Client和错误,需要输入上下文
func NewDialWithContext(ctx context.Context, dial DialFunc, options ...OptionClient) (*Client, error) {
	ctxParent, cancelParent := context.WithCancel(ctx)
	c := &Client{
		ctxParent:    ctxParent,
		cancelParent: cancelParent,
	}
	c.SetDialFunc(dial)
	err := c.Dial(options...)
	return c, err
}

// NewClient 标准库io.ReadWriterCloser转*Client
// 和隐性的MessageReadWriteCloser转*Client,后续1.18之后改成泛型
func NewClient(i ReadWriteCloser, options ...OptionClient) *Client {
	return NewClientWithContext(context.Background(), i, options...)
}

// NewClientWithContext 标准库io.ReadWriterCloser转*Client,需要输入上下文
func NewClientWithContext(ctx context.Context, i ReadWriteCloser, options ...OptionClient) *Client {
	if c, ok := i.(*Client); ok && c != nil {
		return c
	}
	ctxParent, cancelParent := context.WithCancel(ctx)
	c := &Client{
		ctxParent:    ctxParent,
		cancelParent: cancelParent,
	}
	c.reset(i, c.Pointer(), options...)
	return c
}

/*
Client 通用IO客户端
各种设置,当Run函数执行时生效
可以作为普通的io.ReadWriteCloser(Run函数不执行)
*/
type Client struct {

	//reader
	readFunc func(buf *bufio.Reader) ([]byte, error) //读取函数
	dealFunc func(msg Message)                       //处理数据函数
	readChan chan Message                            //读取最新数据chan
	mReader  MessageReader                           //接口MessageReader,兼容Reader

	//writer
	writeFunc      func(p []byte) ([]byte, error) //写入函数,处理写入内容
	writeQueue     chan []byte                    //写入队列
	writeQueueOnce sync.Once                      //写入队列初始化

	//closer
	redialMaxTime time.Duration                                     //最大尝试退避重连时间
	dialFunc      DialFunc                                          //连接函数
	closeFunc     func(ctx context.Context, c *Client, msg Message) //关闭函数
	timeout       time.Duration                                     //超时时间,读取
	timeoutReset  chan struct{}                                     //超时重置
	running       uint32                                            //是否在运行
	closed        uint32                                            //是否关闭(不公开,做原子操作),0是未关闭,1是已关闭
	closeErr      error                                             //错误信息
	ctx           context.Context                                   //子级上下文
	cancel        context.CancelFunc                                //子级上下文
	ctxParent     context.Context                                   //父级上下文,主动关闭时,用于关闭redial
	cancelParent  context.CancelFunc                                //父级上下文,主动关闭时,用于关闭redial

	//runtime
	Key                         //自定义标识
	*logger                     //日志
	pointer     *string         //唯一标识,指针地址
	i           ReadWriteCloser //接口,实例,传入的原始参数
	buf         *bufio.Reader   //buffer
	tag         *maps.Safe      //标签,用于记录连接的一些信息
	createTime  time.Time       //创建时间
	readTime    time.Time       //最后读取时间
	readBytes   int64           //读取的字节数
	readNumber  int64           //读取的次数
	writeTime   time.Time       //最后写入时间
	writeBytes  int64           //写入的字节数
	writeNumber int64           //写入的次数
}

//================================Nature================================

func (this *Client) reset(i ReadWriteCloser, key string, options ...OptionClient) *Client {
	if v, ok := i.(*Client); ok {
		this.reset(v.i, key, options...)
	}
	this.ctx, this.cancel = context.WithCancel(this.ctxParent)
	this.readChan = make(chan Message)
	this.redialMaxTime = time.Second * 32
	this.timeoutReset = make(chan struct{})
	this.logger = defaultLogger()
	this.running = 0
	this.closed = 0
	this.mReader = nil
	this.buf = nil
	this.tag = nil
	//this.closeErr = nil
	switch v := i.(type) {
	case nil:
	case MessageReader:
		this.mReader = v
		this.i = struct {
			*bytes.Buffer
			Closer
		}{
			Buffer: bytes.NewBuffer(nil),
			Closer: &closer{},
		}
	default:
		this.i = i
	}
	//todo 优化缓存大小可配置
	this.buf = bufio.NewReaderSize(i, DefaultBufferSize+1)

	this.SetKey(key)
	this.Debug()
	this.SetReadFunc(buf.Read1KB)
	this.SetOptions(options...)
	return this
}

// ReadWriteCloser 读写接口,实例,传入的原始参数
func (this *Client) ReadWriteCloser() io.ReadWriteCloser {
	return this.i
}

// ID 获取唯一标识,不可改,溯源,不用因为改了key而出现数据错误的bug
func (this *Client) ID() string {
	return this.Pointer()
}

// Pointer 获取指针地址
func (this *Client) Pointer() string {
	if this.pointer == nil {
		pointer := fmt.Sprintf("%p", this)
		this.pointer = &pointer
	}
	return *this.pointer
}

// CreateTime 创建时间
func (this *Client) CreateTime() time.Time {
	return this.createTime
}

func (this *Client) GetTag(key string, def ...string) string {
	return this.Tag().GetString(key, def...)
}

// Tag 自定义信息,方便记录连接信息 例:c.Tag().GetString("imei")
func (this *Client) Tag() *maps.Safe {
	if this.tag == nil {
		this.tag = maps.NewSafe()
	}
	return this.tag
}

// SetLogger 设置日志
func (this *Client) SetLogger(logger Logger) *Client {
	this.logger = NewLogger(logger)
	return this
}

// WriteReadWithTimeout 同步写读,超时
func (this *Client) WriteReadWithTimeout(request []byte, timeout time.Duration) (response []byte, err error) {
	if _, err = this.Write(request); err != nil {
		return
	}
	return this.ReadLast(timeout)
}

// WriteRead 同步写读
func (this *Client) WriteRead(request []byte, timeout ...time.Duration) (response []byte, err error) {
	return this.WriteReadWithTimeout(request, conv.GetDefaultDuration(DefaultResponseTimeout, timeout...))
}

// Ping 测试连接
func (this *Client) Ping(timeout ...time.Duration) error {
	_, err := this.WriteRead([]byte(Ping), conv.DefaultDuration(time.Second, timeout...))
	return err
}

// GoAfter 延迟执行函数
func (this *Client) GoAfter(after time.Duration, fn func()) {
	go this.After(after, fn)
}

// SetOptions 设置选项
func (this *Client) SetOptions(options ...OptionClient) *Client {
	for _, v := range options {
		v(this)
	}
	return this
}

//================================SetFunc================================

// SetReadWriteWithPkg 设置读写为默认分包方式
func (this *Client) SetReadWriteWithPkg() *Client {
	this.SetWriteWithPkg()
	this.SetReadWithPkg()
	return this
}

// SetReadWriteWithSimple 设置读写为简易包
func (this *Client) SetReadWriteWithSimple() *Client {
	this.SetWriteFunc(WriteWithSimple)
	this.SetReadFunc(ReadWithSimple)
	return nil
}

// SetReadWriteWithStartEnd 设置读取写入数据根据包头包尾
func (this *Client) SetReadWriteWithStartEnd(packageStart, packageEnd []byte) *Client {
	this.SetWriteWithStartEnd(packageStart, packageEnd)
	this.SetReadWithStartEnd(packageStart, packageEnd)
	return this
}

// Redial 重新链接,重试,因为指针复用,所以需要根据上下文来处理(例如关闭)
func (this *Client) Redial(options ...OptionClient) *Client {
	this.SetCloseFunc(func(ctx context.Context, c *Client, msg Message) {
		<-time.After(time.Second)
		this.MustDial(func(c *Client) {
			c.Redial(options...)
		})
	})
	this.SetOptions(options...)
	//新建客户端时已经能确定连接成功,为了让用户控制是否输出,所以在Run的时候打印
	//this.Logger.Infof("[%s] 连接服务端成功...\n", this.GetKey())
	go this.Run()
	return this
}

// Swap IO数据交换
func (this *Client) Swap(i ReadWriteCloser) {
	this.SwapClient(NewClient(i))
}

// SwapClient IO数据交换
func (this *Client) SwapClient(c *Client) {
	SwapClient(this, c)
}

//================================net.Conn================================

// NetConn 断言net.Conn
func (this *Client) NetConn() (net.Conn, bool) {
	v, ok := this.i.(net.Conn)
	return v, ok
}

func (this *Client) LocalAddr() net.Addr {
	if v, ok := this.i.(net.Conn); ok {
		return v.LocalAddr()
	}
	return &net.TCPAddr{}
}

func (this *Client) RemoteAddr() net.Addr {
	if v, ok := this.i.(net.Conn); ok {
		return v.RemoteAddr()
	}
	return &net.TCPAddr{}
}

func (this *Client) SetDeadline(t time.Time) error {
	if v, ok := this.i.(net.Conn); ok {
		return v.SetDeadline(t)
	}
	return nil
}

func (this *Client) SetReadDeadline(t time.Time) error {
	if v, ok := this.i.(net.Conn); ok {
		return v.SetReadDeadline(t)
	}
	return nil
}

func (this *Client) SetWriteDeadline(t time.Time) error {
	if v, ok := this.i.(net.Conn); ok {
		return v.SetWriteDeadline(t)
	}
	return nil
}
