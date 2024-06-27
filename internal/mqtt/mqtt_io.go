package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"io/ioutil"
	"strings"
	"time"
)

var _ io.ReadWriteCloser = (*MQTT)(nil)

type Config = mqtt.ClientOptions

type MQTT struct {
	mqtt.Client
	topic *Topic
	ch    chan mqtt.Message
	io.Reader
}

func (this *MQTT) ReadAck() (io.Acker, error) {
	msg := <-this.ch
	return &Message{msg}, nil
}

func (this *MQTT) Write(p []byte) (int, error) {
	var err error
	for _, v := range this.topic.Publish {
		token := this.Client.Publish(v.Topic, v.Qos, v.Retained, p)
		token.Wait()
		if token.Error() != nil {
			err = token.Error()
		}
	}
	return len(p), err
}

func (this *MQTT) Close() error {
	var err error
	for _, v := range this.topic.Subscribe {
		token := this.Client.Unsubscribe(v.Topic)
		token.Wait()
		if token.Error() != nil {
			err = token.Error()
		}
	}
	this.Client.Disconnect(0)
	close(this.ch)
	return err
}

func New(cfg *Config, topic *Topic) (io.ReadWriteCloser, error) {
	c := mqtt.NewClient(cfg)
	token := c.Connect()
	if !token.WaitTimeout(cfg.ConnectTimeout) {
		return nil, io.ErrWithConnectTimeout
	}
	if token.Error() != nil {
		return nil, token.Error()
	}
	r := &MQTT{
		Client: c,
		topic:  topic,
		ch:     make(chan mqtt.Message, io.DefaultChannelSize),
	}
	r.Reader = io.AReaderToReader(r)
	var err error
	for _, v := range topic.Subscribe {
		token := c.Subscribe(v.Topic, v.Qos, func(client mqtt.Client, message mqtt.Message) {
			r.ch <- message
		})
		token.Wait()
		if token.Error() != nil {
			err = token.Error()
		}
	}
	return r, err
}

func NewEasy(cfg *EasyConfig, topic *Topic) (io.ReadWriteCloser, error) {
	cfg.init()
	return New(WithMQTTBase(cfg), topic)
}

type EasyConfig struct {
	BrokerURL string        //必选,不要忘记 tcp://
	ClientID  string        //必选,服务器Topic地址
	Username  string        //用户名
	Password  string        //密码
	Timeout   time.Duration //连接超时时间,
	KeepAlive time.Duration //心跳时间,0是不启用该机制

	TLS    bool
	CAFile string //server
	CCFile string // client-crt
	CKFile string // client-key
}

func (this *EasyConfig) init() {
	if !strings.HasPrefix(this.BrokerURL, "tcp://") {
		this.BrokerURL = "tcp://" + this.BrokerURL
	}
	if len(this.ClientID) == 0 {
		this.ClientID = conv.String(time.Now().UnixNano())
	}
}

func WithMQTTBase(cfg *EasyConfig) *Config {
	cfg.init()
	return mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetConnectTimeout(cfg.Timeout).
		SetKeepAlive(cfg.KeepAlive).
		SetAutoReconnect(false). //自动重连
		SetCleanSession(false).  //重连后恢复session
		SetTLSConfig(func() *tls.Config {
			if !cfg.TLS {
				return nil
			}
			certPool := x509.NewCertPool()
			ca, err := ioutil.ReadFile(cfg.CAFile)
			if err != nil {
				return nil
			}
			certPool.AppendCertsFromPEM(ca)
			clientKeyPair, err := tls.LoadX509KeyPair(cfg.CCFile, cfg.CKFile)
			if err != nil {
				return nil
			}
			return &tls.Config{
				RootCAs:            certPool,
				ClientAuth:         tls.NoClientCert,
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{clientKeyPair},
			}
		}())
}

/*



 */

type Topic struct {
	Subscribe []Subscribe
	Publish   []Publish
}

type Publish struct {
	Topic    string
	Qos      uint8
	Retained bool
}

type Subscribe struct {
	Topic string
	Qos   uint8
}
