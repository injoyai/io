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
	c := newClient(ctx)
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
	c := newClient(ctx)
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
	c := newClient(ctx)
	c.reset(i, c.ID(), options...)
	return c
}

func newClient(ctx context.Context) *Client {
	ctxParent, cancelParent := context.WithCancel(ctx)
	return &Client{
		ctxParent:     ctxParent,
		cancelParent:  cancelParent,
		CreateTime:    time.Now(),
		logger:        defaultLogger(),
		redialMaxTime: time.Second * 32,
	}
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
	Key         string          //自定义标识
	*logger                     //日志
	pointer     string          //唯一标识,指针地址
	i           ReadWriteCloser //接口,实例,传入的原始参数
	buf         *bufio.Reader   //buffer
	tag         *maps.Safe      //标签,用于记录连接的一些信息
	CreateTime  time.Time       //创建时间,对象创建时间,重连不会改变
	DialTime    time.Time       //连接时间,每次重连会改变
	ReadTime    time.Time       //本次连接,最后读取到数据的时间
	ReadCount   uint64          //本次连接,读取的字节数量
	WriteTime   time.Time       //本次连接,最后写入数据时间
	WriteCount  uint64          //本次连接,写入的字节数量
	WriteNumber uint64          //本次连接,写入的次数

	//连接成功事件,可以手动进行数据的读写,或者关闭,返回错误会关闭连接
	//如果设置了重连,则会再次建立连接而触发连接事件
	//所以固定返回错误的话,会陷入无限连接断开的情况,
	connectFunc []func(c *Client) error

	//连接断开事件,连接断开的时候触发,可以调用Dial方法进行重连操作
	//怕用户设置多个redial,重连之后越连越多,固只设置一个函数
	closeFunc func(ctx context.Context, c *Client, err error)

	//从流中读取数据
	readFunc func(buf *bufio.Reader) (Acker, error)

	//处理数据事件,可以进行打印或者其他逻辑操作
	dealFunc []func(c *Client, msg Message) (ack bool)

	//写入数据事件,可以进行封装或者打印等操作
	writeFunc []func(p []byte) ([]byte, error)

	//当写入结束时出发
	writeResultFunc []func(c *Client, err error)

	//当key变化时触发
	keyChangeFunc []func(c *Client, oldKey string)
}

//================================Nature================================

func (this *Client) reset(i ReadWriteCloser, key string, options ...OptionClient) *Client {
	if v, ok := i.(*Client); ok {
		this.reset(v.i, key, options...)
	}

	this.latestChan = make(chan Message)
	this.writeQueueOnce = sync.Once{}
	this.writeQueue = nil

	this.redialMaxTime = time.Second * 32
	this.redialMaxNum = 0
	this.dealFunc = nil

	this.timeout = 0
	this.timeoutReset = make(chan struct{})
	this.running = 0
	this.closed = 0
	//错误初始化,初始化会执行用户option,
	//例如用户在option中设置了Close
	//初始化之后会判断错误信息
	//所以这里得初始化错误
	this.closeErr = nil
	this.ctx, this.cancel = context.WithCancel(this.ctxParent)
	//父级上下文保留
	//this.ctxParent = this.ctxParent
	//this.cancelParent= this.cancelParent

	this.Key = key
	//在外部声明,连接失败的时候需要打印日志,还没到初始化这一步
	this.logger = defaultLogger()
	//还是原来的对象,使用的原先的指针
	//this.pointer=this.pointer
	this.i = i
	//buf在下面的处理
	//this.buf
	this.tag = nil //是否保留tag信息,重新设置就能覆盖
	//使用的是第一次连接的时间
	//this.CreateTime=this.CreateTime
	//当连接成功的时候会进行重置操作
	this.DialTime = time.Now()
	this.ReadTime = time.Time{}
	this.ReadCount = 0
	this.WriteTime = time.Time{}
	this.WriteCount = 0
	this.WriteNumber = 0

	//初始化事件函数
	this.connectFunc = nil
	this.closeFunc = nil
	this.readFunc = nil
	this.dealFunc = nil
	this.writeFunc = nil
	this.writeResultFunc = nil
	this.keyChangeFunc = nil

	/*



	 */

	defaultReadFunc := ReadFuncToAck(buf.Read1KB)
	switch v := i.(type) {
	case nil:

	case AReader:
		aReader := AReaderToReader(v)
		defaultReadFunc = aReader.ReadAck
		i = NewReadWriteCloser(aReader, i, i)

	case MReader:
		mReader := MReaderToReader(v)
		defaultReadFunc = ReadFuncToAck(mReader.ReadFunc)
		i = NewReadWriteCloser(mReader, i, i)

	}
	//默认buf大小,可自定义缓存大小
	this.buf = bufio.NewReaderSize(i, DefaultBufferSize+1)

	//设置默认的Option
	this.SetKey(key) //设置唯一标识
	this.Debug()     //设置打印日志

	//设置默认事件
	this.SetConnectWithNil().SetConnectWithLog()
	this.SetReadAckFunc(defaultReadFunc)
	this.SetDealWithNil().SetDealWithDefault()
	this.SetWriteWithNil().SetWriteWithLog()
	this.SetCloseWithLog()

	//设置用户的Option
	this.SetOptions(options...)

	return this
}

// GetKey 获取标识
func (this *Client) GetKey() string {
	return this.Key
}

// SetKey 设置标识
func (this *Client) SetKey(key string) *Client {
	if key == this.Key {
		return this
	}
	oldKey := this.Key
	this.Key = key
	for _, v := range this.keyChangeFunc {
		v(this, oldKey)
	}
	return this
}

// ReadWriteCloser 读写接口,实例,传入的原始参数
func (this *Client) ReadWriteCloser() io.ReadWriteCloser {
	return this.i
}

// ID 获取唯一标识,不可改,溯源,不用因为改了key而出现数据错误的bug
func (this *Client) ID() string {
	if this.pointer == "" {
		pointer := fmt.Sprintf("%p", this)
		this.pointer = pointer
	}
	return this.pointer
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
