package proxy

import (
	"bufio"
	"encoding/base64"
	"github.com/injoyai/io/buf"
)

var (
	defaultStart    = []byte{0x03, 0x03}
	defaultEnd      = []byte{0x04, 0x04}
	defaultReadFunc = buf.NewReadWithStartEnd(defaultStart, defaultEnd)
)

// DefaultReadFunc 默认读函数
func DefaultReadFunc(buf *bufio.Reader) ([]byte, error) {
	req, err := defaultReadFunc(buf)
	if err != nil {
		return nil, err
	}
	if len(req) > len(defaultStart)+len(defaultEnd) {
		req = req[len(defaultStart) : len(req)-len(defaultEnd)]
	}
	return base64.StdEncoding.DecodeString(string(req))
}

// DefaultWriteFunc 默认写函数
func DefaultWriteFunc(req []byte) []byte {
	req = []byte(base64.StdEncoding.EncodeToString(req))
	req = append(append(defaultStart, req...), defaultEnd...)
	return req
}
