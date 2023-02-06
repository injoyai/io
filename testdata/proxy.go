package testdata

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"net"
)

func Client(udpPort, tcpPort int) {

	io.NewServer(dial.TCPListenFunc(tcpPort))
	io.NewServer(dial.TCPListenFunc(tcpPort))

}

func ListenProxy(udpPort, tcpPort int) {
	c := io.NewClient()
	serUDP := io.NewServer(func() (io.Listener, error) {
		return net.ListenUDP("udp", &net.UDPAddr{Port: udpPort})
	})
}
