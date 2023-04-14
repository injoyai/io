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
func Redial(dial DialFunc, fn ...func(ctx context.Context, c *Client)) *Client {
	return RedialWithContext(context.Background(), dial, fn...)
}

// RedialWithContext 一直尝试连接,直到成功,需要输入上下文
func RedialWithContext(ctx context.Context, dial DialFunc, fn ...func(ctx context.Context, c *Client)) *Client {
	x := NewICloserWithContext(ctx, nil)
	x.Debug()
	x.SetRedialFunc(dial)
	x.SetKey(conv.String(dial))
	c := NewClientWithContext(ctx, x.Redial(ctx))
	c.SetRedialFunc(dial)
	c.Redial(fn...)
	return c
}

// NewDial 尝试连接,返回*Client和错误
func NewDial(dial DialFunc) (*Client, error) {
	return NewDialWithContext(context.Background(), dial)
}

// NewDialWithContext 尝试连接,返回*Client和错误,需要输入上下文
func NewDialWithContext(ctx context.Context, dial DialFunc) (*Client, error) {
	c, err := dial()
	if err != nil {
		return nil, err
	}
	cli := NewClientWithContext(ctx, c)
	cli.SetRedialFunc(dial)
	return cli, nil
}

// NewClient 标准库io.ReadWriterCloser转*Client
// 和隐性的MessageReadWriteCloser转*Client,后续1.18之后改成泛型
func NewClient(i ReadWriteCloser) *Client {
	return NewClientWithContext(context.Background(), i)
}

// NewClientWithContext 标准库io.ReadWriterCloser转*Client,需要输入上下文
func NewClientWithContext(ctx context.Context, i ReadWriteCloser) *Client {
	if c, ok := i.(*Client); ok && c != nil {
		return c
	}
	c := &Client{
		IReadCloser: NewIReadCloserWithContext(ctx, i),
		IWriter:     NewWriter(i),
		i:           i,
		tag:         maps.NewSafe(),
		createTime:  time.Now(),
	}
	c.SetKey(fmt.Sprintf("%p", i))
	return c
}

/*
Client 通用IO客户端
各种设置,当Run函数执行时生效
可以作为普通的io.ReadWriteCloser(Run函数不执行)
*/
type Client struct {
	*IReadCloser
	*IWriter

	i          ReadWriteCloser //接口,实例,传入的原始参数
	tag        *maps.Safe      //标签,用于记录连接的一些信息
	createTime time.Time       //创建时间
}

//================================Nature================================

// ReadWriteCloser 读写接口,实例,传入的原始参数
func (this *Client) ReadWriteCloser() io.ReadWriteCloser {
	return this.i
}

// Pointer 获取指针地址
func (this *Client) Pointer() string {
	return fmt.Sprintf("%p", this.ReadWriteCloser())
}

// CreateTime 创建时间
func (this *Client) CreateTime() time.Time {
	return this.createTime
}

// Tag 自定义信息,方便记录连接信息 例:c.Tag().GetString("imei")
func (this *Client) Tag() *maps.Safe {
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

// Debug 调试模式,打印日志
func (this *Client) Debug(b ...bool) *Client {
	this.IWriter.Debug(b...)
	this.IReadCloser.Debug(b...)
	return this
}

// WriteQueue 按队列写入
func (this *Client) WriteQueue(p []byte) {
	queue, _ := this.Tag().GetOrSetByHandler("_write_queue", func() (interface{}, error) {
		return this.IWriter.NewWriteQueue(this.Ctx()), nil
	})
	queue.(chan []byte) <- p
}

// WriteReadWithTimeout 同步写读,超时
func (this *Client) WriteReadWithTimeout(request []byte, timeout time.Duration) (response []byte, err error) {
	if _, err = this.Write(request); err != nil {
		return
	}
	return this.ReadLast(timeout)
}

// WriteRead 同步写读,不超时
func (this *Client) WriteRead(request []byte) (response []byte, err error) {
	return this.WriteReadWithTimeout(request, 0)
}

// GoTimerWriter 协程,定时写入数据,生命周期(一次链接,单次连接断开)
func (this *Client) GoTimerWriter(interval time.Duration, write func(c *IWriter) error) {
	go this.ICloser.Timer(interval, func() error {
		return write(this.IWriter)
	})
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

//================================SetFunc================================

// SetOptions 设置选项
func (this *Client) SetOptions(fn ...func(ctx context.Context, c *Client)) *Client {
	for _, v := range fn {
		v(this.Ctx(), this)
	}
	return this
}

// SetDealFunc 设置处理数据函数,默认响应ping>pong,忽略pong
func (this *Client) SetDealFunc(fn func(msg *IMessage)) {
	this.IReadCloser.SetDealFunc(func(msg Message) {
		switch msg.String() {
		case Ping:
			this.WriteString(Pong)
		case Pong:
		default:
			fn(NewIMessage(this, msg))
		}
	})
}

// SetCloseFunc 设置关闭函数
func (this *Client) SetCloseFunc(fn func(ctx context.Context, msg *IMessage)) {
	this.IReadCloser.SetCloseFunc(func(ctx context.Context, msg Message) {
		fn(ctx, NewIMessage(this, msg))
	})
}

// SetPrintFunc 设置打印函数
func (this *Client) SetPrintFunc(fn PrintFunc) *Client {
	this.IWriter.SetPrintFunc(fn)
	this.IReadCloser.SetPrintFunc(fn)
	return this
}

// SetPrintWithHEX 设置打印HEX
func (this *Client) SetPrintWithHEX() {
	this.SetPrintFunc(PrintWithHEX)
}

// SetPrintWithASCII 设置打印ASCII
func (this *Client) SetPrintWithASCII() {
	this.SetPrintFunc(PrintWithASCII)
}

// SetPrintWithBase 设置打印ASCII,基础信息
func (this *Client) SetPrintWithBase() {
	this.SetPrintFunc(PrintWithBase)
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
func (this *Client) Redial(fn ...func(ctx context.Context, c *Client)) *Client {
	this.SetCloseFunc(func(ctx context.Context, msg *IMessage) {
		<-time.After(time.Second)
		readWriteCloser := this.IReadCloser.Redial(ctx)
		if readWriteCloser == nil {
			this.ICloser.Print(NewMessageFormat(" 连接断开(%v),未设置重连或主动关闭", this.ICloser.Err()), TagErr, this.GetKey())
			return
		}
		this.ICloser.Print(NewMessageFormat("连接断开(%v),重连成功", this.ICloser.Err()), TagInfo, this.GetKey())
		redialFunc := this.IReadCloser.redialFunc
		key := this.GetKey()
		*this = *NewClient(readWriteCloser)
		this.SetKey(key)
		this.SetRedialFunc(redialFunc)
		this.Redial(fn...)
		go this.Run()
	})
	this.SetOptions(fn...)
	//新建客户端时已经能确定连接成功,为了让用户控制是否输出,所以在Run的时候打印
	this.Print(NewMessage("连接服务端成功..."), TagInfo, this.GetKey())
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
