package dial

import (
	"github.com/goburrow/serial"
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"net"
	"net/http"
	"os"
)

// TCP 连接
func TCP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("tcp", addr)
}

// UDP 连接
func UDP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("udp", addr)
}

// File 打开文件
func File(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}

// Serial 打开串口
func Serial(cfg *serial.Config) (io.ReadWriteCloser, error) {
	return serial.Open(cfg)
}

// MQTT

// HTTP

// MQ

// Websocket 连接
func Websocket(url string, header http.Header) (io.MessageReadWriteCloser, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, header)
	return &_websocket{Conn: c}, err
}

type _websocket struct {
	*websocket.Conn
}

func (this *_websocket) Write(p []byte) (int, error) {
	err := this.Conn.WriteMessage(websocket.TextMessage, p)
	return len(p), err
}

func (this *_websocket) ReadMessage() ([]byte, error) {
	_, bytes, err := this.Conn.ReadMessage()
	return bytes, err
}

func (this *_websocket) Close() error {
	return this.Conn.Close()
}
