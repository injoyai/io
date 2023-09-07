package p2p

import "net"

/*
Node
一个端口占用为一个节点
一个程序可以有多个节点

*/
type Node interface {
	// ID 节点标识
	ID() string

	// LocalAddr 地址
	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	SetRemoteAddr(net.Addr)
}
