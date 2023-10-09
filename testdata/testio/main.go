package main

import (
	"bufio"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"os"
	"time"
)

func main() {

	Test(0)
}

func Test(n int) {
	switch n {
	case 1:
		/*
			局域网测试结果:
			[调试]2023/08/03 15:03:32 main.go:52: [处理]传输耗时: 11.1MB/s
		*/
		logs.SetShowColor(false)
		var start time.Time  //当前时间
		length := 1000 << 20 //传输的数据大小
		totalDeal := 0
		listen.RunTCPServer(10086, func(s *io.Server) {
			s.Logger.SetLevel(io.LevelInfo)
			s.SetDealFunc(func(msg *io.IMessage) {
				if start.IsZero() {
					start = time.Now()
				}
				totalDeal += msg.Len()
				if totalDeal >= length {
					logs.Debugf("[处理]传输耗时: %0.1fMB/s", float64(totalDeal/(1<<20))/time.Now().Sub(start).Seconds())
				}
			})
		})
	case 0:
		/*
			测试结果:
			[调试]2023/08/03 15:03:30 main.go:62: [发送]传输耗时: 4507.1MB/s
			[调试]2023/08/03 15:03:32 main.go:25: [读取]传输耗时: 490.8MB
			[调试]2023/08/03 15:03:32 main.go:52: [处理]传输耗时: 490.7MB/s
		*/
		start := time.Now()  //当前时间
		length := 1000 << 20 //传输的数据大小
		totalRead := 0
		readAll := func(buf *bufio.Reader) (bytes []byte, err error) {
			defer func() {
				totalRead += len(bytes)
				if totalRead >= length {
					logs.Debugf("[读取]传输耗时: %0.1fMB/s", float64(totalRead/(1<<20))/time.Now().Sub(start).Seconds())
				}
			}()

			//read,单次读取大小不影响速度
			num := 4096
			for {
				data := make([]byte, num)
				length, err := buf.Read(data)
				if err != nil {
					return nil, err
				}
				bytes = append(bytes, data[:length]...)
				if length < num || buf.Buffered() == 0 {
					//缓存没有剩余的数据
					return bytes, err
				}
			}
		}

		totalDeal := 0
		go listen.RunTCPServer(10086, func(s *io.Server) {
			s.Logger.SetLevel(io.LevelError)
			s.SetReadFunc(readAll)
			s.SetDealFunc(func(msg *io.IMessage) {
				totalDeal += msg.Len()
				if totalDeal >= length {
					logs.Debugf("[处理]传输耗时: %0.1fMB/s", float64(totalDeal/(1<<20))/time.Now().Sub(start).Seconds())
					os.Exit(1)
				}
			})
		})
		<-time.After(time.Second)
		<-dial.RedialTCP("127.0.0.1:10086", func(c *io.Client) {
			c.SetPrintWithBase()
			data := make([]byte, length)
			start = time.Now()
			c.Write(data)
			logs.Debugf("[发送]传输耗时: %0.1fMB/s", float64(length/(1<<20))/time.Now().Sub(start).Seconds())
			start = time.Now()
			c.SetDealFunc(func(msg *io.IMessage) {
				logs.Debug(msg)
			})
		}).DoneAll()

	}
}
