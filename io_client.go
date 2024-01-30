package io

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"io"
	"time"
)

// Redial 一直连接,直到成功
func Redial(dial DialFunc, options ...OptionClient) *Client {
	return RedialWithContext(context.Background(), dial, options...)
}

// RedialWithContext 一直尝试连接,直到成功,需要输入上下文
func RedialWithContext(ctx context.Context, dial DialFunc, options ...OptionClient) *Client {
	x := NewICloserWithContext(ctx, nil)
	x.Logger.Debug()
	x.SetRedialFunc(dial)
	x.SetKey(conv.String(dial))
	r, key := x.MustDial(ctx)
	return NewClientWithContext(ctx, r, func(c *Client) {
		c.SetKey(key)
		c.SetRedialFunc(dial)
		c.Redial(options...)
		//用户控制输出,需要在SetOptions之后打印
		c.Logger.Infof("[%s] 连接服务端成功...\n", c.GetKey())
	})
}

// NewDial 尝试连接,返回*Client和错误
func NewDial(dial DialFunc, options ...OptionClient) (*Client, error) {
	return NewDialWithContext(context.Background(), dial, options...)
}

// NewDialWithContext 尝试连接,返回*Client和错误,需要输入上下文
func NewDialWithContext(ctx context.Context, dial DialFunc, options ...OptionClient) (*Client, error) {
	c, key, err := dial(ctx)
	if err != nil {
		return nil, err
	}
	cli := NewClientWithContext(ctx, c, func(c *Client) {
		c.SetKey(key)
		c.SetRedialFunc(dial)
		c.SetOptions(options...)
		if !c.Closed() {
			//如果在option关闭的话,会先打印关闭,再打印连接成功,所以判断下连接是否还在
			//用户控制输出,需要在SetOptions之后打印
			c.Logger.Infof("[%s] 连接服务端成功...\n", c.GetKey())
		}
	})
	return cli, nil
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
	c := &Client{
		Key:         "",
		IReadCloser: NewIReadCloserWithContext(ctx, i),
		IWriter:     NewIWriter(i),
		i:           i,
		tag:         nil,
		createTime:  time.Now(),
	}
	c.SetKey(fmt.Sprintf("%p", i))
	c.Debug()
	c.SetOptions(options...)
	return c
}

/*
Client 通用IO客户端
各种设置,当Run函数执行时生效
可以作为普通的io.ReadWriteCloser(Run函数不执行)
*/
type Client struct {
	Key
	*IReadCloser
	*IWriter

	pointer    *string         //唯一标识,指针地址
	i          ReadWriteCloser //接口,实例,传入的原始参数
	tag        *maps.Safe      //标签,用于记录连接的一些信息
	createTime time.Time       //创建时间
}

//================================Nature================================

// ReadLastTime 最后读取时间
func (this *Client) ReadLastTime() time.Time {
	return this.IReader.LastTime()
}

// WriteLastTime 最后写入时间
func (this *Client) WriteLastTime() time.Time {
	return this.IWriter.LastTime()
}

// ReadBytesCount 读取的字节数
func (this *Client) ReadBytesCount() int64 {
	return this.IReader.BytesCount()
}

// WriteBytesCount 写入的字节数
func (this *Client) WriteBytesCount() int64 {
	return this.IWriter.BytesCount()
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
		pointer := fmt.Sprintf("%p", this.ReadWriteCloser())
		this.pointer = &pointer
	}
	return *this.pointer
}

// CreateTime 创建时间
func (this *Client) CreateTime() time.Time {
	return this.createTime
}

// Tag 自定义信息,方便记录连接信息 例:c.Tag().GetString("imei")
func (this *Client) Tag() *maps.Safe {
	if this.tag == nil {
		this.tag = maps.NewSafe()
	}
	return this.tag
}

// SetKey 设置唯一标识
func (this *Client) SetKey(key string) *Client {
	this.IWriter.SetKey(key)
	this.IReadCloser.SetKey(key)
	return this
}

// GetKey 获取唯一标识
func (this *Client) GetKey() string {
	return this.IReadCloser.GetKey()
}

// WriteQueue 按队列写入
func (this *Client) WriteQueue(p []byte) *Client {
	queue, _ := this.Tag().GetOrSetByHandler(writeQueueKey, func() (interface{}, error) {
		return this.IWriter.NewWriteQueue(this.Ctx()), nil
	})
	queue.(chan []byte) <- p
	return this
}

