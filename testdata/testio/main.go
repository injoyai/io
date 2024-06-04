package main

import (
	"bufio"
	"github.com/injoyai/io"
	"github.com/injoyai/io/buf"
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
			s.SetLevel(io.LevelInfo)
			s.SetDealFunc(func(c *io.Client, msg io.Message) {
				if start.IsZero() {
					start = time.Now()
				}
				totalDeal += msg.Len()
				if totalDeal >= length {
					logs.Debugf("[处理]传输耗时: %0.1fMB/s\n", float64(totalDeal/(1<<20))/time.Now().Sub(start).Seconds())
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
		readAll := func(r *bufio.Reader) (bytes []byte, err error) {
			defer func() {
				totalRead += len(bytes)
				if totalRead >= length {
					logs.Debugf("[读取]传输耗时: %0.1fMB/s\n", float64(totalRead/(1<<20))/time.Now().Sub(start).Seconds())
				}
			}()

			return buf.Read1KB(r)
		}

		totalDeal := 0
		go listen.RunTCPServer(20145, func(s *io.Server) {
			s.SetLevel(io.LevelError)
			s.Debug(false)
			s.SetReadFunc(readAll)
			s.SetDealFunc(func(c *io.Client, msg io.Message) {
				totalDeal += msg.Len()
				if totalDeal >= length {
					logs.Debugf("[处理]传输耗时: %0.1fMB/s\n", float64(totalDeal/(1<<20))/time.Now().Sub(start).Seconds())
					os.Exit(1)
				}
			})
		})
		<-time.After(time.Second)
		<-dial.RedialTCP("127.0.0.1:20145", func(c *io.Client) {
			c.Debug(false)
			c.SetLevelInfo()
			data := make([]byte, length)
			start = time.Now()
			c.Write(data)
			logs.Debugf("[发送]传输耗时: %0.1fMB/s\n", float64(length/(1<<20))/time.Now().Sub(start).Seconds())
			start = time.Now()
			c.SetDealFunc(func(c *io.Client, msg io.Message) {
				logs.Debug(msg)
			})
		}).DoneAll()

	}
}
