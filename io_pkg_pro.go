package io

type Pro struct {
	MsgID    uint8  `json:"msgID"`    //消息id,用来确定是哪个命令,循环使用
	Total    uint8  `json:"total"`    //总包数,数据做了分包才有效
	Order    uint8  `json:"order"`    //包序号,分包序号,例如UDP分片,
	Protocol uint16 `json:"protocol"` //协议类型,0是无协议(自定义),1是子包
	Data     []byte `json:"data"`     //数据

}
