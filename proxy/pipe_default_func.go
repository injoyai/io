package proxy

import (
	"bufio"
	"encoding/base64"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
	"github.com/injoyai/io/dial"
	"log"
	"time"
)

var (
	WriteFunc   = DefaultWrite
	ReadFunc    = DefaultRead
	PrintFunc   = DefaultPrint
	ConnectFunc = DefaultConnect
)

var (
	defaultStart    = []byte{0x03, 0x03}
	defaultEnd      = []byte{0x04, 0x04}
	defaultReadFunc = buf.NewReadWithStartEnd(defaultStart, defaultEnd)
)

// DefaultWrite 数据通讯加密,通道客户端服务端的发送数据函数
func DefaultWrite(req []byte) []byte {
	req = []byte(base64.StdEncoding.EncodeToString(req))
	req = append(append(defaultStart, req...), defaultEnd...)
	return req
}

// DefaultRead 数据通讯解密,通道客户端服务端的读取数据函数
func DefaultRead(buf *bufio.Reader) (bytes []byte, err error) {
	bytes, err = defaultReadFunc(buf)
	if err != nil {
		return nil, err
	}
	if len(bytes) > len(defaultStart)+len(defaultEnd) {
		bytes = bytes[len(defaultStart) : len(bytes)-len(defaultEnd)]
	}
	return base64.StdEncoding.DecodeString(string(bytes))
}

// DefaultPrint 通道服务端打印函数
func DefaultPrint(tag, key string, msg io.Message) {
	log.Printf("[pipe][%s][%s] %s\n", tag, key, msg.String())
}

// DefaultConnect 默认连接函数
func DefaultConnect(i io.Writer) func(key, addr string) (cli io.ReadWriteCloser, err error) {
	return func(key, addr string) (io.ReadWriteCloser, error) {
		c, err := io.NewDial(dial.TCPFunc(addr))
		if err == nil {
			log.Printf("[连接][pipe][%s] 连接服务成功...", addr)
			c.SetPrintFunc(func(tag, key string, msg io.Message) {
				log.Printf("[%s][pipe][%s] %s\n", tag, key, msg.String())
			})
			c.SetTimeout(time.Second * 10)
			c.SetKey(key)
			c.Debug(false)
			c.SetWriteFunc(nil)
			c.SetReadFunc(buf.ReadWithAll)
			c.SetDealFunc(func(msg *io.ClientMessage) {
				i.Write(NewWriteMsg(key, addr, msg.Bytes()).Bytes())
			})
			c.SetCloseFunc(func(msg *io.ClientMessage) {
				i.Write(NewCloseMsg(key, msg.String()).Bytes())
				log.Printf("[错误][%s] %s", msg.GetKey(), msg.String())
			})
		}
		return c, err
	}
}
