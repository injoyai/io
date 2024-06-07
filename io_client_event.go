package io

import (
	"bufio"
	"context"
	"errors"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io/buf"
	"io"
	"net"
	"time"
)

// OnKeyChange key变化事件
func (this *Client) OnKeyChange(f func(c *Client, oldKey string)) *Client {
	return this.SetKeyChangeFunc(f)
}

// OnConnect 连接成功事件
func (this *Client) OnConnect(f func(c *Client) error) *Client {
	return this.SetConnectFunc(f)
}

// OnDisconnect 断开连接事件
func (this *Client) OnDisconnect(f func(ctx context.Context, c *Client, err error)) *Client {
	return this.SetCloseFunc(f)
}

// OnReadBuffer 接收字节流事件,返回的数据会触发OnReadMessage
func (this *Client) OnReadBuffer(f func(r *bufio.Reader) ([]byte, error)) *Client {
	return this.SetReadFunc(f)
}

// OnReadMessage 读取到消息事件
func (this *Client) OnReadMessage(f func(c *Client, msg Message) (ack bool)) *Client {
	return this.SetDealAckFunc(f)
}

// OnWriteMessage 写入消息事件
func (this *Client) OnWriteMessage(f func(p []byte) ([]byte, error)) *Client {
	return this.SetWriteFunc(f)
}

// OnWriteResult 写入结果事件
func (this *Client) OnWriteResult(f func(c *Client, err error)) *Client {
	return this.SetWriteResultFunc(f)
}

//================================KeyChangeFunc================================

// SetKeyChangeFunc 设置key变化事件
func (this *Client) SetKeyChangeFunc(f func(c *Client, oldKey string)) *Client {
	this.keyChangeFunc = append(this.keyChangeFunc, f)
	return this
}

// SetKeyChangeWithNil 设置key变化事件为空
func (this *Client) SetKeyChangeWithNil() *Client {
	this.keyChangeFunc = nil
	return this
}

//================================ConnectFunc================================

// SetConnectFunc 连接成功事件
func (this *Client) SetConnectFunc(f func(c *Client) error) *Client {
	this.connectFunc = append(this.connectFunc, f)
	return this
}

// SetConnectWithLog 设置连接成功打印日志
func (this *Client) SetConnectWithLog() *Client {
	return this.SetConnectFunc(func(c *Client) error {
		this.Infof("[%s] 连接服务端成功...\n", this.GetKey())
		return nil
	})
}

func (this *Client) SetConnectWithNil() *Client {
	this.connectFunc = nil
	return this
}

//================================ReadFunc================================

// SetReadFunc 设置读取函数
// 后台循环执行(在使用Run之后),从字节留中间截取符合协议的数据,默认最大读取1字节数据
// 例modbus,读取crc校验正确的数据 ,如下图截取,后续数据等待下次截取
// 01 03 00 01 00 02 xx xx | 01 03 00 01 00 02 xx xx | 01 03 00 01 00 02 xx xx
// 截取的数据下一步会在DealFunc中执行
func (this *Client) SetReadFunc(fn func(r *bufio.Reader) ([]byte, error)) *Client {
	return this.SetReadAckFunc(ReadFuncToAck(fn))
}

func (this *Client) SetReadAckFunc(fn func(r *bufio.Reader) (Acker, error)) *Client {
	this.readFunc = func(reader *bufio.Reader) (Acker, error) {

		if fn == nil {
			fn = ReadFuncToAck(buf.Read1KB)
		}

		//执行用户设置的函数
		ack, err := fn(reader)
		if err != nil {
			return nil, err
		}

		if bs := ack.Payload(); len(bs) > 0 {
			//设置最后读取有效数据时间
			this.ReadTime = time.Now()
			this.ReadCount += uint64(len(bs))

			//尝试加入通道,如果设置了监听,则有效
			select {
			case this.latestChan <- bs:
			default:
			}

			//尝试加入通道,超时定时器重置
			select {
			case this.timeoutReset <- struct{}{}:
			default:
			}
		}

		return ack, nil

	}
	return this
}

// SetReadWithPkg 使用默认读包方式
func (this *Client) SetReadWithPkg() *Client {
	return this.SetReadFunc(ReadWithPkg)
}

// SetReadWith1KB 每次读取1字节
func (this *Client) SetReadWith1KB() {
	this.SetReadFunc(buf.Read1KB)
}

