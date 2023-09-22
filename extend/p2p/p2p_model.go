package p2p

const (
	TypePing = "ping"
	TypePong = "pong"
)

type Msg struct {
	Type string `json:"type"` //消息类型
	Data string `json:"data"` //消息数据
}

type MsgBase struct {
	Name       string `json:"name"`       //名称
	Version    string `json:"version"`    //版本信息
	StartTime  int64  `json:"startTime"`  //运行时间
	ConnectKey string `json:"connectKey"` //连接秘钥
}

// MsgFind 查找站点信息
type MsgFind struct {
}
