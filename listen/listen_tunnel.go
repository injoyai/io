package listen

import (
	"context"
	"errors"
	"github.com/injoyai/base/g"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
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

func NewTunnelClient(s *io.Server, tunDial io.DialFunc, proxyAddr string, options ...io.OptionClient) {
	pool := io.NewPool(tunDial, options...)
	s.ClientManage.SetOptions(func(c *io.Client) {
		c.SetConnectFunc(func(client *io.Client) error {
			tun, err := pool.Get()
			if err != nil {
				return err
			}

			{
				client.SetReadWith1KB()
				client.SetDealFunc(func(c *io.Client, msg io.Message) {
					//写入数据
					tun.Write(append([]byte{TypeWrite, TypeRequest}, msg.Bytes()...))
				})
				client.SetCloseFunc(func(ctx context.Context, c *io.Client, err error) {
					//发送关闭信息
					tun.Write(append([]byte{TypeClose, TypeRequest}, conv.Bytes(err)...))
					//放回连接池
					pool.Put(tun)
				})
			}

			{
				tun.SetKey(client.GetKey())
				tun.SetReadWriteWithPkg()
				tun.SetCloseWithCloser(client)
				tun.SetDealFunc(func(c *io.Client, msg io.Message) {
					m, err := decodeTunnelMessage(msg.Bytes())
					if err != nil {
						tun.Write(append([]byte{TypeClose, TypeRequest}, []byte(err.Error())...))
						client.CloseWithErr(errors.New(err.Error()))
						return
					}
					switch m.Type {
					case TypeConnect:
						//连接成功响应
					case TypeWrite:
						//写入数据
						client.Write(m.Data)
					default: //TypeClose
						client.CloseWithErr(errors.New(string(m.Data)))
					}
				})

				//发送连接信息
				_, err = tun.WriteRead(append([]byte{TypeConnect, TypeRequest}, []byte(proxyAddr)...), io.DefaultConnectTimeout)
				if err != nil {
					return err
				}
			}

			return nil
		})
	})
}

func NewTunnelServer(s *io.Server) {
	s.ClientManage.SetOptions(func(c *io.Client) {
		c.SetConnectFunc(func(tun *io.Client) error {
			var c *io.Client
			tun.SetReadWriteWithPkg()
			tun.SetDealFunc(func(c2 *io.Client, msg io.Message) {
				m, err := decodeTunnelMessage(msg.Bytes())
				if err == nil {
					switch m.Type {
					case TypeConnect:
						switch m.Model {
						default:
							c, err = dial.NewTCP(string(m.Data), func(c *io.Client) {
								c.Debug(false)
								c.SetReadWith1KB()
								c.SetDealFunc(func(c *io.Client, msg io.Message) {
									//写入数据
									tun.Write(append([]byte{TypeWrite, TypeRequest}, msg.Bytes()...))
								})
								c.SetCloseFunc(func(ctx context.Context, c *io.Client, err error) {
									//发送关闭信息
									tun.Write(append([]byte{TypeClose, TypeRequest}, conv.Bytes(err)...))
								})
								tun.SetCloseWithCloser(c)
							})
							if err != nil {
								//发送关闭信息
								tun.Write(append([]byte{TypeClose, TypeRequest}, []byte(err.Error())...))
								return
							}
							tun.Write(append([]byte{TypeConnect, TypeResponse}, []byte("连接成功")...))
							go c.Run()
						}
					case TypeWrite:
						if c == nil || c.Closed() {
							//发送关闭信息
							tun.Write(append([]byte{TypeClose, TypeRequest}, []byte("无连接")...))
							return
						}
						//写入数据
						c.Write(m.Data)
					case TypeClose:
						//关闭连接
						if c != nil {
							c.Close()
						}

					}
				}
			})
			return nil
		})
	})
}
