package dial

import (
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
	return mqtt.NewClientOptions()
}

func MQTT(iocfg *MQTTIOConfig, cfg *MQTTConfig) (io.ReadWriteCloser, string, error) {
	if cfg == nil {
		cfg = NewMQTTConfig()
	}
	c := mqtt.NewClient(cfg)
	token := c.Connect()
	if !token.WaitTimeout(cfg.ConnectTimeout) {
		return nil, "", errors.New("连接超时")
	}
	if token.Error() != nil {
		return nil, "", token.Error()
	}
	r := &MQTTClient{
		Client: c,
		iocfg:  iocfg,
		ch:     make(chan mqtt.Message, 1000),
	}
	c.Subscribe(iocfg.Subscribe, iocfg.SubscribeQos, func(client mqtt.Client, message mqtt.Message) {
		r.ch <- message
	})
	return r, cfg.Servers[0].Host, token.Error()
}

func WithMQTT(iocfg *MQTTIOConfig, cfg *MQTTConfig) io.DialFunc {
	return func() (io.ReadWriteCloser, string, error) { return MQTT(iocfg, cfg) }
}

func NewMQTT(iocfg *MQTTIOConfig, cfg *MQTTConfig, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithMQTT(iocfg, cfg), options...)
}

func RedialMQTT(iocfg *MQTTIOConfig, cfg *MQTTConfig, options ...io.OptionClient) *io.Client {
	return io.Redial(WithMQTT(iocfg, cfg), options...)
}

type MQTTClient struct {
	mqtt.Client
	iocfg *MQTTIOConfig
	ch    chan mqtt.Message
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
	token := this.Client.Publish(this.iocfg.Publish, this.iocfg.PublishQos, this.iocfg.Retained, p)
	token.Wait()
	return len(p), token.Error()
}

func (this *MQTTClient) Close() error {
	token := this.Client.Unsubscribe(this.iocfg.Subscribe)
	token.Wait()
	return token.Error()
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

type MQTTIOConfig struct {
	Subscribe    string
	SubscribeQos uint8
	Publish      string
	PublishQos   uint8
	Retained     bool
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
		SetAutoReconnect(true). //自动重连
		SetCleanSession(false). //重连后恢复session
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
