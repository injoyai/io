package pipe

import (
	"bufio"
	"github.com/injoyai/io/buf"
)

var (
	defaultStart    = []byte{0x88, 0x88, 0x88, 0x88}
	defaultEnd      = []byte{0x89, 0x89, 0x89, 0x89}
	defaultReadFunc = buf.NewReadWithStartEnd(defaultStart, defaultEnd)
)

//// DefaultReadFunc 默认使用base64解码
//func DefaultReadFunc(buf *bufio.Reader) ([]byte, error) {
//	req, err := defaultReadFunc(buf)
//	if err != nil {
//		return nil, err
//	}
//	if len(req) > len(defaultStart)+len(defaultEnd) {
//		req = req[len(defaultStart) : len(req)-len(defaultEnd)]
//	}
//	return DefaultDecode(req)
//}
//
//// DefaultWriteFunc 默认使用base64编码
//func DefaultWriteFunc(req []byte) ([]byte, error) {
//	req = []byte(base64.StdEncoding.EncodeToString(req))
//	req = append(append(defaultStart, req...), defaultEnd...)
//	return req, nil
//}
//
//// DefaultDecode 默认数据解码
//func DefaultDecode(req []byte) ([]byte, error) {
//	return base64.StdEncoding.DecodeString(string(req))
//}

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
