package io

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"sync/atomic"

	"io"
	"log"
	"sync"
	"time"
)

func RedialClient(connect func() (ReadWriteCloser, error), fn ...func(c *Client)) *Client {
	return MustClient(connect).Redial(fn...)
}

func MustClient(connect func() (ReadWriteCloser, error)) *Client {
	t := time.Second
	for {
		client, err := connect()
		if err == nil {
			log.Printf("[信息] 连接服务成功...\n")
			return NewClient(client).SetRedialFunc(connect)
		}
		if t < time.Second*32 {
			t = 2 * t
		}
		log.Println("[错误]", dealErr(err), ",等待", t, "重试")
		time.Sleep(t)
	}
}

func NewDial(dial func() (ReadWriteCloser, error)) (*Client, error) {
	return NewDialWithContext(context.Background(), dial)
}

func NewDialWithContext(ctx context.Context, dial func() (ReadWriteCloser, error)) (*Client, error) {
	c, err := dial()
	if err != nil {
		return nil, err
	}
	return NewClientWithContext(ctx, c).SetRedialFunc(dial), nil
}

func NewClient(i ReadWriteCloser) *Client {
	return NewClientWithContext(context.Background(), i)
}

func NewClientWithContext(ctx context.Context, i ReadWriteCloser) *Client {
	if c, ok := i.(*Client); ok && c != nil {
		return c
	}
	c := &Client{
		buf:        bufio.NewReader(i),
		key:        fmt.Sprintf("%p", i),
		i:          i,
		tag:        maps.NewSafe(),
		printFunc:  PrintWithHEX,
		closeFunc:  nil,
		redialFunc: nil,
		readFunc:   ReadWithAll,
		dealFunc:   nil,
		writeFunc:  nil,
		writeChan:  make(chan []byte, 1),
		readChan:   make(chan []byte),

		timer: time.NewTimer(0),
		//timerRead:    time.NewTimer(0),
		//timerWrite:   time.NewTimer(0),
		timerKeep: time.NewTimer(0),
		timeout:   0,
		//timeoutRead:  0,
		//timeoutWrite: 0,
		closeErr: ErrWithContext,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)
	if c.timeout <= 0 {
		<-c.timer.C
	}
	if c.keepAlive <= 0 {
		<-c.timerKeep.C
	}
	return c
}

type Client struct {
	buf *bufio.Reader   //buff
	key string          //唯一标识
	i   ReadWriteCloser //接口
	mu  sync.Mutex      //锁
	tag *maps.Safe      //标签

	ClientReader
	ClientWriter

	closeFunc  func(*Message)                  //结束信号,链接关闭触发事件
	redialFunc func() (ReadWriteCloser, error) //重连函数

	timer     *time.Timer   //超时定时器,时间范围内没有发送数据或者接收数据,则断开链接
	timeout   time.Duration //超时时间
	timerKeep *time.Timer   //正常通讯不发送心跳
	keepAlive time.Duration //保持连接

	createTime time.Time //创建时间,链接成功时间
	lastTime   time.Time //最后通讯时间(读取或者发送)

	debug    bool          //调试,是否打印数据
	closeAll bool          //全部关闭,主要是为了,为了区别主动关闭和错误关闭
	closeErr error         //错误信息
	running  *atomic.Value //是否已经执行,防止重复执行

	cancel context.CancelFunc //上下文关闭
	ctx    context.Context    //上下文
}

// ReadWriteCloser 读写接口
func (this *Client) ReadWriteCloser() io.ReadWriteCloser {
	return this.i
}

// Interface 读写接口
func (this *Client) Interface() io.ReadWriteCloser {
	return this.i
}

// Buffer 极大的增加读取速度
func (this *Client) Buffer() *bufio.Reader {
	return this.buf
}

// Tag 自定义信息
func (this *Client) Tag() *maps.Safe {
	return this.tag
}

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
	return this
}

// GetKey 获取唯一标识
func (this *Client) GetKey() string {
	return this.key
}

// Ctx 上下文
func (this *Client) Ctx() context.Context {
	return this.ctx
}

// Running 是否运行中
func (this *Client) Running() bool {
	v := this.running.Load()
	return v != nil && v.(bool)
}

