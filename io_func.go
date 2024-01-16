package io

import (
	"bufio"
	"bytes"
	"io"
)

// NewMessageReader Reader转MessageReader
func NewMessageReader(r io.Reader, read func(buf *bufio.Reader) ([]byte, error)) MessageReader {
	return &messageReader{bufio.NewReader(r), read}
}

// DealMessageReader 处理MessageReader
func DealMessageReader(r MessageReader, fn func(msg Message) error) error {
	for {
		bs, err := r.ReadMessage()
		if err != nil {
			return err
		}
		if err = fn(bs); err != nil {
			return err
		}
	}
}

// DealReader 处理Reader
func DealReader(r io.Reader, fn func(buf *bufio.Reader) error) (err error) {
	buf := bufio.NewReader(r)
	for ; err == nil; err = fn(buf) {
	}
	return
}

// ReadPrefix 读取Reader符合的头部,返回成功(nil),或者错误
func ReadPrefix(r Reader, prefix []byte) ([]byte, error) {
	cache := []byte(nil)
	for index := 0; index < len(prefix); {
		b, err := ReadByte(r)
		if err != nil {
			return cache, err
		}
		cache = append(cache, b)
		if b == prefix[index] {
			index++
		} else {
			for len(cache) > 0 {
				//only one error in this ReadPrefix ,it is EOF,and not important
				cache2, _ := ReadPrefix(bytes.NewReader(cache[1:]), prefix)
				if len(cache2) > 0 {
					cache = cache2
					break
				}
				cache = cache[1:]
			}
			index = len(cache)
		}
	}
	return cache, nil
}

// ReadByte 读取一字节
func ReadByte(r Reader) (byte, error) {
	if i, ok := r.(interface{ ReadByte() (byte, error) }); ok {
		return i.ReadByte()
	}
	b := make([]byte, 1)
	_, err := io.ReadAtLeast(r, b, 1)
	return b[0], err
}

// ReadBytes 读取固定字节的数据
func ReadBytes(r Reader, length int) ([]byte, error) {
	if i, ok := r.(interface{ ReadBytes() ([]byte, error) }); ok {
		return i.ReadBytes()
	}
	bs := make([]byte, length)
	n, err := io.ReadAtLeast(r, bs, length)
	return bs[:n], err
}

/*




 */

// CopyWith 复制数据,每次固定4KB,并提供函数监听
func CopyWith(w Writer, r Reader, fn func(buf []byte)) (int, error) {
	return CopyNWith(w, r, DefaultBufferSize, fn)
}

// CopyNWith 复制数据,每次固定大小,并提供函数监听
func CopyNWith(w Writer, r Reader, n int64, fn func(buf []byte)) (int, error) {
	buff := bufio.NewReader(r)
	length := 0
	buf := make([]byte, n)
	for {
		num, err := buff.Read(buf)
		if err != nil && err != io.EOF {
			return length, err
		}
		length += num
		if fn != nil {
			fn(buf[:num])
		}
		if _, err := w.Write(buf[:num]); err != nil {
			return length, err
		}
		if err == io.EOF {
			return length, nil
		}
	}
}

/*



 */

// MultiCloser 多个关闭合并
func MultiCloser(closer ...Closer) Closer {
	return &multiCloser{closer: closer}
}

// PublisherToWriter Publisher to Writer
func PublisherToWriter(p Publisher, topic string) Writer {
	return &publishToWriter{topic: topic, Publisher: p}
}

func NewReadWriter(r Reader, w Writer) ReadWriteCloser {
	return &readWriter{Reader: r, Writer: w}
}

// SwapClient 数据交换
func SwapClient(c1, c2 *Client) {
	c1.SetReadWithWriter(c2)
	c1.SetCloseWithCloser(c2)
	c2.SetReadWithWriter(c1)
	c2.SetCloseWithCloser(c1)
	go c1.Run()
	go c2.Run()
}

// Swap same two Copy IO数据交换
func Swap(i1, i2 ReadWriter) error {
	go Copy(i1, i2)
	_, err := Copy(i2, i1)
	return err
}

// Bridge 桥接,桥接两个ReadWriter
// 例如,桥接串口(客户端)和网口(tcp客户端),可以实现通过串口上网
func Bridge(i1, i2 ReadWriter) error {
	return Swap(i1, i2)
}

/*



 */

type messageReader struct {
	buf  *bufio.Reader
	read func(buf *bufio.Reader) ([]byte, error)
}

func (this *messageReader) ReadMessage() ([]byte, error) {
	return this.read(this.buf)
}

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

type readWriter struct {
	Reader
	Writer
}

func (this *readWriter) Close() error { return nil }

type Read func(p []byte) (int, error)

func (this Read) Read(p []byte) (int, error) {
	return this(p)
}

type Write func(p []byte) (int, error)

func (this Write) Write(p []byte) (int, error) {
	return this(p)
}

//=================

// MustChan chan []byte 实现 io.Writer,必须等到写入成功为止
type MustChan chan []byte

func (this MustChan) Write(p []byte) (int, error) {
	this <- p
	return len(p), nil
}

// TryChan chan []byte 实现 io.Writer,尝试写入,不管是否成功
type TryChan chan []byte

func (this TryChan) Write(p []byte) (int, error) {
	select {
	case this <- p:
		return len(p), nil
	default:
		return 0, nil
	}
}

//====================

// Count 统计写入字节数量
type Count struct {
	io.Writer
	count int64
}

func (this *Count) Count() int64 {
	return this.count
}

func (this *Count) Write(p []byte) (int, error) {
	n, err := this.Writer.Write(p)
	this.count += int64(n)
	return n, err
}
