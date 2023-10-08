package dial

import (
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/io"
	"github.com/injoyai/io/internal/common"
	"golang.org/x/crypto/ssh"
	"net"
	"net/http"
	gourl "net/url"
	"os"
	"strings"
	"time"
)

//================================TCPDial================================

// TCP 连接
func TCP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial(io.TCP, addr)
}

// WithTCP 连接函数
func WithTCP(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return TCP(addr) }
}

// NewTCP 新建TCP连接
func NewTCP(addr string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithTCP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

// RedialTCP 一直连接TCP服务端,并重连
func RedialTCP(addr string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithTCP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

//================================UDPDial================================

// UDP 连接
func UDP(addr string) (io.ReadWriteCloser, error) {
	return net.Dial("udp", addr)
}

// WithUDP 连接函数
func WithUDP(addr string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return UDP(addr) }
}

func NewUDP(addr string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithUDP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

// RedialUDP 一直连接UDP服务端,并重连
func RedialUDP(addr string, options ...io.OptionClient) *io.Client {
	return io.Redial(WithUDP(addr), func(c *io.Client) {
		c.SetKey(addr)
		c.SetOptions(options...)
	})
}

var udpMap *maps.Safe

func WriteUDP(addr string, p []byte) error {
	if udpMap == nil {
		udpMap = maps.NewSafe()
	}
	v := udpMap.GetInterface(addr)
	if v == nil {
		c, err := net.Dial("udp", addr)
		if err != nil {
			return err
		}
		udpMap.Set(addr, c)
		v = c
	}
	c := v.(net.Conn)
	_, err := c.Write(p)
	return err
}

//================================FileDial================================

// File 打开文件
func File(path string) (io.ReadWriteCloser, error) {
	return os.Open(path)
}

// WithFile 打开文件函数
func WithFile(path string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return os.Open(path)
	}
}

func NewFile(path string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithFile(path), func(c *io.Client) {
		c.SetKey(path)
		c.SetOptions(options...)
	})
}

//================================WebsocketDial================================

// Memory 内存
func Memory(key string) (io.ReadWriteCloser, error) {
	s := common.MemoryServerManage.MustGet(key)
	if s == nil {
		return nil, errors.New("服务不存在")
	}
	return s.(*common.MemoryServer).Connect()
}

func WithMemory(key string) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return Memory(key)
	}
}

func NewMemory(key string, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithMemory(key), func(c *io.Client) {
		c.SetKey(key)
		c.SetOptions(options...)
	})
}

//================================MQTTDial================================

type MQTTConfig = mqtt.ClientOptions

var NewMQTTOptions = mqtt.NewClientOptions()

func MQTT(subscribe, publish string, qos byte, cfg *MQTTConfig) (io.ReadWriteCloser, error) {
	c := mqtt.NewClient(cfg)
	token := c.Connect()
	if !token.WaitTimeout(cfg.WriteTimeout) {
		return nil, errors.New("连接超时")
	}
	if token.Error() != nil {
		return nil, token.Error()
	}
	r := &MQTTClient{
		Client:    c,
		subscribe: subscribe,
		publish:   publish,
		qos:       qos,
		ch:        make(chan mqtt.Message, 1000),
	}
	c.Subscribe(subscribe, qos, func(client mqtt.Client, message mqtt.Message) {
		r.ch <- message
	})
	return r, token.Error()
}

func WithMQTT(clientID, topic string, qos byte, cfg *MQTTConfig) func() (io.ReadWriteCloser, error) {
	return func() (closer io.ReadWriteCloser, err error) {
		return MQTT(clientID, topic, qos, cfg)
	}
}

func NewMQTT(clientID, topic string, qos byte, cfg *MQTTConfig) (*io.Client, error) {
	c, err := io.NewDial(WithMQTT(clientID, topic, qos, cfg))
	if err == nil {
		c.SetKey(topic)
	}
	return c, err
}

func RedialMQTT(clientID, topic string, qos byte, cfg *MQTTConfig, options ...io.OptionClient) *io.Client {
	//cfg.SetAutoReconnect(false)
	return io.Redial(WithMQTT(clientID, topic, qos, cfg), func(c *io.Client) {
		c.SetKey(topic)
		c.SetOptions(options...)
	})
}

type MQTTClient struct {
	mqtt.Client
	subscribe string
	publish   string
	qos       byte
	ch        chan mqtt.Message
}

func (this *MQTTClient) Read(p []byte) (int, error) {
	return 0, io.ErrUseReadMessage
}

func (this *MQTTClient) ReadMessage() ([]byte, error) {
	msg := <-this.ch
	defer msg.Ack()
	return msg.Payload(), nil
}

func (this *MQTTClient) Write(p []byte) (int, error) {
	token := this.Client.Publish(this.publish, this.qos, false, p)
	token.Wait()
	return len(p), token.Error()
}

func (this *MQTTClient) Close() error {
	token := this.Client.Unsubscribe(this.subscribe)
	token.Wait()
	return token.Error()
}

//================================RabbitmqDial================================

//================================WebsocketDial================================

// Websocket 连接
func Websocket(url string, header http.Header) (io.MessageReadWriteCloser, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, header)
	return &WebsocketClient{Conn: c}, err
}

func WithWebsocket(url string, header http.Header) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		c, _, err := websocket.DefaultDialer.Dial(url, header)
		return &WebsocketClient{Conn: c}, err
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

//================================SSH================================

type SSHConfig struct {
	Addr        string
	User        string
	Password    string //类型为password
	Timeout     time.Duration
	High        int    //高
	Wide        int    //宽
	Term        string //样式
	ECHO        uint32 // 禁用回显（0禁用，1启动）
	Type        string //password 或者 key
	key         string //类型为key
	keyPassword string //类型为key
}

func (this *SSHConfig) new() *SSHConfig {
	if !strings.Contains(this.Addr, ":") {
		this.Addr += ":22"
	}
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
	if len(this.Term) == 0 {
		this.Term = "xterm-256color"
	}
	return this
}

type SSHClient struct {
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
	sshClient, err := ssh.Dial(io.TCP, cfg.Addr, config)
	if err != nil {
		return nil, err
	}
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          cfg.ECHO, // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400,    // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400,    //output speed = 14.4kbaud
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
	if err := session.RequestPty(cfg.Term, cfg.High, cfg.Wide, modes); err != nil {
		return nil, err
	}
	if err := session.Shell(); err != nil {
		return nil, err
	}
	return &SSHClient{
		Writer:  writer,
		Reader:  reader,
		Session: session,
		err:     outputErr,
	}, nil
}

func WithSSH(cfg *SSHConfig) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return SSH(cfg)
	}
}

func NewSSH(cfg *SSHConfig, options ...io.OptionClient) (*io.Client, error) {
	c, err := io.NewDial(WithSSH(cfg))
	if err == nil {
		c.SetKey(cfg.Addr).SetOptions(options...)
	}
	return c, err
}

func RedialSSH(cfg *SSHConfig, options ...io.OptionClient) *io.Client {
	return io.Redial(WithSSH(cfg), func(c *io.Client) {
		c.SetKey(cfg.Addr)
		c.SetOptions(options...)
	})
}

//================================OtherDial================================
