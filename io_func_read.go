package io

import (
	"bufio"
	"github.com/injoyai/io/internal/frame"
	"io"
	"time"
)

func DealReader(r io.Reader, fn DealReaderFunc) (err error) {
	buf := bufio.NewReader(r)
	for ; err == nil; err = fn(buf) {
	}
	return
}

func ReadWithKB4(buf *bufio.Reader) ([]byte, error) {
	bytes := make([]byte, KB4)
	length, err := buf.Read(bytes)
	return bytes[:length], err
}

func ReadWithAll(buf *bufio.Reader) (bytes []byte, err error) {
	//read,单次读取大小不影响速度
	num := KB4
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

func ReadWithLine(buf *bufio.Reader) (bytes []byte, err error) {
	bytes, _, err = buf.ReadLine()
	return
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

// NewReadWithWriter 新建读取到数据立马写入到io.writer
func NewReadWithWriter(write io.Writer) ReadFunc {
	return func(buf *bufio.Reader) (bytes []byte, err error) {
		_, err = io.Copy(write, buf)
		return
	}
}

// NewReadWithStartEnd 新建buf.Reader , 根据帧头帧尾
func NewReadWithStartEnd(start, end []byte) ReadFunc {
	f := &frame.Frame{StartEndFrame: &frame.StartEndFrame{Start: start, End: end}}
	return f.ReadMessage
}

// NewWriteWithStartEnd 新建buf.Writer,根据帧头帧尾
func NewWriteWithStartEnd(start, end []byte) WriteFunc {
	return func(req []byte) ([]byte, error) {
		return append(start, append(req, end...)...), nil
	}
}

// NewReadWithLen 根据长度配置分包
func NewReadWithLen(l *frame.LenFrame) ReadFunc {
	f := &frame.Frame{LenFrame: l}
	return f.ReadMessage
}

// NewReadWithTimeout 读取全部数据,根据超时时间分包
func NewReadWithTimeout(timeout time.Duration) ReadFunc {
	f := &frame.Frame{Timeout: timeout}
	return f.ReadMessage
}

// NewReadWithFrame 根据Frame配置读取数据
func NewReadWithFrame(f *frame.Frame) ReadFunc {
	return f.ReadMessage
}
