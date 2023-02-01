package proxy

import (
	"errors"
)

var (
	ErrPortBind    = errors.New("端口已被绑定")
	ErrNoConnected = errors.New("通道客户端未连接")
)
