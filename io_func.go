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

// CopyWith 复制数据,每次固定4KB,并提供函数监听
func CopyWith(w Writer, r Reader, fn func(buf []byte)) (int64, error) {
	return CopyNWith(w, r, DefaultBufferSize, fn)
}

// CopyWithPlan 复制数据,返回进度情况
func CopyWithPlan(w Writer, r Reader, f func(p *Plan)) (int64, error) {
	p := &Plan{
		Index:   0,
		Current: 0,
		Total:   0,
		Bytes:   nil,
	}
	return CopyWith(w, r, func(buf []byte) {
		p.Index++
		p.Current += int64(len(buf))
		p.Bytes = buf
		if f != nil {
			f(p)
		}
	})
}

// CopyNWith 复制数据,每次固定大小,并提供函数监听
func CopyNWith(w Writer, r Reader, n int64, fn func(buf []byte)) (int64, error) {
	buff := bufio.NewReader(r)
	length := int64(0)
	buf := make([]byte, n)
	for {
		num, err := buff.Read(buf)
		if err != nil && err != io.EOF {
			return length, err
		}
		length += int64(num)
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
