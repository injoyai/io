package dial

import (
	"errors"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/io"
)

//================================MQTT================================

type MQTTConfig = mqtt.ClientOptions

// NewMQTTOptions 新建默认配置信息
func NewMQTTOptions() *mqtt.ClientOptions {
	return mqtt.NewClientOptions()
}

func MQTT(subscribe, publish string, qos byte, cfg *MQTTConfig) (io.ReadWriteCloser, error) {
	if cfg == nil {
		cfg = NewMQTTOptions()
	}
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
