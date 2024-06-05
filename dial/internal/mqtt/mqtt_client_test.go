package client

import (
	"testing"
)

func TestNewSubscribe(t *testing.T) {
	sub, err := NewSubscribe(WithEasy(&EasyConfig{
		BrokerURL: "192.168.10.23:1883",
	}), "test", 0)
	if err != nil {
		t.Error(err)
		return
	}
	for {
		ack, err := sub.ReadAck()
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(ack.Payload()))
		if err := sub.Publish("ack", ack.Payload()); err != nil {
			t.Error(err)
		}
	}
}
