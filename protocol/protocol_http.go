package protocol

import (
	"fmt"
	"github.com/injoyai/conv"
	"net/http"
	"strings"
)

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
