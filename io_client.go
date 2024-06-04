package io

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"io"
	"net"
	"sync"
	"time"
)

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
	err := c.MustDial(ctx, func(c *Client) {
		c.Redial(options...)
	})
	c.CloseWithErr(err)
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
以客户端的指针为唯一标识,key作为辅助标识(展示给用户看)
各种设置,当Run函数执行时生效
可以作为普通的io.ReadWriteCloser(Run函数不执行)
*/
type Client struct {

	//reader
	latestChan chan Message //读取最新数据chan

	//writer
	writeQueue     chan []byte //写入队列
	writeQueueOnce sync.Once   //写入队列初始化

	//closer
	redialMaxTime time.Duration //最大尝试退避重连时间
	redialMaxNum  int           //最大尝试重连的次数
	dialFunc      DialFunc      //连接函数

	timeout      time.Duration      //超时时间,读取
	timeoutReset chan struct{}      //超时重置
	running      uint32             //是否在运行,1是运行,0是没运行
	closed       uint32             //是否关闭(不公开,做原子操作),0是未关闭,1是已关闭
	closeErr     error              //错误信息
	ctx          context.Context    //子级上下文
	cancel       context.CancelFunc //子级上下文
	ctxParent    context.Context    //父级上下文,主动关闭时,用于关闭redial,好像没啥用得用一个协程来监听
	cancelParent context.CancelFunc //父级上下文,主动关闭时,用于关闭redial

	//runtime
	Key                         //自定义标识
	*logger                     //日志
	pointer     string          //唯一标识,指针地址
	i           ReadWriteCloser //接口,实例,传入的原始参数
	buf         *bufio.Reader   //buffer
	tag         *maps.Safe      //标签,用于记录连接的一些信息
	createTime  time.Time       //创建时间
	readTime    time.Time       //最后读取时间
	readBytes   int64           //读取的字节数
	writeTime   time.Time       //最后写入时间
	writeBytes  int64           //写入的字节数
	writeNumber int64           //写入的次数

	//连接成功事件,可以手动进行数据的读写,或者关闭,返回错误会关闭连接
	//如果设置了重连,则会再次建立连接而触发连接事件
	//所以固定返回错误的话,会陷入无限连接断开的情况,
	connectFunc []func(c *Client) error

	//连接断开事件,连接断开的时候触发,可以调用Dial方法进行重连操作
	closeFunc []func(ctx context.Context, c *Client, err error) //关闭函数

	readFunc func(buf *bufio.Reader) (Acker, error) //读取函数

	//处理数据事件,可以进行打印或者其他逻辑操作
	dealFunc []func(c *Client, msg Message) (ack bool)

	//写入数据事件,可以进行封装或者打印等操作
	writeFunc []func(p []byte) ([]byte, error)
}

//================================Nature================================

func (this *Client) reset(i ReadWriteCloser, key string, options ...OptionClient) *Client {
	if v, ok := i.(*Client); ok {
		this.reset(v.i, key, options...)
	}
	this.ctx, this.cancel = context.WithCancel(this.ctxParent)
	this.latestChan = make(chan Message)
	this.redialMaxTime = time.Second * 32
	this.timeoutReset = make(chan struct{})
	this.logger = defaultLogger()
	this.running = 0
	this.closed = 0
	this.i = i
	this.tag = nil
	//this.closeErr = nil

	defaultReadFunc := ReadFuncToAck(buf.Read1KB)
	switch v := i.(type) {
	case nil:
	case MessageReader:
		mReader := MReaderToReader(v)
		defaultReadFunc = ReadFuncToAck(mReader.ReadFunc)
		i = struct {
			WriteCloser
			Reader
		}{
			WriteCloser: i,
			Reader:      mReader,
		}

	case AckReader:
		aReader := AReaderToReader(v)
		defaultReadFunc = aReader.ReadAck
		i = struct {
			WriteCloser
			Reader
		}{
			WriteCloser: i,
			Reader:      aReader,
		}

	}
	//默认buf大小,可自定义缓存大小
	this.buf = bufio.NewReaderSize(i, DefaultBufferSize+1)

	//设置默认的Option
	this.SetKey(key) //唯一标识
	this.Debug()     //打印日志

	//设置默认事件
	this.SetConnectWithNil().SetConnectWithLog()
	this.SetReadAckFunc(defaultReadFunc)
	this.SetDealWithNil().SetDealWithDefault()
	this.SetWriteWithNil().SetWriteWithLog()
	this.SetCloseWithNil().SetCloseWithLog()

	//设置用户的Option
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
	if this.pointer == "" {
		pointer := fmt.Sprintf("%p", this)
		this.pointer = pointer
	}
	return this.pointer
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

// WriteRead 同步写读,写入数据,并监听,例如串口
func (this *Client) WriteRead(request []byte, timeout ...time.Duration) ([]byte, error) {
	if _, err := this.Write(request); err != nil {
		return nil, err
	}
	return this.ReadLatest(conv.GetDefaultDuration(DefaultResponseTimeout, timeout...))
}

// Ping 测试连接, 在默认处理时候会返回pong才有用
func (this *Client) Ping(timeout ...time.Duration) error {
	resp, err := this.WriteRead([]byte(Ping), conv.DefaultDuration(time.Second, timeout...))
	if err != nil {
		return err
	}
	if string(resp) != Pong {
		return errors.New("ping error")
	}
	return nil
}

// SetOptions 设置选项
func (this *Client) SetOptions(options ...OptionClient) *Client {
	for _, v := range options {
		v(this)
	}
	return this
}

//================================net.Conn================================

// NetConn 断言net.Conn
func (this *Client) NetConn() (net.Conn, bool) {
	v, ok := this.i.(net.Conn)
	return v, ok
}
