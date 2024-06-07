package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"time"
)

/*
2024-06-07 08:58:33 [信息] [:20123] 开启服务成功...
2024-06-07 08:58:33 [错误] [127.0.0.1:20123] 断开连接: 主动关闭
2024-06-07 08:58:33 [信息] [127.0.0.1:53763] 新的客户端连接...
2024-06-07 08:58:33 [错误] [127.0.0.1:53763] 断开连接: EOF
2024-06-07 08:58:33 [信息] [127.0.0.1:20123] 连接服务端成功...
2024-06-07 08:58:33 [信息] [127.0.0.1:53766] 新的客户端连接...
*/
func main() {
	go listen.RunTCPServer(20123)
	dial.NewTCP("127.0.0.1:20123", func(c *io.Client) {
		c.Close()
	})
	dial.NewTCP("127.0.0.1:20123")
	<-time.After(time.Second * 3)
}
