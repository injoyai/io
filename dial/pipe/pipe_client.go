package pipe

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

func Redial(dial io.DialFunc, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(dial, func(ctx context.Context, c *io.Client) {
		c.SetWriteFunc(DefaultWriteFunc)
		c.SetReadFunc(DefaultReadFunc)
		for _, v := range fn {
			v(ctx, c)
		}
		c.SetKeepAlive(io.Timeout)
	})
}

// NewClient 新建管道客户端
func NewClient(dial io.DialFunc) (*io.Client, error) {
	c, err := io.NewDial(dial)
	if err != nil {
		return nil, err
	}
	c.SetWriteFunc(DefaultWriteFunc)
	c.SetReadFunc(DefaultReadFunc)
	c.Redial()
	return c, nil
}
