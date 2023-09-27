package p2p

import (
	"encoding/json"
	"github.com/injoyai/base/maps"
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

	TypeRegister = "register"
)

type Peer interface {
	LocalAddr() net.Addr    //本地地址
	Ping(addr string) error //ping下地址,如果协议一直,则有消息返回

}

func NewPeer(port int, options ...io.OptionServer) (*peer, error) {
	s, err := listen.NewUDPServer(port, func(s *io.Server) {
		s.SetReadWriteWithPkg()
		s.SetDealFunc(func(msg *io.IMessage) {

			m := new(Msg)
			json.Unmarshal(msg.Bytes(), m)

			switch m.Type {
			case TypeRegister:
				registerMsg := new(MsgRegister)
				json.Unmarshal(conv.Bytes(m.Data), registerMsg)

				s.Tag.Set("register", registerMsg)

			}

		})
		s.SetOptions(options...)
	})
	if err != nil {
		return nil, err
	}
	return &peer{
		Server:    s,
		localAddr: &net.UDPAddr{Port: port},
		clients:   maps.NewSafe(),
	}, nil
}

type peer struct {
	*io.Server
	Name      string
	localAddr *net.UDPAddr
	clients   *maps.Safe
	nat       *maps.Safe
}

func (this *peer) WriteTo(addr string, p []byte) (int, error) {
	c, err := this.GetClientOrDial(addr, func() (io.ReadWriteCloser, error) {
		return this.Listener().(*listen.UDPServer).NewUDPClient(addr)
	})
	if err != nil {
		return 0, err
	}
	return c.Write(p)
}

func (this *peer) Ping(addr string) error {
	c, err := this.GetClientOrDial(addr, func() (io.ReadWriteCloser, error) {
		return this.Listener().(*listen.UDPServer).NewUDPClient(addr)
	})
	if err != nil {
		return err
	}
	return c.Ping()
}

// Register 向服务端注册节点信息
func (this *peer) Register(addr string) error {
	c, err := this.GetClientOrDial(addr, func() (io.ReadWriteCloser, error) {
		return this.Listener().(*listen.UDPServer).NewUDPClient(addr)
	})
	if err != nil {
		return err
	}

	_, err = c.WriteAny(Msg{
		Type: TypeRegister,
		Data: MsgRegister{
			Name:       this.Name,
			Version:    Version,
			StartTime:  StartTime.Unix(),
			ConnectKey: "",
			LocalAddr:  this.localAddr.String(),
		},
	})
	return err
}

func (this *peer) Find(addr string) (MsgFind, error) {
	//TODO implement me
	panic("implement me")
}

func (this *peer) Connect(addr string) error {
	//TODO implement me
	panic("implement me")
}

func (this *peer) LocalAddr() net.Addr {
	return this.localAddr
}
