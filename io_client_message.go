package io

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"io"
	"strings"
)

const (
	TagRead   = "接收"
	TagWrite  = "发送"
	TagClose  = "关闭"
	TagRedial = "重连"
	TagErr    = "错误"
	Ping      = "ping"
	Pong      = "pong"
)

func NewMessage(bs []byte) *Message {
	return &Message{
		Buffer: bytes.NewBuffer(bs),
	}
}

type Message struct {
	*bytes.Buffer
}

func (this *Message) Error() string {
	return this.String()
}

func (this *Message) HEX() string {
	return strings.ToUpper(hex.EncodeToString(this.Bytes()))
}

func (this *Message) ASCII() string {
	return string(this.Bytes())
}

func (this *Message) Base64() string {
	return base64.StdEncoding.EncodeToString(this.Bytes())
}

func (this *Message) WriteBytes(bs []byte) error {
	_, err := this.Buffer.Write(bs)
	return err
}

func (this *Message) CopyTo(dst io.Writer) error {
	_, err := io.Copy(dst, this)
	return err
}

func (this *Message) CopyOf(src io.Reader) error {
	_, err := io.Copy(this, src)
	return err
}

func (this *Message) Reader() *bufio.Reader {
	return bufio.NewReader(this)
}

func (this *Message) Writer() *bufio.Writer {
	return bufio.NewWriter(this)
}
