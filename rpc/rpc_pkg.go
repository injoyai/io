package rpc

import (
	"fmt"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"hash/crc32"
)

const (
	PStart byte = 0x03
	PEnd   byte = 0x04
)

type Type int

const (
	String byte = 0x01
	Bool   byte = 0x02
	Int    byte = 0x03
	Float  byte = 0x04
)

/*
.===================================.
|构成	|字节	|类型	|说明		|
|-----------------------------------|
|帧头 	|1字节 	|Byte	|固定0x03	|
|-----------------------------------|
|帧长  	|2字节	|HEX	|大端		|
|-----------------------------------|
|帧类型	|1字节	|Bin	|详见帧类型	|
|-----------------------------------|
|消息号	|1字节	|Byte	|消息id		|
|-----------------------------------|
|内容	|可变	|Byte	|数据内容	|
|-----------------------------------|
|校验和	|4字节	|Byte	|crc IEEE	|
|-----------------------------------|
|帧尾 	|1字节	|Byte	|固定0x04	|
^===================================^


*/

const (
	minLength = 10
)

type Pkg struct {
	Type  uint8
	MsgID uint8
	Data  []byte
}

func (this *Pkg) Bytes() bytes.Entity {
	data := []byte{PStart}
	length := len(this.Data) + minLength
	data = append(data, byte(length>>8), byte(length))
	data = append(data, this.Type)
	data = append(data, this.MsgID)
	data = append(data, this.Data...)
	data = append(data, conv.Bytes(crc32.ChecksumIEEE(data))...)
	data = append(data, PEnd)
	return data
}

func DecodePkg(bs []byte) (*Pkg, error) {

	//校验基础数据长度
	if len(bs) <= 10 {
		return nil, fmt.Errorf("数据长度小于(%d)", minLength)
	}

	//校验帧头
	if bs[0] != PStart {
		return nil, fmt.Errorf("帧头错误,需要(%x),得到(%x)", PStart, bs[0])
	}

	//获取总数据长度
	length := conv.Int(bs[1:3])

	//校验总长度
	if len(bs) != length {
		return nil, fmt.Errorf("数据总长度错误,需要(%d),得到(%d)", length, len(bs))
	}

	//校验crc32
	if crc1, crc2 := crc32.ChecksumIEEE(bs[:length-5]), conv.Uint32(bs[length-5:length-1]); crc1 != crc2 {
		return nil, fmt.Errorf("数据CRC校验错误,需要(%x),得到(%x)", crc1, crc2)
	}

	return &Pkg{
		Type:  bs[3],
		MsgID: bs[4],
		Data:  bs[4 : length-5],
	}, nil

}
