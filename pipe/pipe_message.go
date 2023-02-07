package pipe

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"github.com/injoyai/io/buf"
)

var (
	defaultStart    = []byte{0x03, 0x03}
	defaultEnd      = []byte{0x04, 0x04}
	defaultReadFunc = buf.NewReadWithStartEnd(defaultStart, defaultEnd)
)

func DefaultReadFunc(buf *bufio.Reader) ([]byte, error) {
	req, err := defaultReadFunc(buf)
	if err != nil {
		return nil, err
	}
	if len(req) > len(defaultStart)+len(defaultEnd) {
		req = req[len(defaultStart) : len(req)-len(defaultEnd)]
	}
	return base64.StdEncoding.DecodeString(string(req))
}

func DefaultWriteFunc(req []byte) []byte {
	req = []byte(base64.StdEncoding.EncodeToString(req))
	req = append(append(defaultStart, req...), defaultEnd...)
	return req
}

const (
	TypeConnect Type = "connect" //代理,通讯,建立新的链接
	TypeWrite   Type = "write"   //代理,透传,写数据
	TypeClose   Type = "close"   //代理,通讯,关闭链接

	TypeRegister Type = "register" //通讯,注册
	TypeInfo     Type = "info"     //通讯,和服务端建立通讯
)

type Type string

func decodeMessage(bytes []byte) (*Message, error) {
	m := new(Message)
	err := json.Unmarshal(bytes, m)
	return m, err
}

func newConnectMessage(key, addr string) *Message {
	return &Message{
		Type: TypeConnect,
		Key:  key,
		Addr: addr,
	}
}

func newWriteMessage(key, addr string, data []byte) *Message {
	return &Message{
		Type: TypeWrite,
		Key:  key,
		Addr: addr,
		Data: data,
	}
}

func newCloseMessage(key string, data string) *Message {
	return &Message{
		Type: TypeClose,
		Key:  key,
		Data: []byte(data),
	}
}

// Message 内置消息结构
type Message struct {
	Type Type   //动作类型 建立连接,写入数据,关闭连接
	Key  string //标识
	Addr string //目标地址
	Data []byte //写入的数据(如果有)
}

func (this *Message) String() string {
	return string(this.Bytes())
}

func (this *Message) Bytes() []byte {
	bs, _ := json.Marshal(this)
	return bs
}
