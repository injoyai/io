package proxy

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

func NewPortForwardingClient(addr, sn string, fn ...func(ctx context.Context, c *io.Client, e *Entity)) *io.Client {
	return NewTCPClient(addr, func(ctx context.Context, c *io.Client, e *Entity) {
		for _, v := range fn {
			v(ctx, c, e)
		}
		//注册
		c.Write(NewRegisterMessage(sn, sn).Bytes())
	})
}

type PortForwardingServer struct {
	*io.Server
	listen *maps.Safe
}

// Listen 监听
func (this *PortForwardingServer) Listen(port int, sn, addr string) error {
	s, err := dial.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
		s.Tag.Set("sn", sn)
		s.Tag.Set("addr", addr)
		s.SetDealFunc(func(msg *io.IMessage) {
			pipe := this.GetClient(sn)
			if pipe == nil {
				msg.Client.CloseWithErr(fmt.Errorf("通道客户端未连接,关闭连接"))
				return
			}
			key := fmt.Sprintf("%d#%s", port, msg.GetKey())
			if _, err := pipe.WriteAny(NewWriteMessage(key, addr, msg.Bytes())); err != nil {
				msg.Client.Close()
			}
		})
	})
	this.listen.Set(conv.String(port), s)
	go s.Run()
	return err
}

// NewPortForwardingServer 端口转发服务端
func NewPortForwardingServer(port int) (*PortForwardingServer, error) {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	pipeServer, err := dial.NewPipeServer(port)
	ser := &PortForwardingServer{Server: pipeServer, listen: maps.NewSafe()}
	pipeServer.Debug()
	pipeServer.SetCloseFunc(func(msg *io.IMessage) {
		//通道断开连接,则关闭所有代理连接
		sn := msg.GetKey()
		ser.listen.Range(func(key, value interface{}) bool {
			s := value.(*io.Server)
			if s.Tag.GetString("sn") == sn {
				s.CloseClientAll()
			}
			return true
		})
	})
	pipeServer.SetDealFunc(func(msg *io.IMessage) {
		m, err := DecodeMessage(msg.Bytes())
		if err != nil {
			//msg.Close()
			return
		}
		switch m.OperateType {
		case Register:
			//处理注册信息
			pipeServer.SetClientKey(msg.Client, m.Data)
		default:
			v := ser.listen.MustGet(strings.Split(m.Key, "#")[0])
			if v == nil {
				msg.WriteAny(m.Close("未找到服务"))
				return
			}
			s := v.(*io.Server)
			m.Key = append(strings.Split(m.Key, "#"), "")[1]
			switch m.OperateType {
			case Response:
				//代理响应,写入请求客户端
				_, err = s.WriteClient(m.Key, m.GetData())
			case Close:
				//关闭客户端请求连接
				c := s.GetClient(m.Key)
				if c != nil {
					_ = c.TryCloseWithDeadline()
				}
			default:
				//代理请求 改服务作为代理 发起请求
				logs.Debug(666)

			}

		}
	})
	return ser, err
}
