package io

import (
	"bufio"
	"fmt"
	"github.com/injoyai/base/g"
	"github.com/injoyai/conv"
	"github.com/injoyai/io/buf"
	"net"
)

/*
FrameBridge
桥接数据帧,用于转发字节数据
由起始帧,长度,和客户端地址,客户端数据组成
不做crc等校验,底层tcp已经做了,简洁点
长度取4字节
*/
type FrameBridge struct {
	Function   FunctionBridge //功能码
	ListenPort uint16         //监听的端口
	IP         net.IP         //客户端ip
	Port       uint16         //客户端端口
	Data       []byte         //字节数据
}

func (this *FrameBridge) Bytes() g.Bytes {
	data := []byte(nil)
	data = append(data, pkgStart...)                             //帧头
	data = append(data, conv.Bytes(uint16(len(this.Data)+8))...) //后续数据长度
	data = append(data, uint8(this.Function))                    //功能码
	data = append(data, conv.Bytes(this.ListenPort)...)          //监听的端口
	data = append(data, this.IP.To4()...)                        //客户端的ip
	data = append(data, conv.Bytes(this.Port)...)                //客户端的端口
	data = append(data, this.Data...)                            //数据域部分
	return data
}

func (this *FrameBridge) Address() string {
	return fmt.Sprintf("%s:%d", this.IP.String(), this.Port)
}

type FunctionBridge uint8

const (
	BridgeFunctionSubscribeAddress FunctionBridge = 0x1 //桥接客户端
	BridgeFunctionSubscribePort    FunctionBridge = 0x2 //桥接端口

	BridgeFunctionError         FunctionBridge = 0x80 //自定义错误
	BridgeFunctionErrorNoListen FunctionBridge = 0x81 //没有监听
	BridgeFunctionErrorNoClient FunctionBridge = 0x81 //没有客户端

)

func NewFrameBridge(ip net.IP, port uint16, data []byte) []*FrameBridge {
	list := []*FrameBridge(nil)
	for _, bs := range SplitWithLength(data, 65500) {
		p := &FrameBridge{
			IP:   ip,
			Port: port,
			Data: bs,
		}
		list = append(list, p)
	}
	return list
}

func NewFrameBridgeBytes(ip net.IP, port uint16, data []byte) [][]byte {
	list := [][]byte(nil)
	for _, p := range NewFrameBridge(ip, port, data) {
		list = append(list, p.Bytes())
	}
	return list
}

func DecodeBridge(bs g.Bytes) (*FrameBridge, error) {
	baseLength := 13
	if len(bs) < baseLength {
		return nil, fmt.Errorf("基础数据长度错误:,预期(%d),得到(%d)", baseLength, bs.Len())
	}
	length := conv.Int(bs[2:4])
	if len(bs) != length+4 {
		return nil, fmt.Errorf("数据总长度错误,预期(%d),得到(%d)", length+4, len(bs))
	}
	return &FrameBridge{
		Function:   FunctionBridge(bs[4]),
		ListenPort: conv.Uint16(bs[5:7]),
		IP:         net.IP(bs[7:11]),
		Port:       conv.Uint16(bs[11:13]),
		Data:       bs[13:],
	}, nil
}

func ReadWithFrameBridge(r *bufio.Reader) ([]byte, error) {
	result := []byte(nil)
	if _, err := buf.ReadPrefix(r, pkgStart); err != nil {
		return nil, err
	}
	result = append(result, pkgStart...)
	buffer := make([]byte, 2)
	if _, err := ReadAtLeast(r, buffer, 2); err != nil {
		return nil, err
	}
	result = append(result, buffer...)
	length := conv.Int(buffer)
	buffer = make([]byte, length)
	if _, err := ReadAtLeast(r, buffer, length); err != nil {
		return nil, err
	}
	result = append(result, buffer...)
	return result, nil
}