// Closed 是否断开连接
func (this *Client) Closed() bool {
	select {
	case <-this.ctx.Done():
		return true
	default:
		return false
	}
}

// Err 错误信息,默认有个错误,如果连接正常,错误为默认,则返回nil
func (this *Client) Err() error {
	if this.closeErr == ErrWithContext && !this.Closed() {
		return nil
	}
	return this.closeErr
}

// Run 运行,会阻塞
func (this *Client) Run() error {
	if this.Running() {
		select {
		case <-this.ctx.Done():
			return this.closeErr
		}
	}
	this.running = true
	return this.run()
}

// Debug 调试模式,打印日志
func (this *Client) Debug(b ...bool) *Client {
	this.debug = !(len(b) > 0 && !b[0])
	return this
}

// GetDebug 获取调试模式
func (this *Client) GetDebug() bool {
	return this.debug
}

// Close 主动关闭连接,无法触发重试机制
func (this *Client) Close() error {
	this.closeAll = true
	return this.close(errors.New("主动关闭"))
}

// close 内部关闭连接,触发重试机制
func (this *Client) close(err error) error {
	this.mu.Lock()
	if err != nil && (this.closeErr == ErrWithContext || this.closeErr == nil) {
		//重置关闭原因
		this.closeErr = dealErr(err)
		//关闭上下文
		this.cancel()
		//关闭连接
		this.i.Close()
		msg := NewMessage([]byte(this.closeErr.Error()))
		//触发关闭函数,需要defer,要在Unlock之后,防止锁无法释放,
		//例如设置了重连(递归),其它加锁位置将一直等待
		if this.closeFunc != nil {
			defer this.closeFunc(msg)
		}
		//打印日志
		if this.debug && this.printFunc != nil {
			this.printFunc("关闭", msg)
		}
	}
	//需要在this.closeFunc之前,
	//可能会重试(递归),就无法释放锁
	this.mu.Unlock()
	return nil
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

// ReadChan 数据读取通道
func (this *Client) ReadChan() <-chan []byte {
	return this.readChan
}

// ReadLast 读取最新的数据
func (this *Client) ReadLast(timeout time.Duration) (response []byte, err error) {
	if timeout <= 0 {
		select {
		case <-this.ctx.Done():
			err = this.closeErr
		case response = <-this.readChan:
		}
	} else {
		t := time.NewTimer(timeout)
		select {
		case <-this.ctx.Done():
			err = this.closeErr
		case response = <-this.readChan:
		case <-t.C:
			err = ErrWithTimeout
		}
	}
	return
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

// Done 结束,关闭信号,一定有错误
func (this *Client) Done() <-chan struct{} {
	return this.ctx.Done()
}

// SetKeepAlive 设置连接保持,另外起了携程,服务器不需要,客户端再起一个也没啥问题
// TCP keepalive定义于RFC 1122，但并不是TCP规范中的一部分,默认必需是关闭,连接方不一定支持
func (this *Client) SetKeepAlive(t time.Duration, keeps ...[]byte) *Client {
	keep := conv.GetDefaultBytes([]byte("ping"), keeps...)
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
		}(this.ctx)
	}
	return this
}

// SetReadWriteWithStartEnd 设置读取写入数据根据包头包尾
func (this *Client) SetReadWriteWithStartEnd(packageStart, packageEnd []byte) *Client {
	this.ClientWriter.SetWriteWithStartEnd(packageStart, packageEnd)
	this.ClientReader.SetReadWithStartEnd(packageStart, packageEnd)
	return this
}

// SetCloseFunc 设置关闭函数,重试会设置此函数,另外设置了此函数,重试将会失效
func (this *Client) SetCloseFunc(fn func(msg *Message)) *Client {
	this.closeFunc = fn
	return this
}

// SetCloseWithNil 不设置关闭函数
func (this *Client) SetCloseWithNil() *Client {
	return this.SetCloseFunc(nil)
}

// SetRedialFunc 设置重连函数
func (this *Client) SetRedialFunc(fn func() (ReadWriteCloser, error)) *Client {
	this.redialFunc = fn
	return this
}

