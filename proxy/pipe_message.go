package proxy

import (
	"encoding/base64"
	"encoding/json"
)

const (
	Connect Type = "connect" //代理,通讯,建立新的链接
	Write   Type = "write"   //代理,透传,写数据
	Close   Type = "close"   //代理,通讯,关闭链接

	Register Type = "register" //通讯,注册
	Info     Type = "info"     //通讯,和服务端建立通讯
)

type Type string

type Message struct {
	Key  string `json:"key"`  //标识
	Data string `json:"data"` //内容
	Type Type   `json:"type"` //类型 通道进行的操作
	Addr string `json:"addr"` //目标地址
}

func (this *Message) SetData(data []byte) *Message {
	this.Data = base64.StdEncoding.EncodeToString(data)
	return this
}

func (this *Message) SetType(Type Type) *Message {
	this.Type = Type
	return this
}

func (this *Message) String() string {
	return string(this.Bytes())
}

func (this *Message) Bytes() []byte {
	bs, _ := json.Marshal(this)
	return bs
}

func DecodeMsg(bytes []byte) (*Message, error) {
	m := new(Message)
	err := json.Unmarshal(bytes, m)
	if err != nil {
		return nil, err
	}
	bs, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil {
		return nil, err
	}
	m.Data = string(bs)
	return m, nil
}

func NewCloseMsg(key, data string) *Message {
	m := &Message{
		Key:  key,
		Type: Close,
	}
	return m.SetData([]byte(data))
}

func NewConnectMsg(key, addr string) *Message {
	return &Message{
		Key:  key,
		Addr: addr,
		Type: Connect,
	}
}

func NewWriteMsg(key, addr string, data []byte) *Message {
	m := &Message{
		Key:  key,
		Addr: addr,
		Type: Write,
	}
	return m.SetData(data)
}

func NewRegisterMsg(key, data string) *Message {
	m := &Message{
		Key:  key,
		Type: Register,
	}
	return m.SetData([]byte(data))
}

func NewInfoMsg(key string, data []byte) *Message {
	m := &Message{
		Key:  key,
		Type: Info,
	}
	return m.SetData(data)
}
