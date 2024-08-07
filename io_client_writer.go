package io

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/injoyai/conv"
	"time"
)

//================================Write================================

// Write 写入字节,实现io.Writer
func (this *Client) Write(p []byte) (n int, err error) {
	defer func() {
		err = dealErr(err)
		for _, v := range this.writeResultFunc {
			v(this, err)
		}
	}()

	//执行写入函数,处理写入的数据,进行封装或者打印等操作
	for _, f := range this.writeFunc {
		p, err = f(p)
		if err != nil {
			return 0, err
		}
	}

	//写入数据
	n, err = this.i.Write(p)
	if err != nil {
		return 0, err
	}
	this.WriteTime = time.Now()
	this.WriteCount += uint64(n)
	this.WriteNumber++
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

// WriteChan 监听通道并写入
func (this *Client) WriteChan(c chan interface{}) (int64, error) {
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

// SetKeepAlive 设置连接保持,另外起了携程,服务端不需要,客户端再起一个也没啥问题
// TCP keepalive定义于RFC 1122，但并不是TCP规范中的一部分,默认必需是关闭,连接方不一定支持
func (this *Client) SetKeepAlive(t time.Duration, keeps ...[]byte) {
	this.GoTimerWriter(t, func(c *Client) (int, error) {
		keep := conv.GetDefaultBytes([]byte(Ping), keeps...)
		return c.Write(keep)
	})
}
