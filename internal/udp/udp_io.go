package udp

import (
	"io"
	"net"
)

var _ io.ReadWriteCloser = (*UDP)(nil)

type UDP struct {
	*net.UDPConn
	to *net.UDPAddr
}

func (this *UDP) Read(b []byte) (int, error) {
	//读取设置地址的数据,其他地址过来的数据过滤掉
	for {
		n, addr, err := this.UDPConn.ReadFromUDP(b)
		if err != nil {
			return 0, err
		}
		if addr.String() == this.to.String() {
			return n, nil
		}
	}
}

func (this *UDP) Write(b []byte) (int, error) {
	return this.UDPConn.WriteToUDP(b, this.to)
}

func New(address string, to *net.UDPAddr) (*UDP, error) {
	c, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}
	return &UDP{
		UDPConn: c.(*net.UDPConn),
		to:      to,
	}, nil
}
