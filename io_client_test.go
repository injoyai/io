package io

import (
	"net"
	"testing"
	"time"
)

func TestNewDial(t *testing.T) {
	c, err := NewDial(func() (ReadWriteCloser, error) {
		return net.Dial("tcp", ":10086")
	})
	if err != nil {
		t.Error(err)
	}
	c.Debug()

	c.Redial(func(c *Client) {
		t.Log("重连")
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

	select {}

}
