package pipe

import (
	"github.com/injoyai/io"
)

// NewServer 定义了读写规则的io.Server服务
func NewServer(listen io.ListenFunc, fn ...func(s *io.Server)) (*io.Server, error) {
	return io.NewServer(listen, append([]func(s *io.Server){WithServer}, fn...)...)
}

// NewTransmit 通过客户端数据转发,例如客户端1的数据会广播其他所有客户端
func NewTransmit(listen io.ListenFunc, fn ...func(s *io.Server)) (*io.Server, error) {
	return io.NewServer(listen, func(s *io.Server) {
		s.SetWriteFunc(DefaultWriteFunc)
		s.SetReadFunc(DefaultReadFunc)
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
		for _, v := range fn {
			v(s)
		}
	})
}
