package io

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"time"
)

//================================Nature================================

func (this *Client) WriteTime() time.Time {
	return this.writeTime
}

func (this *Client) WriteCount() int64 {
	return this.writeBytes
}

func (this *Client) WriteNumber() int64 {
	return this.writeNumber
}

//================================Write================================

// Write 写入字节,实现io.Writer
func (this *Client) Write(p []byte) (n int, err error) {
	defer func() {
		err = dealErr(err)
	}()
	//记录错误,用于队列的判断
	if this.writeFunc != nil {
		p, err = this.writeFunc(p)
		if err != nil {
			return 0, err
		}
	}
	//打印实际发送的数据,方便调试
	this.logger.Writeln("["+this.GetKey()+"] ", p)
	//写入数据
	n, err = this.i.Write(p)
	if err != nil {
		return 0, err
	}
	this.writeTime = time.Now()
	this.writeBytes += int64(n)
	this.writeNumber++
	select {
	case this.timeoutReset <- struct{}{}:
	default:
	}
	return
}

// WriteBytes 写入字节,实现bytesWriter
func (this *Client) WriteBytes(p []byte) (int, error) {
	return this.Write(p)
}

// WriteString 写入字符串,实现io.StringWriter
func (this *Client) WriteString(s string) (int, error) {
	return this.Write([]byte(s))
}

// WriteHEX 写入16进制数据
func (this *Client) WriteHEX(s string) (int, error) {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return this.Write(bytes)
}

// WriteBase64 写入base64数据
func (this *Client) WriteBase64(s string) (int, error) {
	bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return this.Write(bytes)
}

// WriteAny 写入任意数据,根据conv转成字节
func (this *Client) WriteAny(any interface{}) (int, error) {
	return this.Write(conv.Bytes(any))
}

// WriteSplit 写入字节,分片写入,例如udp需要写入字节小于(1500-20-8=1472)
func (this *Client) WriteSplit(p []byte, length int) (int, error) {
	if length <= 0 {
		return this.Write(p)
	}
	for len(p) > 0 {
		var data []byte
		if len(p) >= length {
			data, p = p[:length], p[length:]
		} else {
			data, p = p, p[:0]
		}
		_, err := this.Write(data)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// WriteReader io.Reader
func (this *Client) WriteReader(reader Reader) (int64, error) {
	return Copy(this, reader)
}

// WriteFrom io.Reader
func (this *Client) WriteFrom(reader Reader) (int64, error) {
	return Copy(this, reader)
}

// WriteFromChan 监听通道并写入
func (this *Client) WriteFromChan(c chan interface{}) (int64, error) {
	var total int64
	for data := range c {
		n, err := this.Write(conv.Bytes(data))
		if err != nil {
			return 0, err
		}
		total += int64(n)
	}
	return total, nil
}

func (this *Client) WriteQueue(p []byte) (int, error) {
	if this.closeErr != nil {
		return 0, this.closeErr
	}
	this.initQueue()
	select {
	case <-this.Done():
		return 0, this.Err()
	case this.writeQueue <- p:
		return len(p), nil
	}
}

func (this *Client) WriteQueueTry(p []byte) (int, error) {
	this.initQueue()
	select {
	case <-this.Done():
		return 0, this.Err()
	case this.writeQueue <- p:
		return len(p), nil
	default:
		return 0, errors.New("队列已满")
	}
}

func (this *Client) WriteQueueTimeout(p []byte, timeout time.Duration) (int, error) {
	this.initQueue()
	select {
	case <-this.Done():
		return 0, this.Err()
	case this.writeQueue <- p:
		return len(p), nil
	case <-time.After(timeout):
		return 0, ErrWithWriteTimeout
	}
}

func (this *Client) initQueue() {
	this.writeQueueOnce.Do(func() {
		this.writeQueue = make(chan []byte, DefaultChannelSize)
		go func() {
			for p := range this.writeQueue {
				if _, err := this.Write(p); err != nil {
					return
				}
			}
		}()
	})
}

// GoTimerWriter 协程,定时写入数据,生命周期(一次链接,单次连接断开)
func (this *Client) GoTimerWriter(interval time.Duration, write func(w *Client) (int, error)) {
	go this.Timer(interval, func() error {
		_, err := write(this)
		return err
	})
}

// GoTimerWriteBytes 协程,定时写入字节数据
func (this *Client) GoTimerWriteBytes(interval time.Duration, p []byte) {
	this.GoTimerWriter(interval, func(w *Client) (int, error) {
		return w.Write(p)
	})
}

// GoTimerWriteString 协程,定时写入字符数据
func (this *Client) GoTimerWriteString(interval time.Duration, s string) {
	this.GoTimerWriter(interval, func(w *Client) (int, error) {
		return w.WriteString(s)
	})
}

func (this *Client) GoTimerWriteHEX(interval time.Duration, s string) {
	this.GoTimerWriter(interval, func(w *Client) (int, error) {
		return w.WriteHEX(s)
	})
}

// SetKeepAlive 设置连接保持,另外起了携程,服务器不需要,客户端再起一个也没啥问题
// TCP keepalive定义于RFC 1122，但并不是TCP规范中的一部分,默认必需是关闭,连接方不一定支持
func (this *Client) SetKeepAlive(t time.Duration, keeps ...[]byte) {
	this.GoTimerWriter(t, func(c *Client) (int, error) {
		keep := conv.GetDefaultBytes([]byte(Ping), keeps...)
		return c.Write(keep)
	})
}

//================================WriteFunc================================

// SetWriteFunc 设置写入函数,封装数据包,same SetWriteBeforeFunc
func (this *Client) SetWriteFunc(fn func(p []byte) ([]byte, error)) *Client {
	this.writeFunc = fn
	return this
}

// SetWriteWithPkg 默认写入函数
func (this *Client) SetWriteWithPkg() *Client {
	return this.SetWriteFunc(WriteWithPkg)
}

// SetWriteWithNil 取消写入函数
func (this *Client) SetWriteWithNil() *Client {
	return this.SetWriteFunc(nil)
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
