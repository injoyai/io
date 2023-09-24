package p2p

import (
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"net"
)

const (
	UDP = "udp"
)

type Peer interface {
	LocalAddr() net.Addr               //本地地址
	Ping(addr string) error            //ping下地址,如果协议一直,则有消息返回
	Base(addr string) (MsgBase, error) //获取基本信息
	Find(addr string) (MsgFind, error) //
	Connect(addr string) error         //建立连接
}

func NewPeer(port int) (*peer, error) {
	//localAddr := &net.UDPAddr{Port: port}
	//c, err := net.ListenUDP("udp", localAddr)
	//if err != nil {
	//	return nil, err
	//}
	s, err := listen.NewUDPServer(port)
	if err != nil {
		return nil, err
	}
	return &peer{
		port: port,
		s:    s,
		//peer:    c,
		clients: maps.NewSafe(),
	}, nil
}

type peer struct {
	port      int //占用的端口
	localAddr *net.UDPAddr
	s         *io.Server
	peer      *net.UDPConn //监听的服务
	clients   *maps.Safe
}

func (this *peer) WriteTo(addr string, p []byte) (int, error) {
	if c := this.s.GetClient(addr); c != nil {
		return c.Write(p)
	}
	c, err := this.s.DialClient(func() (io.ReadWriteCloser, error) {
		return this.s.Listener().(*listen.UDPServer).NewUDPClient(addr)
	})
	if err != nil {
		return 0, err
	}
	return c.Write(p)
}

func (this *peer) Ping(addr string) error {
	if _, err := this.WriteTo(addr, io.NewPkgPing()); err != nil {
		return err
	}
	//if err := c.SetDeadline(time.Now().Add(time.Second)); err != nil {
	//	return err
	//}
	//p, err := io.ReadPkg(bufio.NewReader(c))
	//if err != nil {
	//	return err
	//}
	//if !p.IsPong() {
	//	return errors.New("响应失败")
	//}
	return nil
}

func (this *peer) Base(addr string) (MsgBase, error) {
	return MsgBase{}, nil
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
