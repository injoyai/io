package proxy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"net/http"
	"net/url"
	"strings"
)

const (
	KeyAddr    = "addr"
	Connection = "HTTP/1.1 200 Connection established\r\n\r\n"
)

type Server struct {
	s        *io.Server               //监听服务
	e        *Entity                  //代理实例,正向,反向
	dealFunc func(msg *Message) error //处理函数
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
	return this.WriteMessage(m)
}

func (this *Server) WriteMessage(m *Message) (int, error) {
	switch m.OperateType {
	case Response:
		//代理响应,写入请求客户端
		_, err := this.s.WriteClient(m.Key, m.GetData())
		return 0, err
	case Close:
		//关闭客户端请求连接
		c := this.s.GetClient(m.Key)
		if c != nil {
			_ = c.TryCloseWithDeadline()
		}
		//关闭代理连接
		return 0, this.e.Close()
	default:
		//代理请求 改服务作为代理 发起请求
		return 0, this.e.WriteMessage(m)
	}
}

func (this *Server) SetDealFunc(fn func(msg *Message) error) {
	this.dealFunc = fn
}

func (this *Server) SetPrintFunc(fn func(msg io.Message, tag ...string)) {
	this.s.SetPrintFunc(fn)
}

func NewServer(dial io.ListenFunc, fn ...func(s *Server)) (*Server, error) {
	ser := &Server{}
	_, err := io.NewServer(dial, func(s *io.Server) {
		//读取全部数据
		s.SetReadWithAll()
		//设置打印函数
		s.SetPrintFunc(func(msg io.Message, tag ...string) {
			io.PrintWithASCII(msg, append([]string{"PR|S"}, tag...)...)
		})
		ser = &Server{s: s, e: New(), dealFunc: func(msg *Message) error {
			m := "未设置处理函数"
			s.Print([]byte("未设置处理函数"), "PR|S", io.TagErr)
			return errors.New(m)
		}}
		s.SetCloseFunc(func(msg *io.IMessage) {
			m := NewCloseMessage(msg.GetKey(), msg.String())
			if ser.dealFunc != nil {
				msg.CloseWithErr(ser.dealFunc(m))
			}
		})
		// 设置处理数据函数
		// 处理监听到的用户数据,只能监听http协议数据
		// 处理http的CONNECT数据,及处理端口等
		s.SetDealFunc(func(msg *io.IMessage) {
			// HTTP 请求
			list := strings.Split(msg.String(), " ")
			switch true {
			case len(list) > 2 && list[0] == http.MethodConnect:
				//http代理请求,例如浏览器
				addr, err := getAddr(msg)
				if err != nil {
					logs.Err(err)
					msg.Close()
					return
				}
				msg.Tag().Set(KeyAddr, addr)
				msg.Client.WriteString(Connection)
			default:
				//后续的包
				addr, err := getAddr(msg)
				if err != nil {
					//这里不做处理
					//logs.Err(err)
					//msg.Close()
					//return
				}
				m := NewWriteMessage(msg.GetKey(), addr, msg.Bytes())
				if ser.dealFunc != nil {
					msg.CloseWithErr(ser.dealFunc(m))
				}
			}
		})
		for _, v := range fn {
			v(ser)
		}
	})
	return ser, err
}

func NewUDPServer(port int, fn ...func(s *Server)) (*Server, error) {
	s, err := NewServer(dial.UDPListenFunc(port), fn...)
	if err == nil {
		s.s.SetKey(fmt.Sprintf(":%d", port))
	}
	return s, err
}

func NewTCPServer(port int, fn ...func(s *Server)) (*Server, error) {
	s, err := NewServer(dial.TCPListenFunc(port), fn...)
	if err == nil {
		s.s.SetKey(fmt.Sprintf(":%d", port))
	}
	return s, err
}

// 获取请求地址
func getAddr(msg *io.IMessage) (string, error) {
	addr := msg.Tag().GetString(KeyAddr)
	if len(addr) > 0 {
		return addr, nil
	}
	//按http尝试解析数据
	r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Bytes())))
	if err != nil {
		return "", err
	}
	addr = r.Host
	hostPortURL, err := url.Parse(r.Host)
	if err == nil {
		if hostPortURL.Opaque == "443" {
			if strings.Index(r.Host, ":") == -1 {
				addr += ":443"
			}
		} else {
			if strings.Index(r.Host, ":") == -1 {
				addr += ":80"
			}
		}
	}
	return addr, nil
}
