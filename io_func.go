package io

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// NewMReader Reader转MessageReader
func NewMReader(r io.Reader, read func(buf *bufio.Reader) ([]byte, error)) MReader {
	return &mReader{bufio.NewReader(r), read}
}

// NewAReader Reader转AckReader
func NewAReader(r io.Reader, read func(buf *bufio.Reader) ([]byte, error)) AReader {
	return &aReader{bufio.NewReader(r), read}
}

// DealMReader 处理MessageReader
func DealMReader(r MReader, fn func(msg Message) error) error {
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

// DealAReader 处理AckReader
func DealAReader(r AReader, fn func(msg Message) error) error {
	for {
		a, err := r.ReadAck()
		if err != nil {
			return err
		}
		if err = fn(a.Payload()); err != nil {
			return err
		}
		a.Ack()
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
	if i, ok := r.(ByteReader); ok {
		return i.ReadByte()
	}
	b := make([]byte, 1)
	_, err := io.ReadAtLeast(r, b, 1)
	return b[0], err
}

// SplitWithLength 按最大长度分割字节
func SplitWithLength(p []byte, max uint64) [][]byte {
	if max == 0 {
		return [][]byte{}
	}
	list := [][]byte(nil)
	for uint64(len(p)) > max {
		list = append(list, p[:max])
		p = p[max:]
	}
	list = append(list, p)
	return list
}

// ReadLeast 读取最少least字节,除非返回错误
func ReadLeast(r io.Reader, least int) ([]byte, error) {
	buf := make([]byte, least)
	n, err := io.ReadAtLeast(r, buf, least)
	return buf[:n], err
}

func ReadFuncToAck(f func(r *bufio.Reader) ([]byte, error)) func(r *bufio.Reader) (Acker, error) {
	return func(r *bufio.Reader) (Acker, error) {
		a, err := f(r)
		return Ack(a), err
	}
}

func NewReadWriteCloser(r io.Reader, w io.Writer, c io.Closer) io.ReadWriteCloser {
	if c == nil {
		c = Null
	}
	if w == nil {
		w = Null
	}
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{r, w, c}
}

func NewAReadWriteCloser(r AReader, w Writer, c io.Closer) AReadWriteCloser {
	if c == nil {
		c = Null
	}
	if w == nil {
		w = Null
	}
	return struct {
		AReader
		Writer
		io.Closer
	}{r, w, c}
}

func NewMReadWriteCloser(r MReader, w Writer, c io.Closer) MReadWriteCloser {
	if c == nil {
		c = Null
	}
	if w == nil {
		w = Null
	}
	return struct {
		MReader
		Writer
		io.Closer
	}{r, w, c}
}

/*




 */

// Copy 如何使用接口约束 [T Reader | MReader | AReader]
func Copy[T any](w Writer, r T) (int64, error) {
	return CopyWith(w, r, nil)
}

// CopyWith 如何使用接口约束 [T Reader | MReader | AReader]
// 复制数据,Reader类型每次固定4KB,并提供函数监听
func CopyWith[T any](w Writer, r T, f func(p []byte) ([]byte, error)) (int64, error) {
	return CopyNWith(w, r, 32*1024, f)
}

// CopyNWith 复制数据,每次固定大小,并提供函数监听
// 如何使用接口约束 [T Reader | MReader | AReader]
func CopyNWith[T any](w Writer, r T, max int64, f func(p []byte) ([]byte, error)) (int64, error) {

	read := func() (Acker, error) {
		switch v := interface{}(r).(type) {
		case Reader:
			buf := make([]byte, max)
			n, err := v.Read(buf)
			if err != nil {
				return nil, err
			}
			return Ack(buf[:n]), nil

		case MReader:
			bs, err := v.ReadMessage()
			return Ack(bs), err

		case AReader:
			return v.ReadAck()

		default:
			return nil, fmt.Errorf("未知类型: %T, 未实现[Reader|MReader|AReader]", r)

		}
	}

	for co, n := int64(0), 0; ; co += int64(n) {
		a, err := read()
		if err != nil {
			if err == io.EOF {
				return co, nil
			}
			return 0, err
		}
		bs := a.Payload()
		if f != nil {
			bs, err = f(bs)
			if err != nil {
				return 0, err
			}
		}

		n, err = w.Write(bs)
		if err != nil {
			return 0, err
		}
		a.Ack()
	}

}

// Swap 如何使用接口约束 [T ReadWriter | MReadWriter | AReadWriter]
func Swap[T io.Writer](i1, i2 T) error {
	go Copy(interface{}(i1).(io.Writer), i2)
	_, err := Copy(interface{}(i2).(io.Writer), i1)
	return err
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

// SwapClose 交换数据并关闭
// 约束[T ReadWritCloser | MReadWritCloser | AReadWritCloser]
func SwapClose(c1, c2 WriteCloser) error {
	defer c1.Close()
	defer c2.Close()
	return Swap(c1, c2)
}

// Bridge 桥接,桥接两个ReadWriter
// 例如,桥接串口(客户端)和网口(tcp客户端),可以实现通过串口上网
func Bridge(i1, i2 Writer) error {
	return Swap(i1, i2)
}

// CopyWithPlan 复制数据,返回进度情况
func CopyWithPlan(w Writer, r Reader, f func(p *Plan)) (int64, error) {
	p := &Plan{}
	return CopyWith(w, r, func(buf []byte) ([]byte, error) {
		if f != nil {
			p.Index++
			p.Current += int64(len(buf))
			p.Bytes = buf
			f(p)
		}
		return buf, nil
	})
}
