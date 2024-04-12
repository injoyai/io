package io

import (
	"bufio"
	"context"
	"errors"
	"github.com/injoyai/base/chans"
	"github.com/injoyai/io/buf"
	"io"
	"net"
	"sync/atomic"
	"time"
)

//================================Nature================================

// ReadTime 最后读取到数据的时间
func (this *Client) ReadTime() time.Time {
	return this.readTime
}

// ReadCount 读取到数据的数量
func (this *Client) ReadCount() int64 {
	return this.readBytes
}

// Buffer 极大的增加读取速度
func (this *Client) Buffer() *bufio.Reader {
	return this.buf
}

// Read io.reader
func (this *Client) Read(p []byte) (int, error) {
	return this.Buffer().Read(p)
}

// ReadByte 读取一字节
func (this *Client) ReadByte() (byte, error) {
	return this.Buffer().ReadByte()
}

// Read1KB 读取1kb数据
func (this *Client) Read1KB() ([]byte, error) {
	return buf.Read1KB(this.Buffer())
}

// ReadMessage 实现MessageReader接口
func (this *Client) ReadMessage() ([]byte, error) {
	if this.readFunc == nil {
		return nil, ErrInvalidReadFunc
	}
	return this.readFunc(this.Buffer())
}

// ReadLast 读取最新的数据
func (this *Client) ReadLast(timeout time.Duration) (response []byte, err error) {
	if timeout <= 0 {
		response = <-this.readChan
	} else {
		select {
		case response = <-this.readChan:
		case <-time.After(timeout):
			err = ErrWithTimeout
		}
	}
	return
}

// WriteTo 写入io.Writer
func (this *Client) WriteTo(writer Writer) (int64, error) {
	return Copy(writer, this)
}

// SetReadIntervalTimeout 设置读取间隔超时时间,需要在Run之前设置
func (this *Client) SetReadIntervalTimeout(timeout time.Duration) *Client {
	this.timeout = timeout
	return this
}

//================================DealFunc================================

// SetDealFunc 设置处理数据函数,默认响应ping>pong,忽略pong
func (this *Client) SetDealFunc(fn func(c *Client, msg Message)) *Client {
	this.dealFunc = append(this.dealFunc, fn)
	return this
}

// SetDealWithNil 不设置数据处理函数,删除之前设置的处理数据函数
func (this *Client) SetDealWithNil() *Client {
	this.dealFunc = nil
	return this
}

// SetDealFuncOnly 取消之前设置的处理函数,并设置新的函数
func (this *Client) SetDealFuncOnly(fn func(c *Client, msg Message)) *Client {
	this.SetDealWithNil()
	return this.SetDealFunc(fn)
}

// SetDealWithDefault 设置默认处理数据函数,打印需要处理的数据,和处理数据ping,pong
func (this *Client) SetDealWithDefault() *Client {
	return this.SetDealFuncOnly(func(c *Client, msg Message) {
		this.logger.Readln("["+c.GetKey()+"] ", msg)
		//先判断长度,减少字节转字符的内存分配,最好用指针的方式(直接用字节的指针)
		if msg.Len() == len(Ping) || msg.Len() == len(Pong) {
			switch msg.String() {
			case Ping:
				this.WriteString(Pong)
				return
			case Pong:
				return
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

//================================ReadFunc================================

// SetReadFunc 设置读取函数
// 后台循环执行(在使用Run之后),从字节留中间截取符合协议的数据,默认最大读取1字节数据
// 例modbus,读取crc校验正确的数据 ,如下图截取,后续数据等待下次截取
// 01 03 00 01 00 02 xx xx | 01 03 00 01 00 02 xx xx | 01 03 00 01 00 02 xx xx
// 截取的数据下一步会在DealFunc中执行
func (this *Client) SetReadFunc(fn func(r *bufio.Reader) ([]byte, error)) *Client {
	this.readFunc = func(reader *bufio.Reader) (bs []byte, err error) {

		if fn == nil {
			fn = buf.Read1KB
		}

		//执行用户设置的函数
		bs, err = fn(reader)
		if err != nil {
			return nil, err
		}

		if len(bs) > 0 {
			//设置最后读取有效数据时间
			this.readTime = time.Now()
			this.readBytes += int64(len(bs))

			//尝试加入通道,如果设置了监听,则有效
			select {
			case this.readChan <- bs:
			default:
			}

			//尝试加入通道,超时定时器重置
			select {
			case this.timeoutReset <- struct{}{}:
			default:
			}
		}

		return bs, nil

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

//================================RunTime================================

// Running 是否在运行
func (this *Client) Running() bool {
	return atomic.LoadUint32(&this.running) == 1
}

// Run 开始运行数据读取
// 2个操作,持续读取数据并处理,监测超时(如果设置了超时,不推荐服务端使用)
func (this *Client) Run() error {

	//原子操作,防止重复执行
	if atomic.SwapUint32(&this.running, 1) == 1 {
		return nil
	}

	//todo is a good idea ?
	if this.timeout > 0 {
		go func(ctx context.Context) {
			timer := time.NewTimer(this.timeout)
			defer timer.Stop()
			for {
				timer.Reset(this.timeout)
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					_ = this.CloseWithErr(ErrWithReadTimeout)
					return
				case <-this.timeoutReset:
				}
			}
		}(this.Ctx())
	}

	//开始循环读取数据,处理数据
	return this.For(func(ctx context.Context) (err error) {

		//读取数据
		bs, err := this.ReadMessage()
		if err != nil || len(bs) == 0 {
			return err
		}

		//处理数据
		for _, dealFunc := range this.dealFunc {
			dealFunc(this, bs)
		}

		return nil
	})

}
