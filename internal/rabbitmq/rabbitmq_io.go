package rabbitmq

import (
	"context"
	"github.com/injoyai/io"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Address    string
	Name       string //队列名称
	Durable    bool   //是否持久化，true为是。持久化会把队列存盘，服务器重启后，不会丢失队列以及队列内的信息。（注：1、不丢失是相对的，如果宕机时有消息没来得及存盘，还是会丢失的。2、存盘影响性能。）
	AutoDelete bool   //是否自动删除，true为是。至少有一个消费者连接到队列时才可以触发。当所有消费者都断开时，队列会自动删除。
	Exclusive  bool   //是否设置排他，true为是。如果设置为排他，则队列仅对首次声明他的连接可见，并在连接断开时自动删除。（注意，这里说的是连接不是信道，相同连接不同信道是可见的）。
	Nowait     bool   //是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
	Debug      bool   //调试模式
}

type Client struct {
	//cfg     *Config
	conn    *amqp.Connection
	channel *amqp.Channel
	ctx     context.Context
}

func (this *Client) Closed() bool {
	return this.conn.IsClosed()
}

func (this *Client) ReadWriteCloser(cfg *Config) (io.AReadWriteCloser, error) {
	r, err := this.AckReader(cfg)
	if err != nil {
		return nil, err
	}
	w, err := this.Writer(cfg)
	if err != nil {
		return nil, err
	}
	return io.NewAReadWriteCloser(r, w, this.conn), nil
}

func (this *Client) AckReader(cfg *Config) (io.AReader, error) {
	queue, err := this.channel.QueueDeclare(cfg.Name, cfg.Durable, cfg.AutoDelete, cfg.Exclusive, cfg.Nowait, nil)
	if err != nil {
		return nil, err
	}
	return &AckReader{
		Client: this,
		queue:  queue,
	}, nil
}

func (this *Client) Writer(cfg *Config) (io.Writer, error) {
	queue, err := this.channel.QueueDeclare(cfg.Name, cfg.Durable, cfg.AutoDelete, cfg.Exclusive, cfg.Nowait, nil)
	if err != nil {
		return nil, err
	}
	return &Writer{
		Client: this,
		queue:  queue,
	}, nil
}

type AckReader struct {
	*Client
	queue amqp.Queue
}

func (this *AckReader) ReadAck() (io.Acker, error) {
	//获取一个消息
	msg, ok, err := this.channel.Get(this.queue.Name, false)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &Message{&msg}, nil
}

type Writer struct {
	*Client
	queue amqp.Queue
}

func (this *Writer) Write(p []byte) (int, error) {
	/*
	   exchange：要发送到的交换机名称，对应图中exchangeName。
	   key：路由键，对应图中RoutingKey。
	   mandatory：直接false，不建议使用，后面有专门章节讲解。
	   immediate ：直接false，不建议使用，后面有专门章节讲解。
	   msg：要发送的消息，msg对应一个Publishing结构，Publishing结构里面有很多参数，这里只强调几个参数，其他参数暂时列出，但不解释。
	*/
	err := this.channel.PublishWithContext(this.ctx, "", this.queue.Name, false, false, amqp.Publishing{
		ContentType:     "text/plain", //消息的类型，通常为“text/plain”
		ContentEncoding: "",           //消息的编码，一般默认不用写
		DeliveryMode:    0,            //消息是否持久化，2表示持久化，0或1表示非持久化。
		Priority:        0,            //消息的优先级 0 to 9
		Body:            p,
	})
	return len(p), err
}

type Message struct {
	*amqp.Delivery
}

func (this *Message) Payload() []byte {
	return this.Delivery.Body
}

func (this *Message) Ack() error {
	//当参数为true时,标识多重确认,会确认之前所有同一队列的未确认
	return this.Delivery.Ack(false)
}
