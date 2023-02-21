package pipe

import (
	"github.com/injoyai/io"
)

// NewServer 定义了读写规则的io服务
func NewServer(listen io.ListenFunc) (*io.Server, error) {
	server, err := io.NewServer(listen)
	if err != nil {
		return nil, err
	}
	server.SetWriteFunc(DefaultWriteFunc)
	server.SetReadFunc(DefaultReadFunc)
	return server, nil
}

// NewTransmit 通过客户端数据转发,例如客户端1的数据会广播其他所有客户端
func NewTransmit(listen io.ListenFunc) (*io.Server, error) {
	server, err := io.NewServer(listen)
	if err != nil {
		return nil, err
	}
	server.SetWriteFunc(DefaultWriteFunc)
	server.SetReadFunc(DefaultReadFunc)
	server.SetDealFunc(func(msg *io.IMessage) {
		for _, v := range server.GetClientMap() {
			if v.GetKey() != msg.GetKey() {
				v.Write(msg.Bytes())
			}
		}
	})
	return server, nil
}
