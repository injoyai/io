package pipe

import (
	"github.com/injoyai/io"
	"log"
)

func NewServer(listen io.ListenFunc) (*Server, error) {
	server, err := io.NewServer(listen)
	if err != nil {
		return nil, err
	}
	s := &Server{
		Server:   server,
		writeLen: 2 << 10,
		dealFunc: func(msg *Message) error {
			log.Printf("[Server]%s\n", string(msg.Data))
			return nil
		},
	}
	server.SetWriteFunc(DefaultWriteFunc)
	server.SetReadFunc(DefaultReadFunc)
	server.SetDealFunc(newDealFunc(s.dealFunc))
	return s, nil
}

type Server struct {

	//io服务端
	*io.Server

	//分包长度,每次写入固定字节长度,默认1k
	//0表示一次性全部写入
	//数据太多可能会影响到其他数据的时效
	writeLen uint

	//处理数据函数
	dealFunc func(msg *Message) error
}

func (this *Server) SetDealFunc(fn func(msg *Message) error) *Server {
	this.dealFunc = fn
	return this
}

func (this *Server) Write(key, addr string, p []byte) (int, error) {
	return writeFunc(this.writeLen, func(p []byte) (int, error) {
		_, err := this.Server.WriteClient(key, p)
		return len(p), err
	}, key, addr, p)
}

func (this *Server) WriteMessage(msg *Message) error {
	_, err := this.Server.WriteClient(msg.Key, msg.Bytes())
	return err
}
