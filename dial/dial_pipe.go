package dial

import (
	"context"
	"github.com/injoyai/io"
)

/*
Client
抽象管道概念
例如 用485线通讯,正常的TCP连接 都属于管道
需要 客户端对客户端 客户端对服务端 2种方式
需要 一个管道通讯多个io数据,并且不能长期占用 写入前建议分包
只做数据加密(可选),不解析数据,不分包数据

提供io.Reader io.Writer接口
写入数据会封装一层(封装连接信息,动作,数据)

*/

// RedialPipe 通道客户端
func RedialPipe(addr string, options ...func(ctx context.Context, c *io.Client)) *io.Client {
	return RedialTCP(addr, func(ctx context.Context, c *io.Client) {
		c.SetReadWriteWithPkg()
		c.SetKeepAlive(io.DefaultTimeout)
		c.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"PI|C"}, tag...)...)
		})
		c.SetOptions(options...)
	})
}

// NewPipeServer 通道服务端
func NewPipeServer(port int, options ...func(s *io.Server)) (*io.Server, error) {
	return NewTCPServer(port, func(s *io.Server) {
		s.SetReadWriteWithPkg()
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"PI|S"}, tag...)...)
		})
		s.SetOptions(options...)
	})
}

// NewPipeTransmit 通过客户端数据转发,例如客户端1的数据会广播其他所有客户端
func NewPipeTransmit(port int, options ...func(s *io.Server)) (*io.Server, error) {
	return NewPipeServer(port, func(s *io.Server) {
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			if len(tag) > 0 {
				switch tag[0] {
				case io.TagWrite, io.TagRead:
				default:
					io.PrintWithASCII(msg, append([]string{"PI|T"}, tag...)...)
				}
			}
		})
		s.SetDealFunc(func(msg *io.IMessage) {
			//当另一端代理未开启时,无法转发数据
			for _, v := range s.GetClientMap() {
				if v.GetKey() != msg.GetKey() {
					//队列执行,避免阻塞其他
					v.WriteQueue(msg.Bytes())
				}
			}
		})
		s.SetOptions(options...)
	})
}
