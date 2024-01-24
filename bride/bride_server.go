package bride

import (
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"net"
	"strings"
)

type Server struct {
	Listener     map[string]map[string]*io.Server
	bridge       *io.Server //用于请求转发数据
	bridgeClient *maps.Safe
}

func (this *Server) Listen(Type, port string, options ...io.OptionServer) (*io.Server, error) {
	var listener io.ListenFunc
	switch Type {
	case io.UDP:
		listener = listen.WithUDP(conv.Int(port))
	default:
		listener = listen.WithTCP(conv.Int(port))
	}
	return io.NewServer(listener, func(s *io.Server) {
		s.SetOptions(options...)
		s.SetBeforeFunc(func(c *io.Client) error {
			conn := c.ReadWriteCloser().(net.Conn)
			addr := conn.RemoteAddr().String()
			list := strings.Split(addr, ":")
			port := list[len(list)-1]
			ip := net.ParseIP(strings.Join(list[:len(list)-1], ":"))
			c.Tag().Set("ip", ip)
			c.Tag().Set("port", port)
			return nil
		})
		s.SetCloseFunc(func(c *io.Client, msg io.Message) {
			this.bridgeClient.Del(c.GetKey())
		})
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			list := io.NewFrameBridgeBytes(
				c.Tag().GetBytes("ip"),
				c.Tag().GetUint16("port"),
				msg,
			)
			for _, bs := range list {
				val, _ := this.bridgeClient.GetOrSetByHandler(Type+"."+port, func() (interface{}, error) {
					return []*io.Client(nil), nil
				})
				for _, v := range val.([]*io.Client) {
					v.Write(bs)
				}
			}
		})
		go s.Run()
	})
}

func (this *Server) Get(Type, port string) *io.Server {
	m, ok := this.Listener[Type]
	if ok {
		return m[port]
	}
	return nil
}

func NewServer(port int, options ...io.OptionServer) (*Server, error) {
	bridge, err := listen.NewTCPServer(port, options...)
	if err != nil {
		return nil, err
	}
	ser := &Server{bridge: bridge}
	ser.bridge.SetBeforeFunc(func(c *io.Client) error {
		return nil
	})
	ser.bridge.SetDealFunc(func(c2 *io.Client, msg io.Message) {
		p, err := io.DecodeBridge(msg)
		if err != nil {
			ser.bridge.Logger.Errorf("decode bridge error:%v", err)
			return
		}
		l := ser.Get(io.TCP, conv.String(p.ListenPort))
		if l == nil {
			//todo 不存在怎么处理
			return
		}
		c := l.GetClient(p.Address())
		if c == nil {
			//todo 不存在怎么处理
			return
		}
		if _, err := c.Write(p.Data); err != nil {
			//todo 错误怎么处理
			return
		}
	})
	return ser, nil
}
