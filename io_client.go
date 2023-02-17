package io

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"io"
	"sync"
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
func NewClient(i ReadWriteCloser) *Client {
	return NewClientWithContext(context.Background(), i)
}

// NewClientWithContext 标准库io.ReadWriterCloser转*Client,需要输入上下文
func NewClientWithContext(ctx context.Context, i ReadWriteCloser) *Client {
	if c, ok := i.(*Client); ok && c != nil {
		return c
	}
	c := &Client{
		buf: bufio.NewReader(i),
		i:   i,
		tag: maps.NewSafe(),

		IReadCloser: NewIReadCloserWithContext(ctx, i),
		IWriter:     NewWriter(i),
		IPrinter:    NewIPrinter(""),

		timerKeep: time.NewTimer(0),
		timer:     time.NewTimer(0),
		timeout:   0,
	}

	c.SetKey(fmt.Sprintf("%p", i))

	if c.timeout <= 0 {
		<-c.timer.C
	}
	if c.keepAlive <= 0 {
		<-c.timerKeep.C
	}
	return c
}

/*
Client 通用IO客户端
各种设置,当Run函数执行时生效
可以作为普通的io.ReadWriteCloser(Run函数不执行)
*/
type Client struct {
	buf *bufio.Reader   //buff
	key string          //唯一标识
	i   ReadWriteCloser //接口
	mu  sync.Mutex      //锁
	tag *maps.Safe      //标签

	*IReadCloser
	*IWriter
	*IPrinter

	timer      *time.Timer   //超时定时器,时间范围内没有发送数据或者接收数据,则断开链接
	timeout    time.Duration //超时时间
	timerKeep  *time.Timer   //正常通讯不发送心跳
	keepAlive  time.Duration //保持连接
	createTime time.Time     //创建时间,链接成功时间
}

//================================Nature================================

// ReadWriteCloser 读写接口
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

// Buffer 极大的增加读取速度
func (this *Client) Buffer() *bufio.Reader {
	return this.buf
}

// Tag 自定义信息
func (this *Client) Tag() *maps.Safe {
	return this.tag
}

// Debug 调试模式,打印日志
func (this *Client) Debug(b ...bool) *Client {
	this.IPrinter.Debug(b...)
	this.IWriter.Debug(b...)
	this.IReadCloser.Debug(b...)
	return this
}

// WriteReadWithTimeout 同步写读,超时
func (this *Client) WriteReadWithTimeout(request []byte, timeout time.Duration) (response []byte, err error) {
	if _, err = this.WriteWithTimeout(request, timeout); err != nil {
		return
	}
	return this.ReadLast(timeout)
}

// WriteRead 同步写读,不超时
func (this *Client) WriteRead(request []byte) (response []byte, err error) {
	return this.WriteReadWithTimeout(request, 0)
}

// GoForWriter 协程执行周期写入数据
func (this *Client) GoForWriter(interval time.Duration, write func(c *IWriter) (int, error)) {
	go func(ctx context.Context, writer *IWriter) {
		t := time.NewTimer(interval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if _, err := write(writer); err != nil {
					return
				}
				t.Reset(interval)
			}
		}
	}(this.Ctx(), this.IWriter)
}

//================================SetFunc================================

// GetTag 获取一个tag
func (this *Client) GetTag(key interface{}) interface{} {
	return this.tag.MustGet(key)
}

// SetTag 设置一个tag
func (this *Client) SetTag(key, value interface{}) {
	this.tag.Set(key, value)
}

// SetKey 设置唯一标识
func (this *Client) SetKey(key string) *Client {
	this.key = key
	this.IWriter.SetKey(key)
	this.IReadCloser.SetKey(key)
	return this
}

// GetKey 获取唯一标识
func (this *Client) GetKey() string {
	return this.key
}

// SetTimeout 设置超时时间
func (this *Client) SetTimeout(timeout time.Duration) *Client {
	this.timeout = timeout
	if timeout <= 0 {
		this.timer.Stop()
	} else {
		this.timer.Reset(timeout)
	}
	return this
}

// SetDealFunc 设置处理数据函数
func (this *Client) SetDealFunc(fn func(msg *ClientMessage)) {
	this.IReadCloser.SetDealFunc(func(msg Message) {
		fn(NewClientMessage(this, msg))
	})
}

// SetCloseFunc 设置关闭函数
func (this *Client) SetCloseFunc(fn func(ctx context.Context, msg *ClientMessage)) {
	this.IReadCloser.SetCloseFunc(func(ctx context.Context, msg Message) {
		fn(ctx, NewClientMessage(this, msg))
	})
}

// SetPrintFunc 设置打印函数
func (this *Client) SetPrintFunc(fn PrintFunc) *Client {
	this.IPrinter.SetPrintFunc(fn)
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

// SetKeepAlive 设置连接保持,另外起了携程,服务器不需要,客户端再起一个也没啥问题
// TCP keepalive定义于RFC 1122，但并不是TCP规范中的一部分,默认必需是关闭,连接方不一定支持
func (this *Client) SetKeepAlive(t time.Duration, keeps ...[]byte) *Client {
	keep := conv.GetDefaultBytes([]byte(Ping), keeps...)
	old := this.keepAlive
	this.keepAlive = t
	if old == 0 && this.keepAlive > 0 {
		this.timerKeep.Reset(this.keepAlive)
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case <-this.timerKeep.C:
					if this.keepAlive <= 0 {
						return
					}
					if _, err := this.Write(keep); err != nil {
						return
					}
				}
			}
		}(this.Ctx())
	}
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
	this.SetCloseFunc(func(ctx context.Context, msg *ClientMessage) {
		readWriteCloser := this.IReadCloser.Redial(ctx)
		if readWriteCloser == nil {
			this.IPrinter.Print(NewMessageFormat(" 连接断开(%v),未设置重连或主动关闭", this.ICloser.Err()), TagClose, this.GetKey())
			return
		}
		this.IPrinter.Print(NewMessageFormat("连接断开(%v),重连成功", this.ICloser.Err()), TagErr, this.GetKey())
		redialFunc := this.IReadCloser.redialFunc
		key := this.GetKey()
		*this = *NewClient(readWriteCloser)
		this.SetKey(key)
		this.SetRedialFunc(redialFunc)
		this.Redial(fn...)
		go this.Run()
	})
	for _, v := range fn {
		v(this.Ctx(), this)
	}
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

//================================RunTime================================

// Closed 是否断开连接
func (this *Client) Closed() bool {
	return this.IReadCloser.Closed()
}

// CloseAll 主动关闭连接,无法触发重试机制
func (this *Client) CloseAll() error {
	this.IReadCloser.CloseAll()
	return nil
}

// Close 手动断开,会触发重试
func (this *Client) Close() error {
	return this.CloseWithErr(ErrHandClose)
}

// CloseWithErr 根据错误关闭
func (this *Client) CloseWithErr(err error) error {
	return this.IReadCloser.CloseWithErr(err)
}

// Err 错误信息,默认有个错误,如果连接正常,错误为默认,则返回nil
func (this *Client) Err() error {
	return this.IReadCloser.Err()
}

// Run 开始执行(读取数据)
func (this *Client) Run() error {
	return this.IReadCloser.Run()
}