// SetReadWithKB 读取固定字节长度
func (this *Client) SetReadWithKB(n uint) *Client {
	return this.SetReadFunc(func(buf *bufio.Reader) ([]byte, error) {
		bytes := make([]byte, n<<10)
		length, err := buf.Read(bytes)
		return bytes[:length], err
	})
}

// SetReadWithStartEnd 设置根据包头包尾读取数据
func (this *Client) SetReadWithStartEnd(packageStart, packageEnd []byte) *Client {
	return this.SetReadFunc(buf.NewReadWithStartEnd(packageStart, packageEnd))
}

// SetReadWithWriter same io.Copy 注意不能设置读取超时
func (this *Client) SetReadWithWriter(writer io.Writer) *Client {
	return this.SetReadFunc(buf.NewReadWithWriter(writer))
}

// SetReadWithTimeout 根据超时时间读取数据(需要及时读取,避免阻塞产生粘包),
// 需要支持SetReadDeadline(t time.Time) error接口
func (this *Client) SetReadWithTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return errors.New("无效超时时间")
	}
	i, ok := this.i.(net.Conn)
	if !ok {
		return errors.New("无法设置超时时间")
	}
	buff := make([]byte, 1024)
	this.SetReadFunc(func(r *bufio.Reader) ([]byte, error) {
		result := []byte(nil)
		for {
			n, err := r.Read(buff)
			result = append(result, buff[:n]...)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					return result, nil
				}
				return nil, err
			}
			if err := i.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				return nil, err
			}
		}
	})
	return nil
}

// Bridge 桥接模式,等同SetReadWithWriter
// 把读取到的数据全部写入到io.Writer
func (this *Client) Bridge(w io.Writer) *Client {
	return this.SetReadFunc(buf.NewReadWithWriter(w))
}

// SetReadWithLenFrame 根据动态长度读取数据
func (this *Client) SetReadWithLenFrame(f *buf.LenFrame) *Client {
	return this.SetReadFunc(buf.NewReadWithLen(f))
}

// SetReadWithFrame 适配预大部分读取
func (this *Client) SetReadWithFrame(f *buf.Frame) *Client {
	return this.SetReadFunc(buf.NewReadWithFrame(f))
}

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

// Swap IO数据交换
func (this *Client) Swap(i ReadWriteCloser) {
	this.SwapClient(NewClient(i))
}

// SwapClient IO数据交换
func (this *Client) SwapClient(c *Client) {
	SwapClient(this, c)
}

//================================DealFunc================================

// SetDealFunc 设置处理数据函数,默认响应ping>pong,忽略pong
func (this *Client) SetDealFunc(fn func(c *Client, msg Message)) *Client {
	return this.SetDealAckFunc(func(c *Client, msg Message) (ack bool) {
		if fn != nil {
			fn(c, msg)
		}
		return true
	})
}

func (this *Client) SetDealAckFunc(fn func(c *Client, msg Message) (ack bool)) *Client {
	this.dealFunc = append(this.dealFunc, fn)
	return this
}

// SetDealWithNil 不设置数据处理函数,删除之前设置的处理数据函数
func (this *Client) SetDealWithNil() *Client {
	this.dealFunc = nil
	return this
}

// SetDealWithLog 设置处理函数为打印日志
func (this *Client) SetDealWithLog() *Client {
	return this.SetDealFunc(func(c *Client, msg Message) {
		//打印实际接收的数据,方便调试
		this.logger.Readln("["+c.GetKey()+"] ", msg)
	})
}

// SetDealWithDefault 设置默认处理数据函数,打印需要处理的数据,和处理数据ping,pong
func (this *Client) SetDealWithDefault() *Client {
	return this.SetDealWithLog().SetDealFunc(func(c *Client, msg Message) {
		//先判断长度,减少字节转字符的内存分配,最好用指针的方式(直接用字节的指针)
		if msg.Len() == len(Ping) || msg.Len() == len(Pong) {
			switch msg.String() {
			case Ping:
				this.WriteString(Pong)
			case Pong:
			}
		}
	})
}

// SetDealWithWriter 设置数据处理到io.Writer
func (this *Client) SetDealWithWriter(writer Writer) *Client {
	return this.SetDealFunc(func(c *Client, msg Message) {
		if _, err := writer.Write(msg); err != nil {
			c.Close()
		}
	})
}

