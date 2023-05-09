package dial

import (
	"github.com/injoyai/io"
)

func ExampleRedialTCP() {
	addr := "127.0.0.1:10086"
	RedialTCP(addr, func(c *io.Client) {
		c.Debug()
	})
}
