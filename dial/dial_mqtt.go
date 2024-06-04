package dial

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"io/ioutil"
	"strings"
	"time"
)

//================================MQTT================================

type MQTTConfig = mqtt.ClientOptions

// NewMQTTConfig 新建默认配置信息
func NewMQTTConfig() *MQTTConfig {
	return mqtt.NewClientOptions().SetAutoReconnect(false)
}

func MQTT(cfg *MQTTConfig, topic *MQTTTopic) (io.ReadWriteCloser, string, error) {
	if cfg == nil {
		cfg = NewMQTTConfig()
	}
	c := mqtt.NewClient(cfg)
	token := c.Connect()
	if !token.WaitTimeout(cfg.ConnectTimeout) {
		return nil, "", io.ErrWithConnectTimeout
	}
	if token.Error() != nil {
		return nil, "", token.Error()
	}
	r := &MQTTClient{
		Client: c,
		topic:  topic,
		ch:     make(chan mqtt.Message, io.DefaultChannelSize),
	}
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
	return r, cfg.Servers[0].Host, err
}

func WithMQTT(cfg *MQTTConfig, topic *MQTTTopic) io.DialFunc {
	return func(ctx context.Context) (io.ReadWriteCloser, string, error) { return MQTT(cfg, topic) }
}

func NewMQTT(cfg *MQTTConfig, topic *MQTTTopic, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithMQTT(cfg, topic), options...)
}

func RedialMQTT(cfg *MQTTConfig, topic *MQTTTopic, options ...io.OptionClient) *io.Client {
	cfg.SetAutoReconnect(false)
	return io.Redial(WithMQTT(cfg, topic), options...)
}

type MQTTClient struct {
	mqtt.Client
	topic *MQTTTopic
	ch    chan mqtt.Message
}

func (this *MQTTClient) Read(p []byte) (int, error) {
	return 0, io.ErrUseReadMessage
}

func (this *MQTTClient) ReadMessage() ([]byte, error) {
	msg := <-this.ch
	if msg == nil {
		return nil, errors.New("已关闭")
	}
	defer msg.Ack()
	return msg.Payload(), nil
}

func (this *MQTTClient) ReadAck() (io.Acker, error) {
	msg := <-this.ch
	return msg, nil
}

func (this *MQTTClient) Write(p []byte) (int, error) {
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

func (this *MQTTClient) Close() error {
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

type MQTTTopic struct {
	Subscribe []MQTTSubscribe
	Publish   []MQTTPublish
}

type MQTTBaseConfig struct {
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

func (this *MQTTBaseConfig) init() {
	if !strings.HasPrefix(this.BrokerURL, "tcp://") {
		this.BrokerURL = "tcp://" + this.BrokerURL
	}
	if len(this.ClientID) == 0 {
		this.ClientID = conv.String(time.Now().UnixNano())
	}
}

func WithMQTTBase(cfg *MQTTBaseConfig) *MQTTConfig {
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

type MQTTPublish struct {
	Topic    string
	Qos      uint8
	Retained bool
}

type MQTTSubscribe struct {
	Topic string
	Qos   uint8
}
