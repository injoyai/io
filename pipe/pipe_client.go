package pipe

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"log"
	"sync"
)

// NewClient 新建管道客户端
func NewClient(dial io.DialFunc) (*Client, error) {
	client, err := io.NewDial(dial)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Client:   client,
		writeLen: 2 << 10,
		mapBuff:  make(map[string]io.ReadWriteCloser),
		mu:       sync.RWMutex{},
	}
	c.init()
	return c, nil
}

/*
Client
抽象管道概念
例如 用485线通讯,正常的TCP连接 都属于管道
需要 客户端对客户端 客户端对服务端 2种方式
需要 一个管道通讯多个io数据,并且不能长期占用(交替使用,伪多线程)

提供io.Reader io.Writer接口
写入数据会封装一层(封装连接信息,动作,数据)

*/
type Client struct {

	//io客户端
	*io.Client

	//分包长度,每次写入固定字节长度,默认1k
	//0表示一次性全部写入
	//数据太多可能会影响到其他数据的时效
	writeLen uint

	//数据缓存,等到数据被读取
	//单通道支持不同连接的数据
	mapBuff map[string]io.ReadWriteCloser

	mu sync.RWMutex
}

func (this *Client) SetBuff() {

}

// init 初始化操作
func (this *Client) init() {
	this.Client.Redial(func(ctx context.Context, c *io.Client) {

		this.Client.Debug()
		this.Client.SetPrintWithHEX()
		this.Client.SetReadFunc(defaultReadFunc)
		this.Client.SetDealFunc(func(msg *io.ClientMessage) {
			bytes, err := decodePackage(msg.Bytes())
			if err != nil {
				log.Println("[错误]", err)
				return
			}
			result, err := decodeMessage(bytes)
			if err != nil {
				log.Println("[错误]", err)
				return
			}

			//写入数据到缓存
			this.mu.RLock()
			writeCloser := this.mapBuff[result.Key]
			this.mu.RUnlock()
			if writeCloser != nil {

				//如果远程已关闭,则断开当前连接,并删除
				if string(result.Data) == io.EOF.Error() {
					writeCloser.Close()
					this.mu.Lock()
					delete(this.mapBuff, result.Key)
					this.mu.Unlock()
					return
				}

				//存在则写入数据
				writeCloser.Write(result.Data)
				return

			}

			//不存在则说明连接已经关闭,或者远程已关闭
			//下发关闭数据到通道另一头
			m := newCloseMessage(result.Key, "连接已关闭")
			this.Client.Write(encodePackage(m.Bytes()))

		})
	})
}

// Write 实现io.Writer
func (this *Client) Write(p []byte) (int, error) {

	key := "key"
	addr := "addr"
	total := len(p)

	//一次性发送全部数据,数据太多可能会影响到其他数据的时效
	if this.writeLen == 0 {
		msg := newWriteMessage(key, addr, p)
		return this.Client.Write(encodePackage(msg.Bytes()))
	}

	//分包发送,避免其他数据不能及时发送
	for len(p) > 0 {
		data := []byte(nil)
		if len(p) > int(this.writeLen) {
			data = p[:this.writeLen]
			p = p[this.writeLen:]
		} else {
			data = p[:]
			p = p[:0]
		}
		msg := newWriteMessage(key, addr, data)
		if _, err := this.Client.Write(encodePackage(msg.Bytes())); err != nil {
			return 0, err
		}
	}

	return total, nil
}

//========================Message========================

var (
	defaultStart    = []byte{0x03, 0x03}
	defaultEnd      = []byte{0x04, 0x04}
	defaultReadFunc = buf.NewReadWithStartEnd(defaultStart, defaultEnd)
)

func encodePackage(req []byte) []byte {
	req = []byte(base64.StdEncoding.EncodeToString(req))
	req = append(append(defaultStart, req...), defaultEnd...)
	return req
}

func decodePackage(req []byte) ([]byte, error) {
	if len(req) > len(defaultStart)+len(defaultEnd) {
		req = req[len(defaultStart) : len(req)-len(defaultEnd)]
	}
	return base64.StdEncoding.DecodeString(string(req))
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
