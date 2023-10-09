package io

func ExampleRedial() {

	/*
		addr := "127.0.0.1:10086"
		c := Redial(dial.TCPFunc(addr),
			func(c *Client) {
				c.Debug()             //开启打印日志
				c.SetPrintWithASCII() //打印日志编码ASCII
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
