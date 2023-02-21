package proxy

import (
	"fmt"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"net/http"
	"net/url"
	"strings"
)

// NewClient 代理请求
func NewClient() {

}

// NewServer 被请求
func NewServer(port int, c *io.Client) (*io.Server, error) {
	s, err := io.NewServer(dial.TCPListenFunc(port))
	if err != nil {
		return nil, err
	}
	s.SetKey(fmt.Sprintf(":%d", port))
	s.SetPrintFunc(func(msg io.Message, tag ...string) {
		logs.Debug(io.PrintfWithASCII(msg, append([]string{"P|S"}, tag...)...))
	})
	s.SetDealFunc(func(msg *io.IMessage) {
		if c == nil {
			return
		}
		// HTTP 请求
		if list := strings.Split(msg.String(), " "); len(list) > 2 && strings.Contains(list[2], "HTTP") {
			if list[0] == http.MethodConnect {
				msg.Client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
				return
			} else {
				// HTTP 普通请求
				u, err := url.Parse(list[1])
				if err == nil {
					port := u.Port()
					if len(port) == 0 {
						switch strings.ToLower(u.Scheme) {
						case "https":
							port = "443"
						default:
							port = "80"
						}
					}
					addr := fmt.Sprintf("%s:%s", u.Hostname(), port)
					c.WriteAny(NewWriteMessage("test", addr, msg.Bytes()))
				}
			}
		}
	})
	return s, nil
}
