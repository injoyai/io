package dial

import (
	"context"
	"errors"
	"github.com/injoyai/base/g"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
)

/*
Client
抽象管道概念
例如 用485线通讯,正常的TCP连接 都属于管道
需要 客户端对客户端 客户端对服务端 2种方式
需要 一个管道通讯多个io数据,并且不能长期占用 写入前建议分包
只做数据加密(可选),不解析数据,不分包数据

提供io.Reader io.Writer接口
写入数据会封装一层(封装连接信息,动作,数据)
mod
*/

const (
	TypeConnect  = 0x01
	TypeWrite    = 0x02
	TypeClose    = 0x03
	TypeRequest  = 0x00
	TypeResponse = 0x01
)

func decodeTunnelMessage(bs []byte) (*TunnelMessage, error) {
	if len(bs) < 2 {
		return nil, errors.New("数据异常")
	}
	return &TunnelMessage{
		Type:  bs[0],
		Model: bs[1],
		Data:  bs[2:],
	}, nil
}

type TunnelMessage struct {
	Type  uint8  //类型
	Model uint8  //模式
	Data  []byte //数据
}

func (this *TunnelMessage) Bytes() g.Bytes {
	data := []byte(nil)
	data = append(data, this.Type, this.Model)
	data = append(data, this.Data...)
	return data
}

func newTunnelMessageBytes(Type, model uint8, data []byte) []byte {
	return io.NewPkg(0, (&TunnelMessage{
		Type:  Type,
		Model: model,
		Data:  data,
	}).Bytes()).Bytes()
}

func NewTunnelClient(s *io.Server, dial io.DialFunc, options ...io.OptionClient) {
	pool := io.NewPool(dial, func(c *io.Client) { c.SetOptions(options...) })
	s.SetBeforeFunc(func(client *io.Client) error {
		tun, err := pool.Get()
		if err != nil {
			logs.Err(err)
			return err
		}

		{
			tun.SetReadWriteWithPkg()
			tun.SetCloseWithCloser(client)
			tun.SetDealFunc(func(msg *io.IMessage) {
				p, err := io.DecodePkg(msg.Bytes())
				if err == nil {
					m, err := decodeTunnelMessage(p.Data)
					if err == nil {
						switch m.Type {
						case TypeConnect:
							switch m.Model {
							default:
								//无效
							}
						case TypeWrite:
							//写入数据
							client.Write(m.Data)
						case TypeClose:
							//关闭连接
							client.Close()
						}
					}
				}
			})

			logs.Debug(666)
			//发送连接信息
			_, err = tun.WriteRead(append([]byte{TypeConnect, 0}, []byte("127.0.0.1:10001")...), io.DefaultConnectTimeout)
			if err != nil {
				logs.Err(err.Error())
				return err
			}
			logs.Debug(667)
		}

		{
			client.SetReadWithAll()
			client.SetDealFunc(func(msg *io.IMessage) {
				if msg.Len() == 0 {
					return
				}
				//写入数据
				tun.Write(append([]byte{TypeWrite, TypeRequest}, msg.Bytes()...))
			})
			client.SetCloseFunc(func(ctx context.Context, msg *io.IMessage) {
				//发送关闭信息
				tun.WriteRead(append([]byte{TypeClose, TypeRequest}, msg.Bytes()...), io.DefaultResponseTimeout)
				//放回连接池
				pool.Put(tun)
			})
		}

		return nil
	})
}

func NewTunnelServer(s *io.Server) {
	s.SetBeforeFunc(func(tun *io.Client) error {
		tun.SetReadWriteWithPkg()
		tun.SetDealFunc(func(msg *io.IMessage) {
			var c *io.Client
			m, err := decodeTunnelMessage(msg.Bytes())
			if err == nil {
				switch m.Type {
				case TypeConnect:
					switch m.Model {
					default:
						c, err = NewTCP(string(m.Data), func(c *io.Client) {
							c.Debug(true)
							c.SetReadWithAll()
							c.SetDealFunc(func(msg *io.IMessage) {
								//写入数据
								tun.Write(append([]byte{TypeWrite, TypeRequest}, msg.Bytes()...))
							})
							c.SetCloseFunc(func(ctx context.Context, msg *io.IMessage) {
								//发送关闭信息
								tun.WriteRead(append([]byte{TypeClose, TypeRequest}, msg.Bytes()...), io.DefaultResponseTimeout)
							})
							tun.SetCloseWithCloser(c)
						})
						if err != nil {
							logs.Debug("建立连接失败:", err)
							//发送关闭信息
							tun.WriteRead(append([]byte{TypeClose, TypeRequest}, []byte("建立连接失败")...), io.DefaultResponseTimeout)
						} else {
							//响应连接成功
							tun.WriteRead(append([]byte{TypeConnect, TypeResponse}, []byte("succ")...), io.DefaultResponseTimeout)
						}
					}
				case TypeWrite:
					//写入数据
					if c != nil && !c.Closed() {
						c.Write(m.Data)
					} else {
						//发送关闭信息
						tun.WriteRead(newTunnelMessageBytes(TypeClose, TypeRequest, []byte("无连接")), io.DefaultResponseTimeout)
					}
				case TypeClose:
					//关闭连接
					if c != nil && !c.Closed() {
						c.Close()
					}

				}
			}
		})
		return nil
	})
}

//// RedialPipe 通道客户端
//func RedialPipe(addr string, options ...io.OptionClient) *io.Client {
//	return RedialTCP(addr, func(c *io.Client) {
//		c.SetReadWriteWithPkg()
//		c.SetKeepAlive(io.DefaultKeepAlive)
//		c.SetPrintFunc(func(msg io.Message, tag ...string) {
//			io.PrintWithASCII(msg, append([]string{"PI|C"}, tag...)...)
//		})
//		c.SetOptions(options...)
//	})
//}

//// NewPipeServer 通道服务端
//func NewPipeServer(port int, options ...io.OptionServer) (*io.Server, error) {
//	return NewTCPServer(port, func(s *io.Server) {
//		s.SetReadWriteWithPkg()
//		s.SetPrintFunc(func(msg io.Message, tag ...string) {
//			io.PrintWithASCII(msg, append([]string{"PI|S"}, tag...)...)
//		})
//		s.SetOptions(options...)
//	})
//}

//// NewPipeTransmit 通过客户端数据转发,例如客户端1的数据会广播其他所有客户端
//func NewPipeTransmit(port int, options ...io.OptionServer) (*io.Server, error) {
//	return NewPipeServer(port, func(s *io.Server) {
//		s.SetPrintFunc(func(msg io.Message, tag ...string) {
//			if len(tag) > 0 {
//				switch tag[0] {
//				case io.TagWrite, io.TagRead:
//				default:
//					io.PrintWithASCII(msg, append([]string{"PI|T"}, tag...)...)
//				}
//			}
//		})
//		s.SetDealFunc(func(msg *io.IMessage) {
//			//当另一端代理未开启时,无法转发数据
//			for _, v := range s.GetClientMap() {
//				if v.GetKey() != msg.GetKey() {
//					//队列执行,避免阻塞其他
//					v.WriteQueue(msg.Bytes())
//				}
//			}
//		})
//		s.SetOptions(options...)
//	})
//}
