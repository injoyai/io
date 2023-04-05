package port_forwarding

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
)

func NewServer() {
	ser, err := dial.NewTCPServer(cfg.GetInt("port", 9000))
	if err != nil {
		logs.Err(err)
		return
	}
	for _, v := range cfg.GetStrings("listen") {
		m := conv.NewMap(v)
		port := m.GetInt("port")
		sn := m.GetString("sn")
		addr := m.GetString("addr")
		_, err = dial.NewTCPServer(port, func(s *io.Server) {
			s.Debug()
			s.SetDealFunc(func(msg *io.IMessage) {
				c := ser.GetClient(sn)
				if c != nil {
					c.WriteAny(proxy.NewWriteMessage(msg.GetKey(), addr, msg.Bytes()))
				}
			})
		})
		logs.PrintErr(err)
	}
	ser.Debug()
	ser.Run()
}

type Server struct {
}
