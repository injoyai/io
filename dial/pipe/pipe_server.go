package pipe

import (
	"github.com/injoyai/io"
)

func NewServer(listen io.ListenFunc) (*Server, error) {
	server, err := io.NewServer(listen)
	if err != nil {
		return nil, err
	}
	s := &Server{Server: server}
	server.SetWriteFunc(DefaultWriteFunc)
	server.SetReadFunc(DefaultReadFunc)
	return s, nil
}

// Server 定义了读写规则的io服务
type Server struct {
	*io.Server
}

func NewTransmit(listen io.ListenFunc) (*Server, error) {
	server, err := io.NewServer(listen)
	if err != nil {
		return nil, err
	}
	s := &Server{Server: server}
	server.SetWriteFunc(DefaultWriteFunc)
	server.SetReadFunc(DefaultReadFunc)
	server.SetDealFunc(func(msg *io.IMessage) {
		for _, v := range server.GetClientMap() {
			if v.GetKey() != msg.GetKey() {
				v.Write(msg.Bytes())
			}
		}
	})
	return s, nil
}

// Transmit 数据转发 单一服务
// 收到的数据会转发一份到所有客户端
type Transmit struct {
	*Server
}
