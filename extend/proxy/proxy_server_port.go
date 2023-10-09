package proxy

import (
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
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
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Tag().Set("sn", sn)
		s.Tag().Set("addr", addr)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			{
				if _addr := regexp.MustCompile("(\\?|&)(_addr=)[0-9.:]+").FindString(msg.String()); len(_addr) > 7 {
					s.Tag().Set("addr", _addr[7:])
				}
				if _sn := regexp.MustCompile("(\\?|&)(_sn=)[0-9a-zA-Z]+").FindString(msg.String()); len(_sn) > 5 {
					s.Tag().Set("sn", _sn[5:])
				}
			}
			sn = s.Tag().GetString("sn")
			addr = s.Tag().GetString("addr")
			pipe := this.GetClient(sn)
			if pipe == nil {
				c.CloseWithErr(fmt.Errorf("通道客户端未连接,关闭连接"))
				return
			}
			key := fmt.Sprintf("%d#%s", port, c.GetKey())
			if _, err := pipe.WriteAny(NewWriteMessage(key, addr, msg.Bytes())); err != nil {
				c.Close()
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
