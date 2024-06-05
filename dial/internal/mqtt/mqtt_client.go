package client

import (
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/io"
	"sync"
)

var (
	cacheMQTT = make(map[string]*Client)
	cacheMu   sync.RWMutex
)

type Config = mqtt.ClientOptions

func NewSubscribe(cfg *Config, topic string, qos byte) (*Subscribe, error) {
	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return c.Subscribe(topic, qos)
}

func NewWriter(cfg *Config, topic string, qos byte, retained bool) (*Write, error) {
	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return c.Writer(topic, qos, retained), nil
}

func NewClient(cfg *Config) (*Client, error) {
	if len(cfg.Servers) == 0 {
		return nil, errors.New("未设置服务器地址")
	}
	key := fmt.Sprintf("%s#%s", cfg.Servers[0].String(), cfg.ClientID)

	//判断缓存是否存在
	cacheMu.RLock()
	client, ok := cacheMQTT[key]
	cacheMu.RUnlock()
	if ok {
		//todo 这里是返回错误还是返回缓存
		return client, nil
	}

	c := mqtt.NewClient(cfg)
	token := c.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}
	return &Client{
		Client:    c,
		key:       fmt.Sprintf("%p", c),
		subscribe: make(map[string]*Subscribe),
	}, nil
}

type Client struct {
	mqtt.Client
	key       string
	write     map[string]*Write
	subscribe map[string]*Subscribe
	mu        sync.RWMutex
}

func (this *Client) closeWrite(key string) error {
	this.mu.Lock()
	delete(this.write, key)
	this.mu.Unlock()

	if len(this.subscribe) == 0 && len(this.write) == 0 {
		//如果客户端没有订阅,则关闭客户端
		this.Client.Disconnect(0)
		cacheMu.Lock()
		delete(cacheMQTT, this.key)
		cacheMu.Unlock()
	}

	return nil
}

func (this *Client) closeSubscribe(topic string) error {

	token := this.Client.Unsubscribe(topic)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	this.mu.Lock()
	delete(this.subscribe, topic)
	this.mu.Unlock()

	if len(this.subscribe) == 0 && len(this.write) == 0 {
		//如果客户端没有订阅,则关闭客户端
		this.Client.Disconnect(0)
		cacheMu.Lock()
		delete(cacheMQTT, this.key)
		cacheMu.Unlock()
	}

	return nil
}

// Subscribe todo 重复订阅会怎么样,待测试,是否需要取消之前的重复订阅
func (this *Client) Subscribe(topic string, qos byte) (*Subscribe, error) {
	ms := &Subscribe{
		Client:         this,
		subscribeTopic: topic,
		subscribeQos:   qos,
		messageChan:    make(chan mqtt.Message),
	}
	token := this.Client.Subscribe(topic, qos, func(client mqtt.Client, message mqtt.Message) {
		ms.messageChan <- message
	})
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}
	this.mu.Lock()
	this.subscribe[topic] = ms
	this.mu.Unlock()
	return ms, nil
}

func (this *Client) Writer(topic string, qos byte, retained bool) *Write {
	w := &Write{
		Client:   this,
		Topic:    topic,
		Qos:      qos,
		Retained: retained,
	}
	w.key = fmt.Sprintf("%p", w)

	this.mu.Lock()
	defer this.mu.Unlock()
	this.write[topic] = w
	return w
}

func (this *Client) Close() error {
	this.mu.Lock()
	for _, v := range this.subscribe {
		v.Close()
	}
	this.subscribe = nil
	for _, v := range this.write {
		v.Close()
	}
	this.write = nil
	this.mu.Unlock()
	this.Client.Disconnect(0)
	cacheMu.Lock()
	delete(cacheMQTT, this.key)
	cacheMu.Unlock()
	return nil
}

type Subscribe struct {
	Client         *Client
	subscribeTopic string
	subscribeQos   byte
	publishConfig  sync.Map
	messageChan    chan mqtt.Message
}

func (this *Subscribe) IO(topic string, qos byte, retained bool) io.ReadWriteCloser {
	w := this.Client.Writer(topic, qos, retained)
	return struct {
		io.Reader
		io.Writer
		io.Closer
	}{
		Reader: this,
		Writer: w,
		Closer: io.CloseFunc(func() error {
			w.Close()
			return this.Close()
		}),
	}
}

func (this *Subscribe) SetPublishConfig(topic string, publishConfig *PublishConfig) {
	this.publishConfig.Store(topic, publishConfig)
}

// Read 实现io.Reader接口
func (this *Subscribe) Read(p []byte) (int, error) {
	return 0, io.ErrUseReadMessage
}

// ReadAck 实现AckReader接口
func (this *Subscribe) ReadAck() (io.Acker, error) {
	msg := <-this.messageChan
	return msg, nil
}

// Publish 实现io.Publisher接口
func (this *Subscribe) Publish(topic string, bs []byte) error {
	var qos byte
	var retained bool
	if c, ok := this.publishConfig.Load(topic); ok && c != nil {
		publishConfig := c.(*PublishConfig)
		qos = publishConfig.Qos
		retained = publishConfig.Retained
	}
	token := this.Client.Publish(topic, qos, retained, bs)
	token.Wait()
	return token.Error()
}

// Close 实现io.Closer接口
func (this *Subscribe) Close() error {
	return this.Client.closeSubscribe(this.subscribeTopic)
}

func (this *Subscribe) Closed() bool {
	cacheMu.RLock()
	c, ok := cacheMQTT[this.Client.key]
	cacheMu.RUnlock()
	if ok {
		return !c.IsConnected()
	}
	return true
}

type Write struct {
	Client   *Client
	key      string
	Topic    string
	Qos      uint8
	Retained bool
}

func (this *Write) Write(p []byte) (int, error) {
	token := this.Client.Publish(this.Topic, this.Qos, this.Retained, p)
	token.Wait()
	return len(p), token.Error()
}

func (this *Write) Close() error {
	return this.Client.closeWrite(this.key)
}

type PublishConfig struct {
	Qos      byte
	Retained bool
}

type IO struct {
	*Subscribe
	*Write
}

func (this *IO) Close() error {
	this.Write.Close()
	return this.Subscribe.Close()
}
