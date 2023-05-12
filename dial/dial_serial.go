package dial

import (
	"github.com/goburrow/serial"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	serial2 "go.bug.st/serial"
	"time"
)

//================================SerialDial================================

const (
	SerialParityNone = "N" //无校验
	SerialParityEven = "E" //奇校验
	SerialParityOdd  = "O" //偶校验
)

type (
	SerialConfig      = serial.Config
	SerialRS485Config = serial.RS485Config
)

// Serial 打开串口
func Serial(cfg *SerialConfig) (io.ReadWriteCloser, error) {
	return serial.Open(cfg)
}

// SerialFunc 打开串口函数
func SerialFunc(cfg *SerialConfig) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return serial.Open(cfg)
	}
}

func NewSerial(cfg *SerialConfig) (*io.Client, error) {
	c, err := io.NewDial(SerialFunc(cfg))
	if err == nil {
		c.SetKey(cfg.Address)
	}
	return c, err
}

func RedialSerial(cfg *SerialConfig, options ...io.OptionClient) *io.Client {
	return io.Redial(SerialFunc(cfg), func(c *io.Client) {
		c.SetKey(cfg.Address).SetOptions(options...)
	})
}

//================================SerialOther================================

// GetSerialPortList 获取当前串口列表
func GetSerialPortList() ([]string, error) {
	return serial2.GetPortsList()
}

// GetSerialBaudRate 获取波特率列表
func GetSerialBaudRate() []int {
	return []int{
		50, 75,
		110, 134, 150, 200, 300, 600,
		1200, 1800, 2400, 4800, 7200, 9600,
		14400, 19200, 28800, 38400, 57600, 76800,
		115200, 230400,
	}
}

// ScanSerial 扫描串口
func ScanSerial(addr string, timeout time.Duration, write []byte) (*SerialConfig, []byte) {
	for _, dataBits := range []int{8, 7, 6, 5} {
		for _, stopBits := range []int{1, 2} {
			for _, parity := range []string{SerialParityNone, SerialParityEven, SerialParityOdd} {
				for _, baudRate := range GetSerialBaudRate() {
					var x *SerialConfig
					cfg, resp, err := func() (*SerialConfig, []byte, error) {
						cfg := &SerialConfig{
							Address:  addr,
							BaudRate: baudRate,
							DataBits: dataBits,
							StopBits: stopBits,
							Parity:   parity,
							Timeout:  timeout,
						}
						x = cfg
						c, err := NewSerial(cfg)
						if err != nil {
							return nil, nil, err
						}
						defer c.Close()
						go c.Run()
						resp, err := c.WriteReadWithTimeout(write, timeout)
						if err != nil {
							return nil, nil, err
						}
						return cfg, resp, nil
					}()
					if err == nil {
						return cfg, resp
					}
					logs.Errf("%v %v", x, err)
				}
			}
		}
	}
	return nil, nil
}
