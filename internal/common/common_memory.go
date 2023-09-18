package common

import (
	"bytes"
	"fmt"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"time"
)

var MemoryServerManage = maps.NewSafe()

func NewMemoryServer(key string) *MemoryServer {
	s, _ := MemoryServerManage.GetOrSetByHandler(key, func() (interface{}, error) {
		return &MemoryServer{
			Key: key,
			Ch:  make(chan io.ReadWriteCloser, 1000),
		}, nil
	})
	return s.(*MemoryServer)
}

// MemoryServer 虚拟服务,为了实现接口
type MemoryServer struct {
	Key string
	Ch  chan io.ReadWriteCloser
}

func (this *MemoryServer) Connect() (io.ReadWriteCloser, error) {
	return this.ConnectWithTimeout(io.DefaultConnectTimeout)
}

func (this *MemoryServer) ConnectWithTimeout(timeout time.Duration) (io.ReadWriteCloser, error) {
	c := &MemoryClient{Buffer: bytes.NewBuffer(nil)}
	select {
	case this.Ch <- c:
	case <-time.After(timeout):
		return nil, io.ErrWithTimeout
	}
	return c, nil
}

func (this *MemoryServer) Accept() (io.ReadWriteCloser, string, error) {
	c := <-this.Ch
	return c, fmt.Sprintf("%p", c), nil
}

func (this *MemoryServer) Close() error {
	MemoryServerManage.Del(this.Key)
	return nil
}

func (this *MemoryServer) Addr() string {
	return fmt.Sprintf("%p", this)
}

type MemoryClient struct {
	*bytes.Buffer
}

func (this *MemoryClient) Close() error {
	this.Reset()
	return nil
}
