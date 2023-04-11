package proxy

import (
	"errors"
	"fmt"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"strings"
)

type ConnectType uint8

func (this ConnectType) Uint8() uint8 {
	return uint8(this)
}

func (this ConnectType) String() string {
	switch this {
	case TCP:
		return "tcp"
	case UDP:
		return "udp"
	case Serial:
		return "serial"
	case File:
		return "file"
	case MQ:
		return "mq"
	case MQTT:
		return "mqtt"
	case HTTP:
		return "http"
	case Websocket:
		return "websocket"
	default:
		return "tcp"
	}
}

const (
	TCP       ConnectType = 0 // "tcp"
	UDP       ConnectType = 1 // "udp"
	Serial    ConnectType = 2 //"serial"
	File      ConnectType = 3 //"file"
	MQ        ConnectType = 4 //"mq"
	MQTT      ConnectType = 5 //"mqtt"
	HTTP      ConnectType = 6 //"http"
	Websocket ConnectType = 7 //"websocket"

	Check = 0x7E //校验位
)

type OperateType uint8

func (this OperateType) Uint8() uint8 {
	return uint8(this)
}

func (this OperateType) String() string {
	switch this {
	case Connect:
		return "connect"
	case Write:
		return "write"
	case Response:
		return "response"
	case Close:
		return "close"
	case Register:
		return "register"
	case Info:
		return "info"
	default:
		return "connect"
	}
}

const (

	// Connect Write 请求
	Connect OperateType = 0 //"connect" //代理,通讯,建立新的链接
	Write   OperateType = 1 //"write"   //代理,透传,写数据

	// Response Close 响应
	Response OperateType = 2 //"response" //代理,响应
	Close    OperateType = 3 //"close"    //代理,通讯,关闭链接

	Register OperateType = 11 //"register" //通讯,注册
	Info     OperateType = 12 //"info"     //通讯,和服务端建立通讯
)

type Message struct {
	OperateType OperateType `json:"ot"`   //操作类型
	ConnectType ConnectType `json:"ct"`   //连接类型 默认tcp
	Key         string      `json:"key"`  //会话标识
	Addr        string      `json:"addr"` //目标地址
	Data        string      `json:"data"` //内容
}

func (this *Message) Response(data []byte) *Message {
	this.SetOperateType(Response)
	this.SetData(data)
	return this
}

func (this *Message) Close(msg interface{}) *Message {
	this.SetOperateType(Close)
	this.SetData(msg)
	return this
}

func (this *Message) SetOperateType(_type OperateType) *Message {
	this.OperateType = _type
	return this
}

func (this *Message) SetConnectType(_type ConnectType) *Message {
	this.ConnectType = _type
	return this
}

func (this *Message) SetData(data interface{}) *Message {
	this.Data = conv.String(data)
	return this
}

func (this *Message) String() string {
	return string(this.Bytes())
}

func (this *Message) Digest() string {
	return fmt.Sprintf("标识:%s   地址:%s   类型:%s(%s)   :   %s", this.Key, this.Addr, this.OperateType, this.ConnectType, func() string {
		r := []rune(string(this.GetData()))
		if len(r) > 100 {
			r = r[:100]
		}
		s := string(r)
		s = strings.ReplaceAll(string(r), "\n", "\\n")
		return s
	}())
}

func (this *Message) Bytes() bytes.Entity {
	data := []byte{Check} //增加校验位,方便查找问题
	data = append(data, this.OperateType.Uint8())
	data = append(data, this.ConnectType.Uint8())
	data = append(data, uint8(len(this.Key)))
	data = append(data, []byte(this.Key)...)
	data = append(data, uint8(len(this.Addr)))
	data = append(data, []byte(this.Addr)...)
	data = append(data, conv.Bytes(uint32(len(this.Data)))...)
	data = append(data, []byte(this.Data)...)
	return data
}

func (this *Message) GetData() []byte {
	return []byte(this.Data)
}

func (this *Message) GetDataString() string {
	return string(this.GetData())
}

func DecodeMessage(bytes []byte) (*Message, error) {
	length := len(bytes)
	if length < 9 {
		return nil, errors.New("数据长度错误:基础长度错误")
	}
	if bytes[0] != Check {
		return nil, fmt.Errorf("数据校验位错误,预期(%x),得到(%x)", Check, bytes[0])
	}
	bytes = bytes[1:]
	length--
	m := new(Message)
	m.OperateType = OperateType(bytes[0])
	m.ConnectType = ConnectType(bytes[1])
	keyLen := int(bytes[2])
	if length < 8+keyLen {
		return nil, errors.New("数据长度错误:key长度错误")
	}
	m.Key = string(bytes[3 : 3+keyLen])
	addrLen := int(bytes[3+keyLen])
	if length < 8+keyLen+addrLen {
		return nil, errors.New("数据长度错误:addr长度错误")
	}
	m.Addr = string(bytes[4+keyLen : 4+keyLen+addrLen])
	dataLen := conv.Int(bytes[4+keyLen+addrLen : 8+keyLen+addrLen])
	if length != 8+keyLen+addrLen+dataLen {
		logs.Debug(bytes[4+keyLen+addrLen : 8+keyLen+addrLen])
		logs.Debug(bytes[:8+keyLen+addrLen])
		return nil, fmt.Errorf("数据长度错误:data长度错误,预期(%d),得到(%d)", 8+keyLen+addrLen+dataLen, length)
	}
	m.Data = string(bytes[8+keyLen+addrLen:])
	return m, nil
}

func NewCloseMessage(key, data string) *Message {
	return (&Message{
		Key:         key,
		OperateType: Close,
	}).SetData(data)
}

func NewConnectMessage(key, addr string) *Message {
	return &Message{
		Key:         key,
		Addr:        addr,
		OperateType: Connect,
	}
}

func NewWriteMessage(key, addr string, data []byte) *Message {
	return (&Message{
		Key:         key,
		Addr:        addr,
		OperateType: Write,
	}).SetData(data)
}

func NewRegisterMessage(key, data string) *Message {
	return &Message{
		Key:         key,
		OperateType: Register,
		Data:        data,
	}
}

func NewInfoMessage(key string, data []byte) *Message {
	return (&Message{
		Key:         key,
		OperateType: Info,
	}).SetData(data)
}

type CMessage struct {
	*io.Client
	*Message
}

func NewCMessage(c *io.Client, m *Message) *CMessage {
	return &CMessage{Client: c, Message: m}
}
