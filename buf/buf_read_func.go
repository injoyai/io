package buf

import (
	"bufio"
	"fmt"
	"io"
	"time"
)

type ReadFunc func(buf *bufio.Reader) (bytes []byte, err error)

// ReadWithAll 默认读取函数,读取全部数据
func ReadWithAll(buf *bufio.Reader) (bytes []byte, err error) {
	//read 1 KB
	num := 1 << 10
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

// ReadWithAllSafe 安全读取, todo 好像bufio.Read出现过一次数组越界 , 待确认
func ReadWithAllSafe(buf *bufio.Reader) (bytes []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	return ReadWithAll(buf)
}

// ReadWithLine 读取一行
func ReadWithLine(buf *bufio.Reader) (bytes []byte, err error) {
	bytes, _, err = buf.ReadLine()
	return
}

/*



 */

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

func NewReadWithFrame(f *Frame) ReadFunc {
	return f.ReadMessage
}
