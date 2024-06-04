package dial

import (
	"bytes"
	"context"
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"net/http"
	gourl "net/url"
)

//================================HTTP================================

func HTTP(c *http.Client, method, url string, fn ...func(r *http.Request)) (io.ReadWriteCloser, string, error) {
	u, err := gourl.Parse(url)
	if err != nil {
		return nil, "", err
	}
	return &HTTPClient{
		Client: c,
		ch:     make(chan io.Message, io.DefaultChannelSize),
		Method: method,
		Url:    url,
		fn:     fn,
	}, u.Path, nil
}

func WithHTTP(c *http.Client, method, url string) io.DialFunc {
	return func(ctx context.Context) (io.ReadWriteCloser, string, error) { return HTTP(c, method, url) }
}

func NewHTTP(c *http.Client, method, url string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithHTTP(c, method, url), options...)
}

func RedialHTTP(c *http.Client, method, url string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithHTTP(c, method, url), options...)
}

type HTTPClient struct {
	*http.Client
	Method string
	Url    string
	ch     chan io.Message
	fn     []func(r *http.Request)
}

func (this *HTTPClient) Read(p []byte) (int, error) {
	return 0, io.ErrUseReadMessage
}

func (this *HTTPClient) Write(p []byte) (int, error) {
	req, err := http.NewRequest(this.Method, this.Url, bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}
	for _, v := range this.fn {
		v(req)
	}
	resp, err := this.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return 0, errors.New(resp.Status)
	}
	select {
	case this.ch <- conv.Bytes(resp.Body).Bytes():
	default:
	}
	return len(p), nil
}

func (this *HTTPClient) ReadMessage() ([]byte, error) {
	data, ok := <-this.ch
	if !ok {
		return nil, errors.New("已关闭")
	}
	return data, nil
}

func (this *HTTPClient) Close() error {
	this.Client.CloseIdleConnections()
	close(this.ch)
	return nil
}
