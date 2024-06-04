package p2p

import (
	"context"
	"encoding/json"
	"github.com/injoyai/base/g"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"net"
	"time"
)

var (
	StartTime = time.Now()
)

const (
	Version = "1.0.0"

	CodeNotFind = 403
	CodeSuccess = 200

	TypeError            = "error"            //错误信息
	TypeRegisterReq      = "registerReq"      //注册自己节点信息
	TypeRegisterRes      = "registerRes"      //注册自己节点信息
	TypeGetRegisterReq   = "getRegisterReq"   //获取其他节点信息
	TypeGetRegisterRes   = "getRegisterRes"   //获取节点信息响应
	TypeConnectReq       = "connectReq"       //连接其他节点
	TypeConnectNoticeReq = "connectNotice"    //连接其他节点通知
	TypeConnectNoticeRes = "connectNoticeRes" //连接其他节点通知
)

type Peer interface {
	LocalAddr() net.Addr    //本地地址
	Ping(addr string) error //ping下地址,如果协议一直,则有消息返回

}

func NewPeer(port int, options ...io.OptionServer) (p *peer, err error) {
	p = &peer{
		localAddr: &net.UDPAddr{Port: port},
		clients:   maps.NewSafe(),
		wait:      wait.New(time.Second * 2),
	}
	p.Server, err = listen.NewUDPServer(port, func(s *io.Server) {
		s.SetReadFunc(io.ReadWithPkg)
		s.SetWriteFunc(io.WriteWithPkg)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {

			m := new(Msg)
			json.Unmarshal(msg.Bytes(), m)

			switch m.Type {

			case TypeError:
				//响应的错误信息
				errMsg := new(MsgError)
				json.Unmarshal(conv.Bytes(m.Data), errMsg)
				switch errMsg.Code {
				case CodeNotFind:
					s.Tag().Del(TypeRegisterReq + "_" + conv.String(errMsg.Data))
				}
				p.wait.Done(m.MsgID, errMsg.Msg)

			case TypeRegisterReq:
				//上报注册信息
				registerMsg := new(MsgRegister)
				json.Unmarshal(conv.Bytes(m.Data), registerMsg)
				registerMsg.RemoteAddr = c.GetKey()
				c.SetKey(registerMsg.NodeID)
				c.Tag().Set("register", registerMsg)
				s.Tag().Set(TypeRegisterReq+"_"+registerMsg.NodeID, registerMsg)

			case TypeRegisterRes:
				//响应注册成功
				p.wait.Done(m.MsgID, nil)

			case TypeGetRegisterReq:
				//上报获取注册信息
				getRegisterMsg := new(MsgGetRegister)
				json.Unmarshal(conv.Bytes(m.Data), getRegisterMsg)
				registerMsg := new(MsgRegister)
				v, ok := s.Tag().Get(getRegisterMsg.NodeID)
				if ok {
					registerMsg = v.(*MsgRegister)
				}
				c.WriteAny(Msg{
					Type:  TypeRegisterRes,
					MsgID: m.MsgID,
					Data:  registerMsg,
				})

			case TypeGetRegisterRes:
				//响应注册信息成功
				registerMsg := new(MsgRegister)
				json.Unmarshal(conv.Bytes(m.Data), registerMsg)
				p.wait.Done(m.MsgID, registerMsg)

			case TypeConnectNoticeReq:
				//连接其他节点
				connectMsg := new(MsgConnectNotice)
				json.Unmarshal(conv.Bytes(m.Data), connectMsg)
				nextPeer := s.GetClient(connectMsg.NodeID)
				if nextPeer == nil {
					c.WriteAny(Msg{
						Type:  TypeError,
						MsgID: m.MsgID,
						Data:  CodeNotFind,
					})
					return
				}
				uuid := g.UUID()
				nextPeer.WriteAny(Msg{
					Type:  TypeConnectReq,
					MsgID: uuid,
					Data:  c.Tag().Get("register"),
				})
				_, err := p.wait.Wait(uuid)
				if err != nil {
					c.WriteAny(Msg{
						Type:  TypeError,
						MsgID: m.MsgID,
						Data: MsgError{
							Code: 0,
							Data: nil,
							Msg:  "",
						},
					})
				} else {
					c.WriteAny(Msg{
						Type:  TypeConnectNoticeRes,
						MsgID: m.MsgID,
					})
				}

			case TypeConnectNoticeRes:
				p.wait.Done(m.MsgID, nil)

			case TypeConnectReq:
				//连接其他节点

			}

		})
		s.SetOptions(options...)
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

type peer struct {
	*io.Server
	NodeID    string
	localAddr *net.UDPAddr
	clients   *maps.Safe
	nat       *maps.Safe
	wait      *wait.Entity
}

func (this *peer) getClient(addr string) (*io.Client, error) {
	return this.GetClientOrDial(addr, func(ctx context.Context) (io.ReadWriteCloser, string, error) {
		c, err := this.Listener().(*listen.UDPServer).NewUDPClient(addr)
		return c, addr, err
	})
}

func (this *peer) WriteTo(addr string, p []byte) (int, error) {
	c, err := this.getClient(addr)
	if err != nil {
		return 0, err
	}
	return c.Write(p)
}

func (this *peer) Ping(addr string, timeout ...time.Duration) error {
	c, err := this.getClient(addr)
	if err != nil {
		return err
	}
	return c.Ping(timeout...)
}

// Register 向服务端注册节点信息
func (this *peer) Register(addr string) error {
	c, err := this.getClient(addr)
	if err != nil {
		return err
	}

	uuid := g.UUID()
	_, err = c.WriteAny(Msg{
		Type:  TypeRegisterReq,
		MsgID: uuid,
		Data: MsgRegister{
			NodeID:     this.NodeID,
			Version:    Version,
			StartTime:  StartTime.Unix(),
			ConnectKey: "",
			LocalAddr:  this.localAddr.String(),
		},
	})

	_, err = this.wait.Wait(uuid)
	return err
}

func (this *peer) GetRegister(addr string, nodeID string) (*MsgRegister, error) {
	c, err := this.getClient(addr)
	if err != nil {
		return nil, err
	}

	uuid := g.UUID()
	_, err = c.WriteAny(Msg{
		Type:  TypeGetRegisterReq,
		MsgID: uuid,
		Data: MsgGetRegister{
			NodeID: nodeID,
		},
	})
	if err != nil {
		return nil, err
	}

	res, err := wait.Wait(uuid)
	if err != nil {
		return nil, err
	}

	return res.(*MsgRegister), nil
}

func (this *peer) Connect() {

}

func (this *peer) LocalAddr() net.Addr {
	return this.localAddr
}
