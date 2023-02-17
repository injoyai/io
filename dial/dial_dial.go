package dial

import (
	"bytes"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/serial"
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"net"
	"net/http"
	"os"
)

//================================TCPDial================================

// TCP 连接
func TCP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("tcp", addr)
}

// TCPFunc 连接函数
func TCPFunc(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return TCP(addr) }
}

// NewTCP 新建TCP连接
func NewTCP(addr string) (*io.Client, error) {
	return io.NewDial(TCPFunc(addr))
}

// RedialTCPWithContext 一直连接TCP服务端,并重连
func RedialTCPWithContext(ctx context.Context, addr string, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.RedialWithContext(ctx, TCPFunc(addr), fn...)
}

// RedialTCP 一直连接TCP服务端,并重连
func RedialTCP(addr string, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return RedialTCPWithContext(context.Background(), addr, fn...)
}

//================================UDPDial================================

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

func NewUDP(addr string) (*io.Client, error) {
	return io.NewDial(UDPFunc(addr))
}

// RedialUDPWithContext 一直连接UDP服务端,并重连
func RedialUDPWithContext(ctx context.Context, addr string, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.RedialWithContext(ctx, TCPFunc(addr), fn...)
}

// RedialUDP 一直连接UDP服务端,并重连
func RedialUDP(addr string, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return RedialTCPWithContext(context.Background(), addr, fn...)
}

//================================FileDial================================

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

func NewFile(path string) (*io.Client, error) {
	return io.NewDial(FileFunc(path))
}

//================================WebsocketDial================================

// Memory 内存
func Memory() (io.ReadWriteCloser, error) {
	return &_memory{Buffer: bytes.NewBuffer(nil)}, nil
}

func NewMemory() (*io.Client, error) {
	return io.NewDial(Memory)
}

type _memory struct {
	*bytes.Buffer
}

func (this *_memory) Close() error {
	this.Reset()
	return nil
}

//================================SerialDial================================

type SerialConfig = serial.Config

// Serial 打开串口
func Serial(cfg *SerialConfig) (io.ReadWriteCloser, error) {
	return serial.Open(cfg)
}

// SerialFunc 打开串口函数
func SerialFunc(cfg *SerialConfig) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return serial.Open(cfg)
	}
}

func NewSerial(cfg *SerialConfig) (*io.Client, error) {
	return io.NewDial(SerialFunc(cfg))
}

func RedialSerial(cfg *SerialConfig, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(SerialFunc(cfg), fn...)
}

//================================HTTPDial================================

//================================MQTTDial================================

type MQTTConfig = mqtt.ClientOptions

func MQTT(clientID, topic string, qos byte, cfg *MQTTConfig) (io.ReadWriteCloser, error) {
	c := mqtt.NewClient(cfg)
	token := c.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}
	r := &_mqtt{
		Client:   c,
		clientID: clientID,
		topic:    topic,
		qos:      qos,
		ch:       make(chan mqtt.Message, 1000),
	}
	c.Subscribe(clientID, qos, func(client mqtt.Client, message mqtt.Message) {
		r.ch <- message
	})
	return r, token.Error()
}

type _mqtt struct {
	mqtt.Client
	clientID string
	topic    string
	qos      byte
	ch       chan mqtt.Message
}

func (this *_mqtt) Read(p []byte) (int, error) {
	return 0, nil
}

func (this *_mqtt) ReadMessage() ([]byte, error) {
	msg := <-this.ch
	defer msg.Ack()
	return msg.Payload(), nil
}

func (this *_mqtt) Write(p []byte) (int, error) {
	token := this.Client.Publish(this.topic, this.qos, false, p)
	token.Wait()
	return len(p), token.Error()
}

func (this *_mqtt) Close() error {
	token := this.Client.Unsubscribe(this.clientID)
	token.Wait()
	return token.Error()
}

//================================RabbitmqDial================================

//================================WebsocketDial================================

// Websocket 连接
func Websocket(url string, header http.Header) (io.MessageReadWriteCloser, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, header)
	return &_websocket{Conn: c}, err
}

func WebsocketFunc(url string, header http.Header) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		c, _, err := websocket.DefaultDialer.Dial(url, header)
		return &_websocket{Conn: c}, err
	}
}

func NewWebsocket(url string, header http.Header) (*io.Client, error) {
	return io.NewDial(WebsocketFunc(url, header))
}

func RedialWebsocket(url string, header http.Header, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(WebsocketFunc(url, header), fn...)
}

type _websocket struct {
	*websocket.Conn
}

func (this *_websocket) Read(p []byte) (int, error) {
	return 0, nil
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
