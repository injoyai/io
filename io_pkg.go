package io

import (
	"compress/gzip"
	"fmt"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
	"hash/crc32"
)

/*

通用封装包

包构成(大端):
.===================================.
|构成	|字节	|类型	|说明		|
|-----------------------------------|
|帧头 	|2字节 	|Byte	|固定0x8888	|
|-----------------------------------|
|帧长  	|2字节	|HEX	|总字节长度	|
|-----------------------------------|
|帧类型	|2字节	|Bin	|详见帧类型	|
|-----------------------------------|
|消息号	|1字节	|Byte	|消息id		|
|-----------------------------------|
|内容	|可变	|Byte	|数据内容	|
|-----------------------------------|
|校验和	|4字节	|Byte	|crc IEEE	|
|-----------------------------------|
|帧尾 	|2字节	|Byte	|固定0x8989	|
^===================================^

包类型:
.=======================================================================================.
|bit15				|bit14			|bit13~11				|bit10	|bit9	|bit8		|
|---------------------------------------------------------------------------------------|
|数据方向0请求,1响应	|1测试通讯,无内容	|压缩方式,0无,1gzip	|预留							|
^=======================================================================================^
|bit7~0																					|
|---------------------------------------------------------------------------------------|
|预留																					|
^=======================================================================================^
*/

var (
	pkgStart = []byte{0x88, 0x88} //帧头
	pkgEnd   = []byte{0x89, 0x89} //帧尾
)

const (
	pkgBaseLength        = 13
	pkgBitBack    uint16 = 0x80 << 8
	pkgBitPing    uint16 = 0x40 << 8
)

func NewPkgPing() []byte {
	//01000000
	return (&Pkg{Type: pkgBitPing}).Bytes()
}

func NewPkgPong() []byte {
	//11000000
	return (&Pkg{Type: pkgBitBack + pkgBitPing}).Bytes()
}

func NewPkg(msgID uint8, data []byte) *Pkg {
	return &Pkg{
		Type:  pkgBitBack,
		MsgID: msgID,
		Data:  data,
	}
}

type Pkg struct {
	Type  uint16
	MsgID uint8
	Data  []byte
}

// SetCompress 设置压缩方式,1是gzip,其他不压缩
func (this *Pkg) SetCompress(n uint8) *Pkg {
	this.Type = this.Type>>14<<14 + uint16(n)<<13>>2 + this.Type<<11>>11
	return this
}

func (this *Pkg) String() string {
	return this.Bytes().HEX()
}

func (this *Pkg) EncodeData() []byte {
	data := this.Data
	switch this.Type << 2 >> 13 {
	case 1:
		// Gzip 压缩字节
		buf := bytes.NewBuffer(nil)
		gzipWriter := gzip.NewWriter(buf)
		gzipWriter.Write(data)
		gzipWriter.Close()
		data = buf.Bytes()
	default:
	}
	return data
}

func (this *Pkg) DecodeData() error {
	switch this.Type << 2 >> 13 {
	case 1:
		// Gzip 解压字节
		reader := bytes.NewReader(this.Data)
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(gzipReader)
		if err != nil {
			logs.Err(err)
			return err
		}
		this.Data = buf.Bytes()
	default:
	}
	return nil
}

func (this *Pkg) Bytes() bytes.Entity {
	data := []byte(nil)
	data = append(data, pkgStart...)
	dataBytes := this.EncodeData()
	length := len(dataBytes) + pkgBaseLength
	data = append(data, byte(length>>8), byte(length))
	data = append(data, byte(this.Type>>8), byte(this.Type))
	data = append(data, this.MsgID)
	data = append(data, dataBytes...)
	data = append(data, conv.Bytes(crc32.ChecksumIEEE(data))...)
	data = append(data, pkgEnd...)
	return data
}

// Resp 生成响应包
func (this *Pkg) Resp(bs []byte) *Pkg {
	this.Type += pkgBitBack
	this.Data = bs
	return this
}

// IsCall 是否请求数据
func (this *Pkg) IsCall() bool {
	return this.Type>>15 == 0
}

// IsBack 是否是响应数据
func (this *Pkg) IsBack() bool {
	return this.Type>>15 == 1
}

// IsPing 是否是ping,需要响应pong
func (this *Pkg) IsPing() bool {
	return this.Type>>14 == 1
}

// IsPong 是否是pong,不需要处理
func (this *Pkg) IsPong() bool {
	return this.Type>>14 == 3
}

// DecodePkg 按自定义的包解析
func DecodePkg(bs []byte) (*Pkg, error) {

	//校验基础数据长度
	if len(bs) <= pkgBaseLength {
		return nil, fmt.Errorf("数据长度小于(%d)", pkgBaseLength)
	}

	//校验帧头
	if bs[0] != pkgStart[0] && bs[1] != pkgStart[1] {
		return nil, fmt.Errorf("帧头错误,需要(%x),得到(%x)", pkgStart, bs[:2])
	}

	//获取总数据长度
	length := conv.Int(bs[2:4])

	//校验总长度
	if len(bs) != length {
		return nil, fmt.Errorf("数据总长度错误,预期(%d),得到(%d)", length, len(bs))
	}

	//校验crc32
	if crc1, crc2 := crc32.ChecksumIEEE(bs[:length-6]), conv.Uint32(bs[length-6:length-2]); crc1 != crc2 {
		return nil, fmt.Errorf("数据CRC校验错误,需要(%x),得到(%x)", crc1, crc2)
	}

	//校验帧尾
	if bs[length-2] != pkgEnd[0] && bs[length-1] != bs[1] {
		return nil, fmt.Errorf("帧尾错误,需要(%x),得到(%x)", pkgEnd, bs[length-2:])
	}

	p := &Pkg{
		Type:  uint16(bs[4])<<8 + uint16(bs[5]),
		MsgID: bs[6],
		Data:  bs[7 : length-6],
	}

	return p, p.DecodeData()

}
