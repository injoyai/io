package proxy

import (
	"encoding/base64"
	"encoding/json"
	"github.com/injoyai/conv"
)

type ConnectType string

const (
	TCP       ConnectType = "tcp"
	UDP       ConnectType = "udp"
	Serial    ConnectType = "serial"
	File      ConnectType = "file"
	MQ        ConnectType = "mq"
	MQTT      ConnectType = "mqtt"
	HTTP      ConnectType = "http"
	Websocket ConnectType = "websocket"
)

type OperateType string

const (
	Connect  OperateType = "connect"  //代理,通讯,建立新的链接
	Write    OperateType = "write"    //代理,透传,写数据
	Close    OperateType = "close"    //代理,通讯,关闭链接
	Register OperateType = "register" //通讯,注册
	Info     OperateType = "info"     //通讯,和服务端建立通讯
)

type Message struct {
	OperateType OperateType `json:"ot"`   //操作类型
	ConnectType ConnectType `json:"ct"`   //连接类型 默认tcp
	Key         string      `json:"key"`  //标识
	Data        string      `json:"data"` //内容
	DataBytes   []byte      `json:"-"`    //内容字节,需要解析
	Addr        string      `json:"addr"` //目标地址
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
	this.Data = base64.StdEncoding.EncodeToString(conv.Bytes(data))
	return this
}

func (this *Message) String() string {
	return string(this.Bytes())
}

func (this *Message) Bytes() []byte {
	bs, _ := json.Marshal(this)
	return bs
}

func (this *Message) GetData() []byte {
	return this.DataBytes
}

func DecodeMessage(bytes []byte) (*Message, error) {
	m := new(Message)
	err := json.Unmarshal(bytes, m)
	if err == nil {
		m.DataBytes, err = base64.StdEncoding.DecodeString(m.Data)
	}
	return m, err
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
