package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/injoyai/base/g"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/maps/wait"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"net/http"
)

// dealFunc 处理远程消息
func dealFunc(bind *maps.Safe, wait *wait.Entity, c *io.Client, msg io.Message) {
	req := new(io.Model)
	if err := json.Unmarshal(msg.Bytes(), req); err != nil {
		c.CloseWithErr(err)
		return
	}
	cm := conv.NewMap(req.Data)
	switch req.Type {
	case io.Register:
		if req.IsRequest() {
			c.Tag().Set(io.Register, true)
			c.WriteAny(req.Resp(
				http.StatusOK,
				g.Map{"key": conv.String(g.UUID())},
				"注册成功",
			))
		} else {
			c.Tag().Set(io.Register, true)
			c.Tag().Set(io.Register+".key", cm.GetString("key"))
		}
		return
	}
	if req.IsRequest() {
		h, ok := bind.Get(req.Type)
		if !ok {
			req.Resp(http.StatusNotFound, nil, "not find")
			return
		}
		//协程处理,防止阻塞
		go func(c *io.Client, h Handler, m *io.Model, cm *conv.Map) {
			data, err := h(context.Background(), cm)
			c.WriteAny(m.Resp(
				conv.SelectInt(err == nil, http.StatusOK, http.StatusInternalServerError),
				data,
				conv.New(err).String("成功"),
			))
		}(c, h.(Handler), req, cm)
		return
	}
	//响应数据
	wait.Done(
		req.UID,
		conv.Select(req.IsSucc(), req.Data, errors.New(req.Msg)),
	)
}

// do 执行远程调用
func do(c io.AnyWriterClosed, wait *wait.Entity, Type string, data interface{}) (interface{}, error) {
	if c == nil || c.Closed() {
		return nil, errors.New("rpc未连接")
	}
	uid := g.UUID()
	m := &io.Model{
		Type: Type,
		UID:  uid,
		Data: data,
	}
	if _, err := c.WriteAny(m); err != nil {
		return nil, err
	}
	return wait.Wait(uid)
}
