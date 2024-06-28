package io

import (
	"bufio"
	"bytes"
)

//cannot use io.ReadWriter in union (io.ReadWriter contains methods)
//type AllReader interface {
//	Reader | MReader | AReader
//}
//
//type AllReadWriter interface {
//	ReadWriter | MReadWriter | AReadWriter
//}
//
//type AllReadCloser interface {
//	ReadCloser | MReadCloser | AReadCloser
//}
//
//type AllReadWritCloser interface {
//	ReadWriteCloser | MReadWriteCloser | AReadWriteCloser
//}

//===============================AReader===============================

type AReader interface {
	ReadAck() (Acker, error)
}

type AReadWriter interface {
	AReader
	Writer
}

type AReadCloser interface {
	AReader
	Closer
}

type AReadWriteCloser interface {
	AReader
	Writer
	Closer
}

type Acker interface {
	Payload() []byte
	Ack() error
}

//===============================MReader===============================

// MReader 读取分包后的数据
type MReader interface {
	ReadMessage() ([]byte, error)
}

type MReadWriter interface {
	MReader
	Writer
}

type MReadCloser interface {
	MReader
	Closer
}

type MReadWriteCloser interface {
	MReader
	Writer
	Closer
}

/*



 */

//=============================

func MReaderToReader(mReader MReader) *_mReaderToReader {
	return &_mReaderToReader{
		mReader: mReader,
		buff:    bytes.NewBuffer(nil),
	}
}

// _mReaderToReader 把MessageReader转成Reader,
// 这样MessageReader能兼容
type _mReaderToReader struct {
	mReader MReader
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

func AReaderToReader(aReader AReader) *_aReaderToReader {
	return &_aReaderToReader{
		mReader: aReader,
		buff:    bytes.NewBuffer(nil),
	}
}

type _aReaderToReader struct {
	mReader AReader
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
