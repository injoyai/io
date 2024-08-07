package dial

import (
	"context"
	"github.com/goburrow/serial"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	serial2 "go.bug.st/serial"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//================================Serial================================

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
func Serial(cfg *SerialConfig) (io.ReadWriteCloser, string, error) {
	if cfg == nil {
		cfg = &SerialConfig{}
	}
	if cfg.Address == "" {
		//获取串口列表的第一个
		if list, _ := GetSerialPortList(); len(list) > 0 {
			cfg.Address = list[0]
		}
	}
	if cfg.BaudRate == 0 {
		cfg.BaudRate = 9600
	}
	if cfg.DataBits == 0 {
		cfg.DataBits = 8
	}
	if cfg.StopBits == 0 {
		cfg.StopBits = 1
	}
	if len(cfg.Parity) == 0 {
		cfg.Parity = SerialParityNone
	}
	c, err := serial.Open(cfg)
	return c, cfg.Address, err
}

// WithSerial 打开串口函数
func WithSerial(cfg *SerialConfig) io.DialFunc {
	return func(ctx context.Context) (io.ReadWriteCloser, string, error) { return Serial(cfg) }
}

func NewSerial(cfg *SerialConfig, options ...io.OptionClient) (*io.Client, error) {
	return io.NewDial(WithSerial(cfg), func(c *io.Client) {
		c.SetOptions(options...)
		exitChan := make(chan os.Signal)
		signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func(c *io.Client) {
			<-exitChan
			c.CloseAll()
			os.Exit(-127)
		}(c)
	})
}

func RedialSerial(cfg *SerialConfig, options ...io.OptionClient) *io.Client {
	c := io.Redial(WithSerial(cfg), options...)
	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c *io.Client) {
		<-exitChan
		c.CloseAll()
		os.Exit(-127)
	}(c)
	return c
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
		115200, 230400, 256000, 460800, 500000, 512000, 600000, 750000, 921600,
		1000000, 1500000, 2000000,
	}
}

// ScanSerial 扫描串口
func ScanSerial(addr string, timeout time.Duration, write []byte) (*SerialConfig, []byte) {
	for _, dataBits := range []int{8, 7, 6, 5} {
		for _, stopBits := range []int{1, 2} {
			for _, parity := range []string{SerialParityNone, SerialParityEven, SerialParityOdd} {
				for _, baudRate := range GetSerialBaudRate() {
					cfg, resp, err := func() (*SerialConfig, []byte, error) {
						cfg := &SerialConfig{
							Address:  addr,
							BaudRate: baudRate,
							DataBits: dataBits,
							StopBits: stopBits,
							Parity:   parity,
							Timeout:  timeout,
						}
						c, err := NewSerial(cfg)
						if err != nil {
							return nil, nil, err
						}
						defer c.Close()
						go c.Run()
						resp, err := c.WriteRead(write, timeout)
						if err != nil {
							return nil, nil, err
						}
						return cfg, resp, nil
					}()
					if err == nil {
						return cfg, resp
					}
					logs.Errf("%v %v\n", cfg, err)
				}
			}
		}
	}
	return nil, nil
}
