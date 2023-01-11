package io

import (
	"net"
	"testing"
	"time"
)

func TestNewDial(t *testing.T) {
	c := MustDial(func() (ReadWriteCloser, error) {
		return net.Dial("tcp", ":10086")
	})
	//if err != nil {
	//	t.Error(err)
	//	return
	//}

	c.Redial(func(c *Client) {
		t.Log("初始化")
		c.Debug()
		c.SetPrintWithHEX()
		c.SetKey("test")
		go func() {
			for {
				time.Sleep(time.Second)
				_, err := c.WriteString("666")
				if err != nil {
					return
				}
			}
		}()
	})

	go c.Run()
	go c.Run()

	select {}

}
