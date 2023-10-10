package p2p

import "github.com/injoyai/io"

const (
	TypePing = io.Ping
	TypePong = io.Pong
)

type Msg struct {
	Code  int         `json:"code,omitempty"` //状态
	Type  string      `json:"type"`           //消息类型
	MsgID string      `json:"msgId"`          //消息标识
	Data  interface{} `json:"data,omitempty"` //消息数据
}

type MsgRegister struct {
	NodeID     string `json:"nodeID"`     //名称
	Version    string `json:"version"`    //版本信息
	StartTime  int64  `json:"startTime"`  //运行时间
	ConnectKey string `json:"connectKey"` //连接秘钥
	LocalAddr  string `json:"localAddr"`  //本地地址
	RemoteAddr string `json:"-"`          //远程地址
}

type MsgGetRegister struct {
	NodeID string `json:"nodeID"` //节点标识
}

type MsgConnectNotice struct {
	NodeID string `json:"nodeID"` //节点标识
}

type MsgConnect struct {
	NodeID string `json:"nodeID"` //名称
}

type MsgError struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}
