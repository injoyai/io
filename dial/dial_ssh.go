package dial

import (
	"github.com/injoyai/io"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
)

//================================SSH================================

type SSHConfig struct {
	Addr          string
	User          string
	Password      string //类型为password
	Timeout       time.Duration
	High          int               //高
	Wide          int               //宽
	Term          string            //样式 xterm
	Type          string            //password 或者 key
	key           string            //类型为key
	keyPassword   string            //类型为key
	Network       string            //tcp,udp ,可选,默认tcp
	TerminalModes ssh.TerminalModes //可选
}

func (this *SSHConfig) new() *SSHConfig {
	if !strings.Contains(this.Addr, ":") {
		this.Addr += ":22"
	}
	if len(this.User) == 0 {
		this.User = "root"
	}
	if this.Timeout == 0 {
		this.Timeout = time.Second
	}
	if this.High == 0 {
		this.High = 32
	}
	if this.Wide == 0 {
		this.Wide = 300
	}
	if len(this.Term) == 0 {
		this.Term = "xterm-256color"
	}
	if len(this.Network) == 0 {
		this.Network = io.TCP
	}
	if this.TerminalModes[ssh.TTY_OP_ISPEED] == 0 {
		this.TerminalModes[ssh.TTY_OP_ISPEED] = 14400 //input speed = 14.4kbaud
	}
	if this.TerminalModes[ssh.TTY_OP_OSPEED] == 0 {
		this.TerminalModes[ssh.TTY_OP_OSPEED] = 14400 //output speed = 14.4kbaud
	}
	if this.TerminalModes[ssh.ECHO] == 0 {
		//禁用回显（0禁用，1启动）
	}
	return this
}

type SSHClient struct {
	io.Writer
	io.Reader
	*ssh.Session
	err io.Reader
}

func SSH(cfg *SSHConfig) (io.ReadWriteCloser, error) {
	cfg.new()
	config := &ssh.ClientConfig{
		Timeout:         cfg.Timeout,
		User:            cfg.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(cfg.Password)},
	}
	switch cfg.Type {
	case "key":
		signer, err := ssh.ParsePrivateKeyWithPassphrase([]byte(cfg.key), []byte(cfg.keyPassword))
		if err != nil {
			return nil, err
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}
	sshClient, err := ssh.Dial(cfg.Network, cfg.Addr, config)
	if err != nil {
		return nil, err
	}
	session, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	reader, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	outputErr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}
	writer, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := session.RequestPty(cfg.Term, cfg.High, cfg.Wide, cfg.TerminalModes); err != nil {
		return nil, err
	}
	if err := session.Shell(); err != nil {
		return nil, err
	}
	return &SSHClient{
		Writer:  writer,
		Reader:  reader,
		Session: session,
		err:     outputErr,
	}, nil
}

func WithSSH(cfg *SSHConfig) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) {
		return SSH(cfg)
	}
}

func NewSSH(cfg *SSHConfig, options ...io.OptionClient) (*io.Client, error) {
	c, err := io.NewDial(WithSSH(cfg))
	if err == nil {
		c.SetKey(cfg.Addr).SetOptions(options...)
	}
	return c, err
}

func RedialSSH(cfg *SSHConfig, options ...io.OptionClient) *io.Client {
	return io.Redial(WithSSH(cfg), func(c *io.Client) {
		c.SetKey(cfg.Addr)
		c.SetOptions(options...)
	})
}
