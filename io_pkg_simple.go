package io

import (
	"bufio"
	"fmt"
	"github.com/injoyai/base/g"
	"github.com/injoyai/conv"
	"io"
)

const (
	SimpleRead      = 0x01
	SimpleWrite     = 0x02
	SimpleSubscribe = 0x03
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
|数据方向0请求,1响应	|响应是否有错误	|预留
^=======================================================================================^



数据域:
数据类型 1字节
长度 1字节
数据  n字节
...




*/

type Simple struct {
	Control uint8 //控制码,基本信息,方向,错误等
	Type    uint8 //类型,1读,2写,3订阅
	Data    []SimpleKeyVal
}

func (this *Simple) sum(bs []byte) byte {
	var sum byte
	for _, v := range bs {
		sum += v
	}
	return sum
}

func (this *Simple) Bytes() g.Bytes {
	bs := []byte{0x68}
	data := []byte(nil)
	for _, v := range this.Data {
		data = append(data, v.Bytes()...)
	}
	length := uint16(len(data) + 3)
	bs = append(bs, conv.Bytes(length)...) //后续数据长度
	bs = append(bs, this.Control)          //控制码
	bs = append(bs, this.Type)             //数据类型
	bs = append(bs, data...)               //数据
	bs = append(bs, this.sum(bs))          //校验
	return bs
}

type SimpleKeyVal struct {
	Key string
	Val interface{}
}

func (this *SimpleKeyVal) Bytes() []byte {
	data := []byte(nil)
	data = append(data, byte(len(this.Key)))
	data = append(data, this.Key...)
	val := conv.Bytes(this.Val)
	data = append(data, byte(len(val)))
	data = append(data, val...)
	return data
}

func DecodeSimple(bs []byte) (*Simple, error) {
	if len(bs) < 6 {
		return nil, fmt.Errorf("数据长度小于(%d)", 6)
	}
	if bs[0] != 0x68 {
		return nil, fmt.Errorf("帧头错误,预期(0x68),得到(%x)", bs[0])
	}
	length := conv.Int(bs[1:3])
	if len(bs) != length+3 {
		return nil, fmt.Errorf("数据总长度错误,预期(%d),得到(%d)", length+3, len(bs))
	}
	p := &Simple{
		Control: bs[3],
		Type:    bs[4],
		Data:    make([]SimpleKeyVal, 0),
	}
	sum := p.sum(bs[:len(bs)-1])
	if sum != bs[len(bs)-1] {
		return nil, fmt.Errorf("数据校验错误,预期(%x),得到(%x)", sum, bs[len(bs)-1])
	}

	data := bs[5 : length-1]
	for len(data) > 0 {
		kv := SimpleKeyVal{}
		keyLen := data[0]
		if len(data) < int(1+keyLen) {
			break
		}
		kv.Key = string(data[1 : 1+keyLen])
		valLen := data[1+keyLen]
		if len(data) < int(1+keyLen+1+valLen) {
			break
		}
		kv.Val = string(data[1+keyLen+1 : 1+keyLen+1+valLen])
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
