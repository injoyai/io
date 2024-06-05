package client

import (
	"crypto/tls"
	"crypto/x509"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/base/g"
	"github.com/injoyai/io"
	"io/ioutil"
	"strings"
	"time"
)

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

func WithEasy(cfg *EasyConfig) *Config {
	if !strings.HasPrefix(cfg.BrokerURL, "tcp://") {
		cfg.BrokerURL = "tcp://" + cfg.BrokerURL
	}
	if len(cfg.ClientID) == 0 {
		cfg.ClientID = g.RandString(8)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = io.DefaultConnectTimeout
	}
	return mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetConnectTimeout(cfg.Timeout).
		SetKeepAlive(cfg.KeepAlive).
		SetAutoReconnect(false). //自动重连
		SetCleanSession(true).   //断开后清除session
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
