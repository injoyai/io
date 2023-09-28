package p2p

import "github.com/injoyai/io"

const (
	TypePing = io.Ping
	TypePong = io.Pong
)

type Msg struct {
	Type  string      `json:"type"`  //消息类型
	MsgID string      `json:"msgId"` //消息标识
	Data  interface{} `json:"data"`  //消息数据
}

type MsgRegister struct {
	Name       string `json:"name"`       //名称
	Version    string `json:"version"`    //版本信息
	StartTime  int64  `json:"startTime"`  //运行时间
	ConnectKey string `json:"connectKey"` //连接秘钥
	LocalAddr  string `json:"localAddr"`  //本地地址
}

type MsgGetPeer struct {
	RemoteAddr string `json:"remoteAddr"` //远程外网地址
}

// MsgFind 查找站点信息
type MsgFind struct {
}
