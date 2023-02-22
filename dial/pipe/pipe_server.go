package pipe

import (
	"github.com/injoyai/io"
)

// NewServer 定义了读写规则的io.Server服务
func NewServer(listen io.ListenFunc, fn ...func(s *io.Server)) (*io.Server, error) {
	return io.NewServer(listen, func(s *io.Server) {
		s.SetWriteFunc(DefaultWriteFunc)
		s.SetReadFunc(DefaultReadFunc)
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"PI|S"}, tag...)...)
		})
		for _, v := range fn {
			v(s)
		}
	})
}

// NewTransmit 通过客户端数据转发,例如客户端1的数据会广播其他所有客户端
func NewTransmit(listen io.ListenFunc) (*io.Server, error) {
	return io.NewServer(listen, func(s *io.Server) {
		s.SetWriteFunc(DefaultWriteFunc)
		s.SetReadFunc(DefaultReadFunc)
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"PI|T"}, tag...)...)
		})
		s.SetDealFunc(func(msg *io.IMessage) {
			for _, v := range s.GetClientMap() {
				if v.GetKey() != msg.GetKey() {
					v.Write(msg.Bytes())
				}
			}
		})
	})
}
