package main

import (
	"github.com/injoyai/logs"
	"net"
	"syscall"
)

func main() {

	listen, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 10086,
	})
	logs.PanicErr(err)
	for {
		c, err := listen.AcceptTCP()
		if !logs.PrintErr(err) {
			fd, err := c.File()
			if !logs.PrintErr(err) {
				syscall.Shutdown(syscall.Handle(fd.Fd()), syscall.SHUT_RDWR)
				fd.Close()
			}
		}
	}
}