// SetDealWithChan 设置数据处理到chan
func (this *Client) SetDealWithChan(ch chan Message) *Client {
	return this.SetDealFunc(func(c *Client, msg Message) {
		ch <- msg
	})
}

// SetDealWithQueue 设置协程队列处理数据
// @num 协程数量
// @fn 处理函数
func (this *Client) SetDealWithQueue(num int, fn func(msg Message)) *Client {
	queue := chans.NewEntity(num).SetHandler(func(ctx context.Context, no, count int, data interface{}) {
		fn(data.(Message))
	})
	this.SetDealFunc(func(c *Client, msg Message) { queue.Do(msg) })
	return this
}

//================================WriteFunc================================

func (this *Client) SetWriteResultFunc(fn func(c *Client, err error)) *Client {
	this.writeResultFunc = append(this.writeResultFunc, fn)
	return this
}

func (this *Client) SetWriteResultWithNil() *Client {
	this.writeResultFunc = nil
	return this
}

// SetWriteFunc 设置写入函数,封装数据包,same SetWriteBeforeFunc
func (this *Client) SetWriteFunc(fn func(p []byte) ([]byte, error)) *Client {
	this.writeFunc = append(this.writeFunc, fn)
	return this
}

// SetWriteWithNil 取消写入函数
func (this *Client) SetWriteWithNil() *Client {
	this.writeFunc = nil
	return this
}

func (this *Client) SetWriteWithLog() *Client {
	return this.SetWriteFunc(func(p []byte) ([]byte, error) {
		//打印实际发送的数据,方便调试
		this.logger.Writeln("["+this.GetKey()+"] ", p)
		return p, nil
	})
}

// SetWriteWithPkg 默认写入函数
func (this *Client) SetWriteWithPkg() *Client {
	return this.SetWriteFunc(WriteWithPkg)
}

// SetWriteWithStartEnd 设置写入函数,增加头尾
func (this *Client) SetWriteWithStartEnd(start, end []byte) *Client {
	return this.SetWriteFunc(buf.NewWriteWithStartEnd(start, end))
}

func (this *Client) SetWriteWithQueue(cap int) *Client {
	//删除老的队列
	if i, ok := this.Tag().Get("_queue"); ok {
		close(i.(chan []byte))
	}
	queue := make(chan []byte, cap)
	go func(w Writer, queue chan []byte) {
		for p := range queue {
			if _, err := this.Write(p); err != nil {
				return
			}
		}
	}(this, queue)
	this.SetWriteFunc(func(p []byte) ([]byte, error) {
		select {
		case <-this.Done():
			return nil, this.Err()
		case <-queue:
			return p, nil
		}
	})
	return this
}

//================================CloseFunc================================

// SetCloseFunc 设置关闭函数
func (this *Client) SetCloseFunc(fn func(ctx context.Context, c *Client, err error)) *Client {
	this.closeFunc = fn
	return this
}

// SetCloseWithLog 设置关闭时打印日志
func (this *Client) SetCloseWithLog() {
	this.SetCloseFunc(func(ctx context.Context, c *Client, err error) {
		this.logger.Errorf("[%s] 断开连接: %v\n", this.GetKey(), err)
	})
}

// SetCloseWithNil 设置无关闭函数
func (this *Client) SetCloseWithNil() *Client {
	this.closeFunc = nil
	return this
}

func (this *Client) SetCloseWithRedial(op ...OptionClient) {
	this.SetCloseFunc(func(ctx context.Context, c *Client, err error) {
		c.MustDial(this.ctxParent, op...)
	})
}

// SetCloseWithCloser 设置关闭函数关闭Closer
func (this *Client) SetCloseWithCloser(closer Closer) *Client {
	return this.SetCloseFunc(func(ctx context.Context, c *Client, err error) {
		closer.Close()
	})
}

//================================Event================================

type Event struct {
	OnConnect      func(c *Client) error
	OnReadBuffer   func(buf *bufio.Reader) ([]byte, error)
	OnDealMessage  func(c *Client, msg Message)
	OnWriteMessage func(bs []byte) ([]byte, error)
	OnDisconnect   func(ctx context.Context, c *Client, err error)
	Options        []OptionClient
}

type Option = OptionClient
