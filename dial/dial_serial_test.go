package dial

import (
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"go.bug.st/serial"
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
		Timeout:  time.Second * 10,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	c.Debug()
	//go func() {
	//	for {
	//		_, err = c.Write([]byte("+++"))
	//		logs.PrintErr(err)
	//		<-time.After(time.Second)
	//	}
	//}()

	_, err = c.Write([]byte("+++"))
	go c.Run()
	logs.PrintErr(err)
	go func() {
		<-time.After(time.Second * 2)
		_, err = c.Write([]byte("++++"))
		logs.PrintErr(err)
	}()

	<-time.After(time.Second * 20)
}

func TestRedialSerial(t *testing.T) {
	portNames, err := serial.GetPortsList()
	if err != nil {
		logs.Err(err)
		return
	}
	logs.Debug("串口列表:", portNames)

	c := RedialSerial(&SerialConfig{
		Address:  "COM7",
		BaudRate: 115200,
		DataBits: 8,
		StopBits: 1,
		Parity:   SerialParityNone,
		Timeout:  time.Second * 10,
	}, func(c *io.Client) {
		c.Debug()
		c.SetPrintWithASCII()
		go func() {
			t.Log(c.WriteRead([]byte("+++")))
			t.Log(c.WriteRead([]byte("+++")))
			t.Log(c.WriteRead([]byte("AT+VER?\n")))
		}()
	})
	defer c.CloseAll()
	<-time.After(time.Second * 100)
}
