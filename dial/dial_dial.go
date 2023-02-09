package dial

import (
	"bytes"
	"github.com/goburrow/serial"
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"net"
	"net/http"
	"os"
)

type SerialConfig = serial.Config

// Memory 内存
func Memory() (io.ReadWriteCloser, error) {
	return &_memory{Buffer: bytes.NewBuffer(nil)}, nil
}

// TCP 连接
func TCP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("tcp", addr)
}

// TCPFunc 连接函数
func TCPFunc(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return TCP(addr) }
}

// UDP 连接
func UDP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("udp", addr)
}

// UDPFunc 连接函数
func UDPFunc(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return net.Dial("udp", addr)
	}
}

// File 打开文件
func File(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}

// FileFunc 打开文件函数
func FileFunc(path string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return os.Open(path)
	}
}

// Serial 打开串口
func Serial(cfg *serial.Config) (io.ReadWriteCloser, error) {
	return serial.Open(cfg)
}

// SerialFunc 打开串口函数
func SerialFunc(cfg *serial.Config) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return serial.Open(cfg)
	}
}

// MQTT 连接

// HTTP 连接

// MQ 连接

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

type _memory struct {
	*bytes.Buffer
}

func (this *_memory) Close() error {
	this.Reset()
	return nil
}
