# io


### 如何使用

#### 如何连接TCP

```go

package main

import (
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"time"
)

func main() {
	addr := "127.0.0.1:10086"
	c := io.Redial(dial.WithTCP(addr),
		func(c *io.Client) {
			c.Debug()             //开启打印日志
			c.SetPrintWithASCII() //打印日志编码ASCII
			c.SetReadWithAll()    //设置读取方式,一次读取全部
			c.SetDealFunc(func(c *io.Client, msg io.Message) {
				// todo 业务逻辑,处理读取到的数据
			})
			c.GoTimerWriter(time.Minute, func(w *io.IWriter) error {
				_,err:= w.WriteString("心跳") //定时发送心跳
				return err
			})
		})
	<-c.DoneAll()
}

```

#### 如何连接SSH

```go

package main

import (
	"bufio"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"os"
)

func main() {
	c := dial.RedialSSH(&dial.SSHConfig{
		Addr:     os.Args[1],
		User:     os.Args[2],
		Password: os.Args[3],
	})
	if err != nil {
		logs.Err(err)
		return
	}
	c.Logger.Debug(false)
	c.SetDealFunc(func(c *io.Client, msg io.Message) {
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


```