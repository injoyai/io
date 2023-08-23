package proxy

import (
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"regexp"
)

//func NewPortForwardingClient(addr, sn string, options ...func(c *io.Client, e *Entity)) *io.Client {
//	return NewTCPClient(addr, func(c *io.Client, e *Entity) {
//		for _, v := range options {
//			v(c, e)
//		}
//		//注册
//		c.Write(NewRegisterMessage(sn, sn).Bytes())
//	})
//}

type PortForwardingServer struct {
	*io.Server
	listen *maps.Safe
}

// Listen 监听
func (this *PortForwardingServer) Listen(port int, sn, addr string, options ...io.OptionServer) error {
	s, err := dial.NewTCPServer(port, func(s *io.Server) {
		s.Tag.Set("sn", sn)
		s.Tag.Set("addr", addr)
		s.SetDealFunc(func(msg *io.IMessage) {
			{
				if _addr := regexp.MustCompile("(\\?|&)(_addr=)[0-9.:]+").FindString(msg.String()); len(_addr) > 7 {
					s.Tag.Set("addr", _addr[7:])
				}
				if _sn := regexp.MustCompile("(\\?|&)(_sn=)[0-9a-zA-Z]+").FindString(msg.String()); len(_sn) > 5 {
					s.Tag.Set("sn", _sn[5:])
				}
			}
			sn = s.Tag.GetString("sn")
			addr = s.Tag.GetString("addr")
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
	if err != nil {
		return err
	}
	s.SetOptions(options...)
	this.listen.Set(conv.String(port), s)
	go s.Run()
	return nil
}

//// NewPortForwardingServer 端口转发服务端
//func NewPortForwardingServer(port int, options ...io.OptionServer) (*PortForwardingServer, error) {
//	pipeServer, err := dial.NewPipeServer(port, options...)
//	if err != nil {
//		return nil, err
//	}
//	ser := &PortForwardingServer{Server: pipeServer, listen: maps.NewSafe()}
//	pipeServer.SetCloseFunc(func(msg *io.IMessage) {
//		//通道断开连接,则关闭所有代理连接
//		sn := msg.GetKey()
//		ser.listen.Range(func(key, value interface{}) bool {
//			s := value.(*io.Server)
//			if s.Tag.GetString("sn") == sn {
//				s.CloseClientAll()
//			}
//			return true
//		})
//	})
//	pipeServer.SetDealFunc(func(msg *io.IMessage) {
//		m, err := DecodeMessage(msg.Bytes())
//		if err != nil {
//			//msg.Close()
//			logs.Err(err)
//			return
//		}
//		switch m.OperateType {
//		case Register:
//			//处理注册信息
//			pipeServer.SetClientKey(msg.Client, m.Data)
//		default:
//			listenPort := strings.Split(m.Key, "#")[0]
//			v := ser.listen.MustGet(listenPort)
//			if v == nil {
//				msg.WriteAny(m.Close(fmt.Sprintf("未找到监听服务(%s)", listenPort)))
//				return
//			}
//			s := v.(*io.Server)
//			m.Key = append(strings.Split(m.Key, "#"), "")[1]
//			switch m.OperateType {
//			case Response:
//				//代理响应,写入请求客户端
//				_, err = s.WriteClient(m.Key, m.GetData())
//			case Close:
//				//关闭客户端请求连接
//				c := s.GetClient(m.Key)
//				if c != nil {
//					_ = c.TryCloseWithDeadline()
//				}
//			default:
//				//代理请求 改服务作为代理 发起请求
//
//			}
//		}
//	})
//	return ser, nil
//}
