package io

import (
	"bufio"
	"io"
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

//func
