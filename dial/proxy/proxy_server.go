package proxy

import (
	"bufio"
	"bytes"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
		c := this.s.GetClient(m.Key)
		if c != nil {
			switch val := c.ReadWriteCloser().(type) {
			case net.Conn:
				val.SetReadDeadline(time.Time{})
				return len(p), nil
			}
		}
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
		case len(list) > 2 && list[0] == http.MethodConnect:
			//http代理请求,例如浏览器
			addr, err := getAddr(msg)
			if err != nil {
				logs.Err(err)
				msg.Close()
				return
			}
			logs.Debug("代理包:", addr)
			msg.Tag().Set(KeyAddr, addr)
			msg.Client.WriteString(Connection)
		default:
			//后续的包
			addr, err := getAddr(msg)
			if err != nil {
				logs.Err(err)
				msg.Close()
				return
			}
			logs.Debug("后续包:", addr)
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
