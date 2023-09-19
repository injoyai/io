package io

import (
	"bufio"
	"io"
)

func NewMessageReader(r io.Reader, read ReadFunc) MessageReader {
	return &messageReader{bufio.NewReader(r), read}
}

func DealMessageReader(r MessageReader, fn DealFunc) error {
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

func DealReader(r io.Reader, fn func(buf *bufio.Reader) error) (err error) {
	buf := bufio.NewReader(r)
	for ; err == nil; err = fn(buf) {
	}
	return
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

// SwapClient 数据交换交换
func SwapClient(c1, c2 *Client) {
	c1.SetReadWithWriter(c2)
	c1.SetCloseWithCloser(c2)
	c2.SetReadWithWriter(c1)
	c2.SetCloseWithCloser(c1)
	go c1.Run()
	go c2.Run()
}

// Swap same two Copy IO数据交换
func Swap(i1, i2 ReadWriter) {
	go Copy(i1, i2)
	Copy(i2, i1)
}

/*



 */

type messageReader struct {
	buf  *bufio.Reader
	read ReadFunc
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
