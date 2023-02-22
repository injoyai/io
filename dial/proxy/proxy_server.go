package proxy

import (
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"net/http"
	"net/url"
	"strings"
)

const (
	KeyAddr    = "addr"
	Connection = "HTTP/1.1 200 Connection established\r\n\r\n"
)

type Server struct {
	s        *io.Server                   //监听服务
	e        *Entity                      //代理实例,正向,反向
	dealFunc func(msg *io.IMessage) error //处理函数
}

func (this *Server) Debug() *Server {
	this.s.Debug()
	return this
}

func (this *Server) Run() error {
	return this.s.Run()
}

func (this *Server) Write(p []byte) (int, error) {
	m, err := DecodeMessage(p)
	if err != nil {
		return 0, err
	}
	switch m.OperateType {
	case Response:
		//代理响应
		_, err = this.s.WriteClient(m.Key, m.GetData())
		return len(p), err
	case Close:
		//关闭请求连接
		this.s.CloseClient(m.Key)
		//关闭代理连接
		return len(p), this.e.WriteMessage(m)
	default:
		//代理请求 逻辑
		return len(p), this.e.WriteMessage(m)
	}
}

func (this *Server) SetDealFunc(fn func(msg *io.IMessage) error) {
	this.dealFunc = fn
}

func (this *Server) SetPrintFunc(fn func(msg io.Message, tag ...string)) {
	this.s.SetPrintFunc(fn)
}

func NewServer(port int, fn ...func(s *Server)) (*Server, error) {
	s, err := dial.NewTCPServer(port, func(s *io.Server) {
		//读取全部数据
		s.SetReadWithAll()
		//设置打印函数
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"PR|S"}, tag...)...)
		})

	})
	if err != nil {
		return nil, err
	}
	ser := &Server{s: s, e: New(), dealFunc: func(msg *io.IMessage) error {
		s.Print([]byte("未设置处理函数"), "PR|S", io.TagErr)
		return msg.Close()
	}}
	// 设置处理数据函数
	// 处理监听到的用户数据,只能监听http协议数据
	// 处理http的CONNECT数据,及处理端口等
	s.SetDealFunc(func(msg *io.IMessage) {
		// HTTP 请求
		list := strings.Split(msg.String(), " ")
		switch true {
		case len(list) > 2 && strings.Contains(list[2], "HTTP"):
			//http请求,头包
			switch list[0] {
			case http.MethodConnect:
				//响应连接成功给请求连接
				msg.Client.WriteString(Connection)
			default:
				u, err := url.Parse(list[1])
				if err == nil {
					port := u.Port()
					if len(port) == 0 {
						switch strings.ToLower(u.Scheme) {
						case "https":
							port = "443"
						default:
							port = "80"
						}
					}
					addr := fmt.Sprintf("%s:%s", u.Hostname(), port)
					msg.Tag().Set(KeyAddr, addr)
					bytes := NewWriteMessage(msg.GetKey(), addr, msg.Bytes()).Bytes()
					if ser.dealFunc != nil {
						msg.CloseWithErr(ser.dealFunc(io.NewIMessage(msg.Client, bytes)))
					}
				}
			}
		default:
			//后续的包
			addr := msg.Tag().GetString(KeyAddr)
			bytes := NewWriteMessage(msg.GetKey(), addr, msg.Bytes()).Bytes()
			if ser.dealFunc != nil {
				msg.CloseWithErr(ser.dealFunc(io.NewIMessage(msg.Client, bytes)))
			}
		}
	})
	for _, v := range fn {
		v(ser)
	}
	return ser, nil
}
