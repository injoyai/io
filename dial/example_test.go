package dial

import (
	"bufio"
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"os"
)

func ExampleRedialTCP() {
	RedialSSH(&SSHConfig{
		Addr:     "192.168.10.40:22",
		User:     "qinalang",
		Password: "ql1123",
	}, func(c *io.Client) {
		c.Debug()
		go func() {
			for {
				msg := ""
				fmt.Scan(&msg)
				c.WriteString(msg)
			}
		}()
	})
	select {}
}

func ExampleRedialSSH() {
	c, err := NewSSH(&SSHConfig{
		Addr:     os.Args[1],
		User:     os.Args[2],
		Password: os.Args[3],
	})
	if err != nil {
		logs.Err(err)
		return
	}
	c.Debug(false)
	c.SetDealFunc(func(msg *io.IMessage) {
		fmt.Print(msg.String())
	})
	go c.Run()
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-c.DoneAll():
			return
		default:
			bs, _, _ := reader.ReadLine()
			c.Write(bs)
		}
	}
}
