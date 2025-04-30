package rpc

import (
	"context"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	uuid "github.com/satori/go.uuid"
	"time"
)

type Handler func(ctx context.Context, c *io.Client, msg *io.Model) (interface{}, error)

type Server struct {
	*io.Server
	bind *maps.Safe[string, Handler]
	wait *wait.Entity[any, any]
}

func (this *Server) Bind(Type string, handler Handler) {
	this.bind.Set(Type, handler)
}

func (this *Server) Do(key, Type string, data interface{}) (interface{}, error) {
	c := this.GetClient(key)
	return do(c, this.wait, Type, data)
}

func (this *Server) Run() error {
	return this.Server.Run()
}

func (this *Server) dealFunc(c *io.Client, msg io.Message) {
	dealFunc(this.bind, this.wait, c, msg)
}

func NewServer(port int, waitTimeout time.Duration, option ...io.OptionServer) (*Server, error) {
	s, err := listen.NewTCPServer(port, option...)
	if err != nil {
		return nil, err
	}
	ser := &Server{
		Server: s,
		bind:   maps.NewSafe[string, Handler](),
		wait:   wait.NewDefault(waitTimeout),
	}
	s.SetReadFunc(io.ReadWithPkg)
	s.SetWriteFunc(io.WriteWithPkg)
	s.SetDealFunc(ser.dealFunc)
	s.SetTimeout(io.DefaultTimeout)
	s.SetTimeoutInterval(io.DefaultKeepAlive)
	ser.Bind(io.Register, func(ctx context.Context, c *io.Client, msg *io.Model) (interface{}, error) {
		key := uuid.NewV4().String()
		m := conv.NewMap(msg.Data)
		c.Tag().Set(io.Register, true)
		c.Tag().Set(io.Register+".key", key)
		c.Tag().Set(io.Register+".name", m.GetString("name"))
		c.Tag().Set(io.Register+".memo", m.GetString("memo"))
		c.Tag().Set(io.Register+".version", m.GetString("version"))
		return map[string]string{"key": key}, nil
	})
	return ser, nil
}
