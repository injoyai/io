package io

import (
	"bufio"
	"bytes"
)

type AckReader interface {
	ReadAck() (Acker, error)
}

type Acker interface {
	Payload() []byte
	Ack() error
}

// MessageReader 读取分包后的数据
type MessageReader interface {
	ReadMessage() ([]byte, error)
}

type MessageReadCloser interface {
	MessageReader
	Closer
}

type MessageReadWriteCloser interface {
	MessageReader
	Writer
	Closer
}

//=============================

func MReaderToReader(mReader MessageReader) *_mReaderToReader {
	return &_mReaderToReader{
		mReader: mReader,
		buff:    bytes.NewBuffer(nil),
	}
}

// _mReaderToReader 把MessageReader转成Reader,
// 这样MessageReader能兼容
type _mReaderToReader struct {
	mReader MessageReader
	buff    *bytes.Buffer
}

func (this *_mReaderToReader) ReadFunc(r *bufio.Reader) ([]byte, error) {
	return this.mReader.ReadMessage()
}

func (this *_mReaderToReader) Read(p []byte) (int, error) {
	if this.buff.Len() == 0 {
		bs, err := this.mReader.ReadMessage()
		if err != nil {
			return 0, err
		}
		//使用buff,可能用户读取的数据比readMessage读取到的少,
		//缓存多余的数据再下次读取,当缓存的数据为0时,再去读取
		this.buff.Write(bs)
	}
	return this.buff.Read(p)
}

//=============================

func AReaderToReader(aReader AckReader) *_aReaderToReader {
	return &_aReaderToReader{
		mReader: aReader,
		buff:    bytes.NewBuffer(nil),
	}
}

type _aReaderToReader struct {
	mReader AckReader
	buff    *bytes.Buffer
}

func (this *_aReaderToReader) ReadAck(r *bufio.Reader) (Acker, error) {
	return this.mReader.ReadAck()
}

func (this *_aReaderToReader) Read(p []byte) (int, error) {
	if this.buff.Len() == 0 {
		ack, err := this.mReader.ReadAck()
		if err != nil {
			return 0, err
		}
		//使用buff,可能用户读取的数据比readMessage读取到的少,
		//缓存多余的数据再下次读取,当缓存的数据为0时,再去读取
		this.buff.Write(ack.Payload())
	}
	return this.buff.Read(p)
}

//=============================

type Ack []byte

func (this Ack) Ack() error { return nil }
func (this Ack) Payload() []byte {
	return this
}
