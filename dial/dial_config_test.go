package dial

import (
	"context"
	"testing"
	"time"
)

func TestWithConfig(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-time.After(time.Second * 20)
		t.Log("cancel")
		cancel()
	}()
	c, err := WithConfigContext(ctx, &Config{
		Dial: WithTCP(":10086"),
	})
	if err != nil {
		t.Error(err)
		return
	}
	c.GoTimerWriteString(time.Second*3, "666")
	c.Run()
}
