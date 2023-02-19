package buf

import (
	"bufio"
	"io"
)

type MessageReadCloser interface {
	MessageReader
	io.Closer
}

type MessageReader interface {
	// ReadMessage 读取拆包后的数据
	ReadMessage() ([]byte, error)
}

type messageReader struct {
	buf      *bufio.Reader
	readFunc ReadFunc
}

// SetReadFunc 设置读取函数,默认读取全部
func (this *messageReader) SetReadFunc(fn ReadFunc) *messageReader {
	this.readFunc = fn
	return this
}

// ReadMessage 读取数据 实现接口 MessageReader
func (this *messageReader) ReadMessage() ([]byte, error) {
	if this.readFunc == nil {
		this.readFunc = ReadWithAll
	}
	return this.readFunc(this.buf)
}

func NewMessageReader(reader io.Reader, fn ReadFunc) MessageReader {
	m := &messageReader{buf: bufio.NewReader(reader)}
	m.SetReadFunc(fn)
	return m
}