// Redial 重新链接,重试
func (this *Client) Redial(fn ...func(c *Client)) *Client {
	this.SetCloseFunc(func(msg *Message) {
		if this.closeAll {
			log.Printf("[信息][%s] 连接断开(%s)", msg.Key(), msg)
			return
		}
		if this.redialFunc == nil {
			log.Printf("[信息][%s] 连接断开(%s),未设置重连函数", msg.Key(), msg)
			return
		}
		log.Printf("[信息][%s] 连接断开(%s),尝试重新连接\n", msg.Key(), msg)
		time.Sleep(time.Second)
		//目前重试只保留了key,其他配置重置成默认值
		key := this.GetKey()
		*this = *MustClient(this.redialFunc)
		this.SetKey(key).Redial(fn...)
	})
	for _, v := range fn {
		v(this)
	}
	if !this.Running() {
		go this.Run()
	}
	return this
}

// SetRunType 设置运行模式
func (this *Client) SetRunType(runType Type) *Client {
	this.runType = runType
	return this
}

func (this *Client) run() (err error) {
	defer this.close(err)
	switch this.runType {
	case TypeWrite:
		return this.runWrite(this.ctx)
	case TypeRead:
		return this.runRead(this.ctx)
	case TypeCallback:
		return this.runCallback(this.ctx)
	default:
		return this.runNormal(this.ctx)
	}
}

func (this *Client) runWrite(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return this.closeErr
		case bs := <-this.writeChan:
			this.close(this.write(bs))
		}
	}
}

func (this *Client) runRead(ctx context.Context) error {
	if this.readFunc == nil {
		return ErrInvalidReadFunc
	}
	for {
		select {
		case <-ctx.Done():
			return this.closeErr
		default:
			this.close(this.read())
		}
	}
}

func (this *Client) runCallback(ctx context.Context) error {
	if this.readFunc == nil {
		return ErrInvalidReadFunc
	}
	for {
		select {
		case <-ctx.Done():
			return this.closeErr
		case bs := <-this.writeChan:
			this.close(this.write(bs))
			this.close(this.read())
		case <-this.timer.C:
			this.close(errors.New("超时"))
		}
	}
}

func (this *Client) runTimer(ctx context.Context) error {
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return this.closeErr
		case <-this.timer.C:
			if this.timeout > 0 {
				this.close(errors.New("超时"))
			}
		}
	}
}

func (this *Client) runNormal(ctx context.Context) error {
	defer this.timer.Stop()
	go func(ctx context.Context) {
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return
			case <-this.timer.C:
				if this.timeout > 0 {
					this.close(ErrWithTimeout)
				}
			}
		}
	}(ctx)
	go this.runRead(ctx)
	return this.runWrite(ctx)
}

// 写数据
func (this *Client) write(bs []byte) error {
	if len(bs) > 0 {
		if this.writeFunc != nil {
			bs = this.writeFunc(bs)
		}
		this.lastWriteTime = time.Now()
		this.lastTime = time.Now()
		if this.debug && this.printFunc != nil {
			this.printFunc("发送", newMessage(this, bs))
		}
		if this.timeout > 0 {
			this.timer.Reset(this.timeout)
		}
		if this.keepAlive > 0 {
			this.timerKeep.Reset(this.keepAlive)
		}
		_, err := this.i.Write(bs)
		return err
	}
	return nil
}

// 读数据,会默认使用默认读取函数,不需要读取数据,需要在外面处理
func (this *Client) read() (err error) {
	defer func() {
		//捕捉错误,有出现bufio越界问题
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	if this.readFunc == nil {
		this.readFunc = ReadWithDefault
	}
	bs, err := this.readFunc(this)
	if err != nil && err != io.EOF {
		return err
	} else if err != nil {
		return ErrRemoteClose
	} else if len(bs) > 0 {
		//设置最后读取有效数据时间
		this.lastReadTime = time.Now()
		//设置最后有效通讯时间
		this.lastTime = time.Now()
		if this.debug && this.printFunc != nil {
			//开启debug,并设置了打印函数,则执行
			this.printFunc("接收", newMessage(this, bs))
		}
		select {
		//尝试加入队列,不消费的话加入不了,队列长度1
		case this.readChan <- bs:
		default:
		}
		if this.dealFunc != nil {
			//执行处理数据函数
			this.dealFunc(newMessage(this, bs))
		}
		if this.timeout > 0 {
			//重置超时时间,如果设置
			this.timer.Reset(this.timeout)
		}
	}
	return
}
