package bridge

import (
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"net"
)

type Server struct {
	Listener     *maps.Safe //监听服务
	bridge       *io.Server //用于请求转发数据
	bridgeClient *maps.Safe //订阅客户端管理
}

func (this *Server) Listen(Type, port string, options ...io.OptionServer) (*io.Server, error) {
	listenKey := Type + "." + port
	var listener io.ListenFunc
	switch Type {
	case io.UDP:
		listener = listen.WithUDP(conv.Int(port))
	default:
		listener = listen.WithTCP(conv.Int(port))
	}
	s, err := io.NewServer(listener, func(s *io.Server) {
		s.SetKey(fmt.Sprintf("%s.%s", Type, port))
		s.SetOptions(options...)
		s.SetBeforeFunc(func(c *io.Client) error {
			conn := c.ReadWriteCloser().(net.Conn)
			addr := conn.RemoteAddr().String()
			c.Tag().Set(io.FliedAddress, addr)
			return nil
		})
		s.SetCloseFunc(func(c *io.Client, msg io.Message) {
			this.bridgeClient.Del(c.GetKey())
		})
		msgID := uint8(0)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			//处理客户端上来的数据
			for _, bs := range io.SplitWithLength(msg, 65500) {
				msgID++
				val, _ := this.bridgeClient.GetOrSetByHandler(listenKey, func() (interface{}, error) {
					return []*io.Client(nil), nil
				})
				for _, v := range val.([]*io.Client) {
					//向订阅者发送客户端上来的数据
					v.Write(io.NewSimple(io.SimpleControl{Type: io.OprWrite}, io.SimpleData{
						io.FliedKey:     []byte(listenKey),
						io.FliedAddress: c.Tag().GetBytes(io.FliedAddress),
						io.FliedData:    bs,
					}, msgID).Bytes())
				}
			}
		})
		go s.Run()
	})
	if err == nil {
		this.Listener.Set(listenKey, s)
	}
	return s, err
}

func (this *Server) CloseListen(key string) error {
	l := this.Listener.GetAndDel(key)
	if l != nil {
		return l.(*io.Server).Close()
	}
	return nil
}

func (this *Server) Run() error {
	return this.bridge.Run()
}

func NewServer(port int, options ...io.OptionServer) (*Server, error) {
	bridge, err := listen.NewTCPServer(port, options...)
	if err != nil {
		return nil, err
	}
	bridge.SetPrintWithHEX()
	ser := &Server{
		Listener:     maps.NewSafe(),
		bridge:       bridge,
		bridgeClient: maps.NewSafe(),
	}
	ser.bridge.SetReadWriteWithSimple()
	ser.bridge.SetCloseFunc(func(c *io.Client, msg io.Message) {
		key := c.Tag().GetString("listenKey")
		val, ok := ser.bridgeClient.Get(key)
		if ok {
			for i, v := range val.([]*io.Client) {
				if v == c {
					ser.bridgeClient.Set(key, append(val.([]*io.Client)[:i], val.([]*io.Client)[i+1:]...))
					break
				}
			}
		}
	})
	ser.bridge.SetDealFunc(func(c2 *io.Client, msg io.Message) {
		p, err := io.DecodeSimple(msg)
		if err != nil {
			ser.bridge.Errorf("decode bridge error:%v", err)
			return
		}

		listenType := string(p.Data[FliedListenType]) //监听服务的类型
		listenPort := string(p.Data[FliedListenPort]) //监听服务的端口
		listenKey := listenType + "." + listenPort
		address := string(p.Data[io.FliedAddress]) //客户端的地址
		data := p.Data[io.FliedMsg]                //消息内容

		//判断订阅客户端的消息类型
		switch p.Control.Type {
		case io.OprSubscribe:
			c2.Infof("[%s] 订阅[%s]\n", c2.GetKey(), listenKey)
			c2.Tag().Set(FliedListenType, listenType)
			c2.Tag().Set(FliedListenPort, listenPort)
			c2.Tag().Set(FliedListenKey, listenKey)
			v, _ := ser.bridgeClient.GetOrSetByHandler(listenKey, func() (interface{}, error) {
				return []*io.Client{}, nil
			})
			ser.bridgeClient.Set(listenKey, append(v.([]*io.Client), c2))
			_, err := c2.Write(p.Resp(io.SimpleData{io.FliedCode: conv.Bytes(uint16(200))}).Bytes())
			if err != nil {
				ser.bridge.Errorf("订阅[%s]失败: %v", listenKey, err)
			}

		case io.OprWrite:
			l := ser.Listener.MustGet(listenKey)
			if l == nil {
				_, err := c2.Write(p.Resp(nil, fmt.Errorf("监听服务[%s.%s]未开启", listenType, listenPort)).Bytes())
				logs.PrintErr(err)
				return
			}
			c := l.(*io.Server).GetClient(address)
			if c == nil {
				_, err := c2.Write(p.Resp(nil, fmt.Errorf("客户端[%s]未连接", address)).Bytes())
				logs.PrintErr(err)
				return
			}
			_, err := c.Write(data)
			logs.PrintErr(err)
		}

	})
	return ser, nil
}
