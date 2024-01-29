package buf

import (
	"bufio"
	"bytes"
	"io"
	"time"
)

type (
	ReadFunc  func(buf *bufio.Reader) (bytes []byte, err error)
	WriteFunc func(req []byte) ([]byte, error)
)

// ReadWithAll 默认读取函数,读取全部数据
func ReadWithAll(buf *bufio.Reader) (bytes []byte, err error) {
	//read,单次读取大小不影响速度
	num := 4096
	for {
		data := make([]byte, num)
		length, err := buf.Read(data)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, data[:length]...)
		if length < num || buf.Buffered() == 0 {
			//缓存没有剩余的数据
			return bytes, err
		}
	}
}

// ReadWithLine 读取一行
func ReadWithLine(buf *bufio.Reader) (bytes []byte, err error) {
	bytes, _, err = buf.ReadLine()
	return
}

// ReadPrefix 从流中读取到头部,并返回已读取部分
func ReadPrefix(buf *bufio.Reader, prefix []byte) ([]byte, error) {
	cache := []byte(nil)
	for index := 0; index < len(prefix); {
		b, err := buf.ReadByte()
		if err != nil {
			return cache, err
		}
		cache = append(cache, b)
		if b == prefix[index] {
			index++
		} else {
			for len(cache) > 0 {
				//only one error in this ReadPrefix ,it is EOF,and not important
				cache2, _ := ReadPrefix(bufio.NewReader(bytes.NewReader(cache[1:])), prefix)
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

// ReadLeast 读取至少x字节,除非返回错误
func ReadLeast(r *bufio.Reader, min int) ([]byte, error) {
	buf := make([]byte, min)
	n, err := io.ReadAtLeast(r, buf, min)
	return buf[:n], err
}

/*



 */

func NewReadWithKB(n uint) func(buf *bufio.Reader) ([]byte, error) {
	return func(buf *bufio.Reader) ([]byte, error) {
		bytes := make([]byte, n<<10)
		length, err := buf.Read(bytes)
		return bytes[:length], err
	}
}

func NewReadWithMB(n uint) func(buf *bufio.Reader) ([]byte, error) {
	return func(buf *bufio.Reader) ([]byte, error) {
		bytes := make([]byte, n<<20)
		length, err := buf.Read(bytes)
		return bytes[:length], err
	}
}

// NewReadWithWriter 新建读取到数据立马写入到io.writer
func NewReadWithWriter(write io.Writer) ReadFunc {
	return func(buf *bufio.Reader) (bytes []byte, err error) {
		_, err = io.Copy(write, buf)
		return
	}
}

// NewReadWithStartEnd 新建buf.Reader , 根据帧头帧尾
func NewReadWithStartEnd(start, end []byte) ReadFunc {
	f := &Frame{StartEndFrame: &StartEndFrame{Start: start, End: end}}
	return f.ReadMessage
}

// NewWriteWithStartEnd 新建buf.Writer,根据帧头帧尾
func NewWriteWithStartEnd(start, end []byte) WriteFunc {
	return func(req []byte) ([]byte, error) {
		return append(start, append(req, end...)...), nil
	}
}

// NewReadWithLen 根据长度配置分包
func NewReadWithLen(l *LenFrame) ReadFunc {
	f := &Frame{LenFrame: l}
	return f.ReadMessage
}

// NewReadWithTimeout 读取全部数据,根据超时时间分包
func NewReadWithTimeout(timeout time.Duration) ReadFunc {
	f := &Frame{Timeout: timeout}
	return f.ReadMessage
}

// NewReadWithFrame 根据Frame配置读取数据
func NewReadWithFrame(f *Frame) ReadFunc {
	return f.ReadMessage
}