// TryWriteQueue 尝试按队列写入,加入不了会丢弃
func (this *Client) TryWriteQueue(p []byte) *Client {
	queue, _ := this.Tag().GetOrSetByHandler(writeQueueKey, func() (interface{}, error) {
		return this.IWriter.NewWriteQueue(this.Ctx()), nil
	})
	select {
	case queue.(chan []byte) <- p:
	default:
	}
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

func (this *Client) Ping(timeout ...time.Duration) error {
	_, err := this.WriteRead([]byte(Ping), conv.DefaultDuration(time.Second, timeout...))
	return err
}

// GoTimerWriter 协程,定时写入数据,生命周期(一次链接,单次连接断开)
func (this *Client) GoTimerWriter(interval time.Duration, write func(w *IWriter) error) {
	go this.ICloser.Timer(interval, func() error {
		return write(this.IWriter)
	})
}

// GoTimerWriteBytes 协程,定时写入字节数据
func (this *Client) GoTimerWriteBytes(interval time.Duration, p []byte) {
	this.GoTimerWriter(interval, func(w *IWriter) error {
		_, err := w.Write(p)
		return err
	})
}

// GoTimerWriteString 协程,定时写入字符数据
func (this *Client) GoTimerWriteString(interval time.Duration, s string) {
	this.GoTimerWriter(interval, func(w *IWriter) error {
		_, err := w.WriteString(s)
		return err
	})
}

// GoAfter 延迟执行函数
func (this *Client) GoAfter(after time.Duration, fn func()) {
	go this.ICloser.After(after, fn)
}

// SetKeepAlive 设置连接保持,另外起了携程,服务器不需要,客户端再起一个也没啥问题
// TCP keepalive定义于RFC 1122，但并不是TCP规范中的一部分,默认必需是关闭,连接方不一定支持
func (this *Client) SetKeepAlive(t time.Duration, keeps ...[]byte) {
	this.GoTimerWriter(t, func(c *IWriter) error {
		keep := conv.GetDefaultBytes([]byte(Ping), keeps...)
		_, err := c.Write(keep)
		return err
	})
}

//================================Logger================================

// Debug 调试模式,打印日志
// 为了实现Debugger接口,不需要返回值
func (this *Client) Debug(b ...bool) {
	this.IWriter.Logger.Debug(b...)
	this.IReadCloser.Debug(b...)
}

// SetLogger 设置日志
func (this *Client) SetLogger(logger Logger) *Client {
	l := newLogger(logger)
	this.Logger = l
	this.IWriter.Logger = l
	this.IReadCloser.IReader.Logger = l
	this.IReadCloser.ICloser.Logger = l
	return this
}

// SetPrintWithHEX 设置打印HEX
func (this *Client) SetPrintWithHEX() *Client {
	this.IWriter.Logger.SetPrintWithHEX()
	this.IReadCloser.SetPrintWithHEX()
	return this
}

// SetPrintWithUTF8 设置打印编码utf-8
func (this *Client) SetPrintWithUTF8() *Client {
	this.IWriter.Logger.SetPrintWithUTF8()
	this.IReadCloser.SetPrintWithUTF8()
	return this
}

// SetLevel 设置日志等级
func (this *Client) SetLevel(level Level) *Client {
	this.IWriter.Logger.SetLevel(level)
	this.IReadCloser.SetLevel(level)
	return this
}

// SetPrintWithAll 设置打印等级为全部
func (this *Client) SetPrintWithAll() *Client {
	return this.SetLevel(LevelAll)
}

// SetPrintWithBase 设置打印基础信息
func (this *Client) SetPrintWithBase() *Client {
	return this.SetLevel(LevelInfo)
}

// SetPrintWithErr 设置打印错误信息
func (this *Client) SetPrintWithErr() *Client {
	return this.SetLevel(LevelError)
}

//================================SetFunc================================

// SetOptions 设置选项
func (this *Client) SetOptions(options ...OptionClient) *Client {
	for _, v := range options {
		v(this)
	}
	return this
}

// SetDealFunc 设置处理数据函数,默认响应ping>pong,忽略pong
func (this *Client) SetDealFunc(fn func(c *Client, msg Message)) *Client {
	this.IReadCloser.SetDealFunc(func(msg Message) {
		switch msg.String() {
		case Ping:
			this.WriteString(Pong)
		case Pong:
		default:
			fn(this, msg)
		}
	})
	return this
}

// SetCloseFunc 设置关闭函数
func (this *Client) SetCloseFunc(fn func(ctx context.Context, c *Client, msg Message)) *Client {
	this.IReadCloser.SetCloseFunc(func(ctx context.Context, msg Message) {
		fn(ctx, this, msg)
	})
	return this
}

// SetReadWriteWithPkg 设置读写为默认分包方式
func (this *Client) SetReadWriteWithPkg() *Client {
	this.IWriter.SetWriteWithPkg()
	this.IReader.SetReadWithPkg()
	return this
}

// SetReadWriteWithStartEnd 设置读取写入数据根据包头包尾
func (this *Client) SetReadWriteWithStartEnd(packageStart, packageEnd []byte) *Client {
	this.IWriter.SetWriteWithStartEnd(packageStart, packageEnd)
	this.IReadCloser.SetReadWithStartEnd(packageStart, packageEnd)
	return this
}

// Redial 重新链接,重试,因为指针复用,所以需要根据上下文来处理(例如关闭)
func (this *Client) Redial(options ...OptionClient) *Client {
	this.SetCloseFunc(func(ctx context.Context, c *Client, msg Message) {
		<-time.After(time.Second)
		readWriteCloser, key := this.IReadCloser.MustDial(ctx)
		if readWriteCloser == nil {
			if this.ICloser.Err() != ErrHandClose {
				this.Logger.Errorf("[%s] 连接断开(%v),未设置重连函数\n", this.GetKey(), this.ICloser.Err())
			}
			return
		}
		this.Logger.Infof("[%s] 连接断开(%v),重连成功\n", this.GetKey(), this.ICloser.Err())
		redialFunc := this.IReadCloser.redialFunc
		//key := this.GetKey()
		*this = *NewClient(readWriteCloser)
		this.SetKey(key)
		this.SetRedialFunc(redialFunc)
		this.Redial(options...)
		go this.Run()
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

func (this *Client) Run() error {
	return this.IReadCloser.Run()
}
