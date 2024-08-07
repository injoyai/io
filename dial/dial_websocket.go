package dial

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"net/http"
	gourl "net/url"
)

//================================Websocket================================

type WebsocketConfig struct {
	Dialer *websocket.Dialer
	Url    string
	Header http.Header
}

func (this *WebsocketConfig) Dial() (io.ReadWriteCloser, string, error) {
	if this.Dialer == nil {
		this.Dialer = websocket.DefaultDialer
	}
	c, _, err := this.Dialer.Dial(this.Url, this.Header)
	return &WebsocketClient{Conn: c}, func() string {
		if u, err := gourl.Parse(this.Url); err == nil {
			return u.Path
		}
		return this.Url
	}(), err
}

// Websocket 连接
func Websocket(url string, header http.Header) (io.ReadWriteCloser, string, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, header)
	return &WebsocketClient{Conn: c}, url, err
}

func WithWebsocket(url string, header http.Header) io.DialFunc {
	return func(ctx context.Context) (io.ReadWriteCloser, string, error) { return Websocket(url, header) }
}

func NewWebsocket(url string, header http.Header, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithWebsocket(url, header), options...)
}

func RedialWebsocket(url string, header http.Header, options ...io.OptionClient) *io.Client {
	return io.Redial(WithWebsocket(url, header), options...)
}

type WebsocketClient struct {
	*websocket.Conn
}

// Read 无效,请使用ReadMessage
func (this *WebsocketClient) Read(p []byte) (int, error) {
	return 0, io.ErrUseReadMessage
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
