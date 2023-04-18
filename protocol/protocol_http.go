package protocol

import (
	"bytes"
	"fmt"
	"github.com/injoyai/conv"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

type HTTP11Request http.Request

func (this *HTTP11Request) AddHeader(key string, val ...string) *HTTP11Request {
	for _, v := range val {
		this.Header.Add(key, v)
	}
	return this
}

func (this *HTTP11Request) AddHeaders(m map[string][]string) *HTTP11Request {
	for k, v := range m {
		this.AddHeader(k, v...)
	}
	return this
}

func (this *HTTP11Request) SetHeader(key string, val ...string) *HTTP11Request {
	this.Header.Del(key)
	for _, v := range val {
		this.Header.Add(key, v)
	}
	return this
}

func (this *HTTP11Request) SetHeaders(m map[string][]string) *HTTP11Request {
	for k, v := range m {
		this.SetHeader(k, v...)
	}
	return this
}

func (this *HTTP11Request) AddCookie(cookies ...*http.Cookie) *HTTP11Request {
	for _, v := range cookies {
		(*http.Request)(this).AddCookie(v)
	}
	return this
}

func (this *HTTP11Request) GetCookies() []*http.Cookie {
	return (*http.Request)(this).Cookies()
}

func (this *HTTP11Request) GetCookie(key string) (*http.Cookie, error) {
	return (*http.Request)(this).Cookie(key)
}

func (this *HTTP11Request) SetUserAgent(s string) *HTTP11Request {
	return this.SetHeader("User-Agent", s)
}

// SetReferer 设置Referer
func (this *HTTP11Request) SetReferer(s string) *HTTP11Request {
	return this.SetHeader("Referer", s)
}

func (this *HTTP11Request) SetAuthorization(s string) *HTTP11Request {
	return this.SetHeader("Authorization", s)
}

func (this *HTTP11Request) SetToken(s string) *HTTP11Request {
	return this.SetHeader("Authorization", s)
}

func (this *HTTP11Request) SetContentType(s string) *HTTP11Request {
	return this.SetHeader("Content-Type", s)
}

func (this *HTTP11Request) SetBody(body interface{}) *HTTP11Request {
	switch val := body.(type) {
	case io.ReadCloser:
		this.Body = val
	case io.Reader:
		this.Body = io.NopCloser(val)
	default:
		this.Body = io.NopCloser(bytes.NewReader(conv.Bytes(body)))
	}
	return this
}

func (this *HTTP11Request) Bytes() []byte {
	bs, _ := httputil.DumpRequest((*http.Request)(this), true)
	return bs
}

func NewHTTP11Request(method, url string, body io.Reader) *HTTP11Request {
	r, _ := http.NewRequest(method, url, body)
	return (*HTTP11Request)(r)
}

type HTTP11Response struct {
	StatusCode     int
	StatusCodeText string
	Header         http.Header
	Body           []byte
}

func (this *HTTP11Response) AddHeader(key, val string) *HTTP11Response {
	this.Header[key] = append(this.Header[key], val)
	return this
}

func (this *HTTP11Response) SetBody(body []byte) *HTTP11Response {
	this.Body = body
	return this
}

func (this *HTTP11Response) Bytes() []byte {
	data := []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", this.StatusCode, this.StatusCodeText))
	for i, v := range this.Header {
		data = append(data, []byte(fmt.Sprintf("%s: %s\r\n", i, strings.Join(v, "; ")))...)
	}
	data = append(data, []byte("\r\n")...)
	return append(data, this.Body...)
}

func NewHTTPResponse(statusCode int, body []byte) *HTTP11Response {
	return &HTTP11Response{
		StatusCode:     statusCode,
		StatusCodeText: "",
		Header: http.Header{
			"Content-Length": conv.Strings(len(body)),
			"Content-Type":   []string{"text/html", "charset=utf-8"},
		},
		Body: body,
	}
}

func NewHTTPResponseBytes(statusCode int, body []byte) []byte {
	return NewHTTPResponse(statusCode, body).Bytes()
}

func NewHTTPResponseBytes200(body []byte) []byte {
	return NewHTTPResponse(200, body).Bytes()
}

func NewHTTPResponseBytes204() []byte {
	return NewHTTPResponse(204, nil).Bytes()
}

func NewHTTPResponseBytes301(addr string) []byte {
	return NewHTTPResponse(301, nil).AddHeader("Location", addr).Bytes()
}

func NewHTTPResponseBytes302(addr string) []byte {
	return NewHTTPResponse(302, nil).AddHeader("Location", addr).Bytes()
}

func NewHTTPResponse400(body []byte) []byte {
	return NewHTTPResponseBytes(400, body)
}

func NewHTTPResponse500(body []byte) []byte {
	return NewHTTPResponseBytes(500, body)
}
