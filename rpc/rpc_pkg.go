package rpc

import (
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

type Pkg struct {
	Type  uint8
	MsgID uint8
	Data  []byte
}

func (this *Pkg) Bytes() bytes.Entity {
	data := []byte{PStart}
	length := len(this.Data) + 10
	data = append(data, byte(length>>8), byte(length))
	data = append(data, this.Type)
	data = append(data, this.MsgID)
	data = append(data, this.Data...)
	data = append(data, conv.Bytes(crc32.ChecksumIEEE(data))...)
	data = append(data, PEnd)
	return data
}
