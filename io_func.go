package io

import (
	"bufio"
	"github.com/injoyai/io/buf"
	"io"
)

// CopyFunc 复制数据,每次固定4KB,并提供函数监听
func CopyFunc(w Writer, r Reader, fn func(buf []byte)) (int, error) {
	return CopyNFunc(w, r, DefaultBufferSize, fn)
}

// CopyNFunc 复制数据,每次固定大小,并提供函数监听
func CopyNFunc(w Writer, r Reader, n int64, fn func(buf []byte)) (int, error) {
	buff := bufio.NewReader(r)
	length := 0
	for {
		buf := make([]byte, n)
		n, err := buff.Read(buf)
		if err != nil && err != io.EOF {
			return length, err
		}
		length += n
		if _, err := w.Write(buf[:n]); err != nil {
			return length, err
		}
		if fn != nil {
			fn(buf[:n])
		}
		if err == io.EOF {
			return length, nil
		}
	}
}

// MultiCloser 多个关闭合并
func MultiCloser(closer ...Closer) Closer {
	return &multiCloser{closer: closer}
}

// PublisherToWriter Publisher to Writer
func PublisherToWriter(p Publisher, topic string) Writer {
	return &publishToWriter{topic: topic, Publisher: p}
}

func NewReadWriter(r Reader, w Writer) ReadWriteCloser {
	return &readWrite{Reader: r, Writer: w}
}

// SwapClient 数据交换交换
func SwapClient(c1, c2 *Client) {
	c1.SetDealWithWriter(c2)
	c2.SetDealWithWriter(c1)
	go c1.Run()
	go c2.Run()
}

// SwapWithReadFunc 根据读取规则来交换数据(例如数据进行了加密),需要返回解密的字节
func SwapWithReadFunc(i1, i2 ReadWriteCloser, readFunc buf.ReadFunc) {
	c1 := NewClient(i1)
	c1.SetReadFunc(readFunc)
	c2 := NewClient(i2)
	c2.SetReadFunc(readFunc)
	SwapClient(c1, c2)
}

// Swap same two Copy IO数据交换
func Swap(i1, i2 ReadWriteCloser) {
	go Copy(i1, i2)
	Copy(i2, i1)
}

/*



 */

// multiCloser
// 合并多个Closer , 变成1个Closer
type multiCloser struct {
	closer []Closer
}

func (this *multiCloser) Close() (err error) {
	for _, v := range this.closer {
		if er := v.Close(); er != nil {
			err = er
		}
	}
	return
}

type publishToWriter struct {
	topic string
	Publisher
}

func (this *publishToWriter) Write(p []byte) (int, error) {
	err := this.Publisher.Publish(this.topic, p)
	return len(p), err
}

type readWrite struct {
	Reader
	Writer
}

func (this *readWrite) Close() error { return nil }

type Read func(p []byte) (int, error)

func (this Read) Read(p []byte) (int, error) {
	return this(p)
}

type Write func(p []byte) (int, error)

func (this Write) Write(p []byte) (int, error) {
	return this(p)
}
