package main

import "github.com/injoyai/io/testdata"

func main() {
	testdata.VPNClient(1082, 1090, ":12000")
	testdata.ProxyTransmit(12000)
	testdata.ProxyClient(":12000")
	select {}
}
