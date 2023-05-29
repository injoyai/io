package io

import (
	"bufio"
	"io"
)

// CopyFunc 复制数据,每次固定32KB,并提供函数监听
func CopyFunc(w Writer, r Reader, fn func(buf []byte)) (int, error) {
	return CopyNFunc(w, r, 32<<20, fn)
}

// CopyNFunc 复制数据,每次固定大小,并提供函数监听
func CopyNFunc(w Writer, r Reader, n int64, fn func(buf []byte)) (int, error) {
	buff := bufio.NewReader(r)
	length := 0
	for {
		buf := make([]byte, n)
		n, err := buff.Read(buf)
		if err != nil && err != io.EOF {
			return length, err
		}
		length += n
		if _, err := w.Write(buf[:n]); err != nil {
			return length, err
		}
		if fn != nil {
			fn(buf[:n])
		}
		if err == io.EOF {
			return length, nil
		}
	}
}
