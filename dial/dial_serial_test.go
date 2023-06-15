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
		Address:  "COM6",
		BaudRate: 115200,
		Timeout:  0,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()
	c.Debug()
	c.SetPrintFunc(func(msg io.Message, tag ...string) {
		logs.Debug(tag, msg)
	})
	c.SetReadWithAll()
	t.Log(c.Run())

	select {}
}

func TestNewSerial2(t *testing.T) {
	c, err := NewSerial(&SerialConfig{
		Address:  "COM7",
		BaudRate: 115200,
	})
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()
	c.Debug()
	c.SetPrintFunc(func(msg io.Message, tag ...string) {
		logs.Debug(tag, msg)
	})
	c.SetReadWithAll()
	go c.Run()

	//_, err = c.Write([]byte("+++"))

	//logs.PrintErr(err)
	go func() {
		for {
			<-time.After(time.Second * 5)
			_, err = c.Write([]byte("777"))
			logs.PrintErr(err)
		}
	}()

	//<-time.After(time.Second * 20)

	select {}
}

func TestRedialSerialF8L10T(t *testing.T) {
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
			t.Log(c.WriteRead([]byte("123")))
			t.Log(c.Write([]byte("+++")))
		}()
	})
	defer c.CloseAll()
	select {}
}

func TestRedialSerialSL101(t *testing.T) {
	portNames, err := serial.GetPortsList()
	if err != nil {
		logs.Err(err)
		return
	}
	logs.Debug("串口列表:", portNames)

	c := RedialSerial(&SerialConfig{
		Address: "COM10",
		Timeout: time.Second * 10,
	}, func(c *io.Client) {
		c.Debug()
		c.SetPrintWithASCII()
	})
	defer c.CloseAll()
	select {}
}
