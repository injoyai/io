# io


### 如何使用

```

    import (
        "github.com/injoyai/io"
        "github.com/injoyai/io/dial"
    )

    func main() {
        addr := "127.0.0.1:10086"
        c:=io.Redial(dial.TCPFunc(addr),
            func(c *io.Client) {
                c.Debug()             //开启打印日志
                c.SetPrintWithASCII() //打印日志编码ASCII
                c.SetReadWithAll()    //设置读取方式,一次读取全部
                c.SetDealFunc(func(msg *io.IMessage) {
                    // todo 业务逻辑,处理读取到的数据
                })
                c.GoTimerWriter(time.Minute, func(c *io.IWriter) (int, error) {
                    return c.WriteString("心跳") //定时发送心跳
                })
            })
        <-c.DoneAll()
    }
		
```