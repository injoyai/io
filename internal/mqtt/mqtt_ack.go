package mqtt

import mqtt "github.com/eclipse/paho.mqtt.golang"

type Message struct {
	mqtt.Message
}

func (this *Message) Ack() error {
	this.Message.Ack()
	return nil
}
