package dial

import (
	"bytes"
	"context"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/goburrow/serial"
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"golang.org/x/crypto/ssh"
	"net"
	"net/http"
	gourl "net/url"
	"os"
	"time"
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
	c, err := io.NewDial(TCPFunc(addr))
	if err == nil {
		c.SetKey(addr)
	}
	return c, err
}

// RedialTCP 一直连接TCP服务端,并重连
func RedialTCP(addr string, options ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(TCPFunc(addr), func(ctx context.Context, c *io.Client) {
		c.SetKey(addr).SetOptions(options...)
	})
}

//================================UDPDial================================

// UDP 连接
func UDP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("udp", addr)
}

// UDPFunc 连接函数
func UDPFunc(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return UDP(addr) }
}

func NewUDP(addr string) (*io.Client, error) {
	c, err := io.NewDial(UDPFunc(addr))
	if err == nil {
		c.SetKey(addr)
	}
	return c, err
}

// RedialUDP 一直连接UDP服务端,并重连
func RedialUDP(addr string, options ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(UDPFunc(addr), func(ctx context.Context, c *io.Client) {
		c.SetKey(addr).SetOptions(options...)
	})
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
	c, err := io.NewDial(SerialFunc(cfg))
	if err == nil {
		c.SetKey(cfg.Address)
	}
	return c, err
}

func RedialSerial(cfg *SerialConfig, options ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(SerialFunc(cfg), func(ctx context.Context, c *io.Client) {
		c.SetKey(cfg.Address).SetOptions(options...)
	})
}

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

func MQTTFunc(clientID, topic string, qos byte, cfg *MQTTConfig) func() (io.ReadWriteCloser, error) {
	return func() (closer io.ReadWriteCloser, err error) {
		return MQTT(clientID, topic, qos, cfg)
	}
}

func NewMQTT(clientID, topic string, qos byte, cfg *MQTTConfig) (*io.Client, error) {
	c, err := io.NewDial(MQTTFunc(clientID, topic, qos, cfg))
	if err == nil {
		c.SetKey(topic)
	}
	return c, err
}

func RedialMQTT(clientID, topic string, qos byte, cfg *MQTTConfig, fn ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(MQTTFunc(clientID, topic, qos, cfg.SetAutoReconnect(true)), func(ctx context.Context, c *io.Client) {
		c.SetKey(topic)
		for _, v := range fn {
			v(ctx, c)
		}
	})
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
	c, err := io.NewDial(WebsocketFunc(url, header))
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

func RedialWebsocket(url string, header http.Header, options ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(WebsocketFunc(url, header), func(ctx context.Context, c *io.Client) {
		c.SetKey(func() string {
			if u, err := gourl.Parse(url); err == nil {
				return u.Path
			}
			return url
		}())
		for _, v := range options {
			v(ctx, c)
		}
	})
}

type _websocket struct {
	*websocket.Conn
}

// Read 无效,请使用ReadMessage
func (this *_websocket) Read(p []byte) (int, error) {
	return 0, nil
}

func (this *_websocket) Write(p []byte) (int, error) {
	err := this.Conn.WriteMessage(websocket.TextMessage, p)
	return len(p), err
}

func (this *_websocket) ReadMessage() ([]byte, error) {
	_, bs, err := this.Conn.ReadMessage()
	return bs, err
}

func (this *_websocket) Close() error {
	return this.Conn.Close()
}

//================================SSH================================

type SSHConfig struct {
	Addr        string
	User        string
	Password    string //类型为password
	Type        string //password 或者 key
	Timeout     time.Duration
	key         string //类型为key
	keyPassword string //类型为key
	High        int    //高
	Wide        int    //宽
}

func (this *SSHConfig) new() *SSHConfig {
	if len(this.User) == 0 {
		this.User = "root"
	}
	if this.Timeout == 0 {
		this.Timeout = time.Second
	}
	if this.High == 0 {
		this.High = 32
	}
	if this.Wide == 0 {
		this.Wide = 300
	}
	return this
}

type Client struct {
	io.Writer
	io.Reader
	*ssh.Session
	err io.Reader
}

func SSH(cfg *SSHConfig) (io.ReadWriteCloser, error) {
	cfg.new()
	config := &ssh.ClientConfig{
		Timeout:         cfg.Timeout,
		User:            cfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(cfg.Password)},
	}
	switch cfg.Type {
	case "key":
		signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.key), []byte(cfg.keyPassword))
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	sshClient, err := ssh.Dial("tcp", cfg.Addr, config)
	if err != nil {
		return nil, err
	}
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}
	reader, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	outputErr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}
	writer, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := session.RequestPty("xterm", cfg.High, cfg.Wide, modes); err != nil {
		return nil, err
	}
	if err := session.Shell(); err != nil {
		return nil, err
	}
	return &Client{
		Writer:  writer,
		Reader:  reader,
		Session: session,
		err:     outputErr,
	}, nil
}

func SSHFunc(cfg *SSHConfig) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return SSH(cfg)
	}
}

func NewSSH(cfg *SSHConfig) (*io.Client, error) {
	c, err := io.NewDial(SSHFunc(cfg))
	if err == nil {
		c.SetKey(cfg.Addr)
	}
	return c, err
}

func RedialSSH(cfg *SSHConfig, options ...func(ctx context.Context, c *io.Client)) *io.Client {
	return io.Redial(SSHFunc(cfg), func(ctx context.Context, c *io.Client) {
		c.SetKey(cfg.Addr).SetOptions(options...)
	})
}

//================================OtherDial================================
