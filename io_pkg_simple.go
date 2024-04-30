package io

import (
	"bufio"
	"fmt"
	"github.com/injoyai/base/g"
	"github.com/injoyai/conv"
	"io"
)

/*

简易封装包

帧头 1字节  0x68
长度 2字节
控制 1字节
数据 n字节
校验 1字节


控制码:
.=======================================================================================.
|bit7				|bit6			|bit5~0												|
|---------------------------------------------------------------------------------------|
|数据方向0请求,1响应	|响应是否有错误	|1读取,2写入,3订阅								    |
^=======================================================================================^



数据域:
数据类型 1字节
长度 1字节
数据  n字节
...




*/

type SimpleControl struct {
	IsResponse bool
	IsErr      bool
	Type       uint8
}

func (this SimpleControl) Byte() uint8 {
	b := this.Type
	if this.IsResponse {
		b |= 0x80
	}
	if this.IsErr {
		b |= 0x40
	}
	return b
}

type Simple struct {
	Control SimpleControl //控制码,基本信息,方向,错误等 类型,1读,2写,3订阅,4通知
	MsgID   uint8         //消息序号
	Data    SimpleData    //数据
}

func (this *Simple) Resp(data SimpleData, err ...error) *Simple {
	if data == nil {
		data = SimpleData{}
	}
	this.Control.IsResponse = true
	if len(err) > 0 && err[0] != nil {
		this.Control.IsErr = true
		data[FliedError] = conv.Bytes(err[0])
	}
	this.Data = data
	return this
}

func (this *Simple) Bytes() g.Bytes {
	bs := []byte{0x68}
	data := this.Data.Bytes()
	length := uint16(len(data) + 3)        // 1(报文头)+2(长度)
	bs = append(bs, conv.Bytes(length)...) //后续数据长度
	bs = append(bs, this.Control.Byte())   //控制码
	bs = append(bs, this.MsgID)            //消息id
	bs = append(bs, data...)               //数据
	bs = append(bs, this.sum(bs))          //校验
	return bs
}

func (this *Simple) sum(bs []byte) byte {
	var sum byte
	for _, v := range bs {
		sum += v
	}
	return sum
}

// SimpleData key和value的长度不能超过255
type SimpleData map[string][]byte

func (this SimpleData) Bytes() g.Bytes {
	data := []byte(nil)
	for k, v := range this {
		data = append(data, byte(len(k)))
		data = append(data, k...)
		data = append(data, byte(len(v)))
		data = append(data, v...)
	}
	return data
}

func (this SimpleData) SMap() map[string]string {
	data := map[string]string{}
	for k, v := range this {
		data[k] = string(v)
	}
	return data
}

func NewSimplePing() *Simple {
	return &Simple{
		Control: SimpleControl{
			Type: OprPing,
		},
	}
}

func NewSimple(control SimpleControl, data SimpleData, msgID ...uint8) *Simple {
	return &Simple{
		Control: control,
		Data:    data,
		MsgID:   conv.GetDefaultUint8(0, msgID...),
	}
}

func DecodeSimple(bs []byte) (*Simple, error) {
	if len(bs) < 7 {
		return nil, fmt.Errorf("数据长度小于(%d)", 7)
	}
	if bs[0] != 0x68 {
		return nil, fmt.Errorf("帧头错误,预期(0x68),得到(%x)", bs[0])
	}
	length := conv.Int(bs[1:3])
	if len(bs) != length+3 {
		return nil, fmt.Errorf("数据总长度错误,预期(%d),得到(%d)", length+3, len(bs))
	}

	p := &Simple{
		Control: SimpleControl{
			IsResponse: bs[3]&0x80 == 0x80,
			IsErr:      bs[3]&0x40 == 0x40,
			Type:       bs[3] & 0x3F,
		},
		MsgID: bs[4],
		Data:  map[string][]byte{},
	}
	sum := p.sum(bs[:len(bs)-1])
	if sum != bs[len(bs)-1] {
		return nil, fmt.Errorf("数据校验错误,预期(%x),得到(%x)", sum, bs[len(bs)-1])
	}

	data := bs[5 : len(bs)-1]
	for len(data) > 0 {
		keyLen := data[0]
		if len(data) < int(1+keyLen) {
			break
		}
		k := string(data[1 : 1+keyLen])

		valLen := data[1+keyLen]
		if len(data) < int(1+keyLen+1+valLen) {
			break
		}
		v := data[1+keyLen+1 : 1+keyLen+1+valLen]
		p.Data[k] = v

		data = data[1+keyLen+1+valLen:]
	}

	return p, nil

}

func WriteWithSimple(bs []byte) ([]byte, error) {
	return bs, nil
}

func ReadWithSimple(r *bufio.Reader) ([]byte, error) {
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		if b == 0x68 {
			result := []byte{0x68}
			buf := make([]byte, 2)
			_, err = io.ReadAtLeast(r, buf, 2)
			if err != nil {
				return nil, err
			}
			result = append(result, buf...)

			length := conv.Int(buf)
			buf = make([]byte, length)
			_, err = io.ReadAtLeast(r, buf, length)
			if err != nil {
				return nil, err
			}
			result = append(result, buf...)

			//校验
			if _, err = DecodeSimple(result); err == nil {
				return result, nil
			}

		}
	}
}
