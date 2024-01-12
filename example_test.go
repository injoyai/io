package io

func ExampleRedial() {

	/*
		addr := "127.0.0.1:10086"
		c := io.Redial(dial.WithTCP(addr),
			func(c *io.Client) {
				c.Debug()             //开启打印日志
				c.SetPrintWithUTF8()  //打印日志编码UTF8
				c.SetReadWithAll()    //设置读取全部
				c.SetDealFunc(func(c *Client, msg Message) {
					//业务逻辑,处理读取到的数据
				})
				c.GoTimerWriter(time.Minute, func(c *IWriter) error {
					//定时发送心跳
					_, err := c.WriteString("心跳")
					return err
				})
			})
		<-c.DoneAll()
	*/

}
