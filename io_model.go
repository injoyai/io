package io

import (
	"encoding/json"
	"github.com/injoyai/base/g"
	"net/http"
)

type Model struct {
	Type string      `json:"type"`           //请求类型,例如测试连接ping,写入数据write... 推荐请求和响应通过code区分
	Code int         `json:"code,omitempty"` //请求结果,推荐 请求:0(或null)  响应: 200成功,500失败... 同http好记一点
	UID  string      `json:"uid,omitempty"`  //消息的唯一ID,例如UUID
	Data interface{} `json:"data,omitempty"` //请求响应的数据
	Msg  string      `json:"msg,omitempty"`  //消息
}

func (this *Model) String() string {
	return string(this.Bytes())
}

func (this *Model) Bytes() g.Bytes {
	bs, _ := json.Marshal(this)
	return bs
}

func (this *Model) IsSucc() bool {
	return this.Code == http.StatusOK
}

func (this *Model) IsFail() bool {
	return this.Code > 0 && this.Code != http.StatusOK
}

func (this *Model) IsRequest() bool {
	return this.Code == 0
}

func (this *Model) IsResponse() bool {
	return this.Code != 0
}

func (this *Model) Resp(code int, data interface{}, msg string) *Model {
	return &Model{
		Type: this.Type,
		Code: code,
		UID:  this.UID,
		Data: data,
		Msg:  msg,
	}
}
