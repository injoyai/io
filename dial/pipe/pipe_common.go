package pipe

import (
	"bufio"
	"context"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"github.com/injoyai/logs"
)

var (
	defaultStart    = []byte{0x88, 0x88}
	defaultEnd      = []byte{0x89, 0x89}
	defaultReadFunc = buf.NewReadWithStartEnd(defaultStart, defaultEnd)
)

// DefaultReadFunc 默认使用base64解码
func DefaultReadFunc(buf *bufio.Reader) ([]byte, error) {
	req, err := defaultReadFunc(buf)
	if err != nil {
		return nil, err
	}
	if len(req) > len(defaultStart)+len(defaultEnd) {
		req = req[len(defaultStart) : len(req)-len(defaultEnd)]
	}
	return req, nil
}

// DefaultWriteFunc 默认使用base64编码
func DefaultWriteFunc(req []byte) ([]byte, error) {
	return append(append(defaultStart, req...), defaultEnd...), nil
}

func WithServer(s *io.Server) {
	s.SetWriteFunc(DefaultWriteFunc)
	s.SetReadFunc(DefaultReadFunc)
	s.SetPrintFunc(func(msg io.Message, tag ...string) {
		io.PrintWithASCII(msg, append([]string{"PI|S"}, tag...)...)
	})
}

func WithClient(ctx context.Context, c *io.Client) {
	c.SetPrintFunc(func(msg io.Message, tag ...string) {
		io.PrintWithASCII(msg, append([]string{"PI|C"}, tag...)...)
		logs.Debug(msg.HEX())
	})
	c.SetWriteFunc(DefaultWriteFunc)
	c.SetReadFunc(DefaultReadFunc)
	c.SetKeepAlive(io.DefaultTimeout)
}
