package port_forwarding

import (
	"github.com/injoyai/io/dial/proxy"
)

var NewClient = proxy.SwapTCPClient
