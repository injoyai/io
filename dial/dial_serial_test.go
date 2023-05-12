package dial

import (
	"testing"
	"time"
)

func TestScanSerialWithTimeout(t *testing.T) {
	t.Log(ScanSerial("COM9", time.Millisecond*50, []byte("ping")))
}

func TestNewSerial(t *testing.T) {
	c, err := NewSerial(&SerialConfig{
		Address:  "COM9",
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   SerialParityNone,
		Timeout:  time.Millisecond * 100,
	})
	if err != nil {
		t.Error(err)
		return
	}

	c.Debug()
	c.Write([]byte("+++"))
	go c.Run()

	select {}
}
