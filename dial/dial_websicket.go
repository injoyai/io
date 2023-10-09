package dial

import (
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"net/http"
	gourl "net/url"
)

//================================Websocket================================

// Websocket 连接
func Websocket(url string, header http.Header) (io.ReadWriteCloser, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, header)
	return &WebsocketClient{Conn: c}, err
}

func WithWebsocket(url string, header http.Header) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return Websocket(url, header)
	}
}

func NewWebsocket(url string, header http.Header) (*io.Client, error) {
	c, err := io.NewDial(WithWebsocket(url, header))
	if err == nil {
		c.SetKey(func() string {
			if u, err := gourl.Parse(url); err == nil {
				return u.Path
			}
			return url
		}())
	}
	return c, err
}

func RedialWebsocket(url string, header http.Header, options ...io.OptionClient) *io.Client {
	return io.Redial(WithWebsocket(url, header), func(c *io.Client) {
		c.SetKey(func() string {
			if u, err := gourl.Parse(url); err == nil {
				return u.Path
			}
			return url
		}())
		c.SetOptions(options...)
	})
}

type WebsocketClient struct {
	*websocket.Conn
}

// Read 无效,请使用ReadMessage
func (this *WebsocketClient) Read(p []byte) (int, error) {
	return 0, nil
}

func (this *WebsocketClient) Write(p []byte) (int, error) {
	err := this.Conn.WriteMessage(websocket.TextMessage, p)
	return len(p), err
}

func (this *WebsocketClient) ReadMessage() ([]byte, error) {
	_, bs, err := this.Conn.ReadMessage()
	return bs, err
}

func (this *WebsocketClient) Close() error {
	return this.Conn.Close()
}
