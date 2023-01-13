package io

func NewClientKey(key string) *ClientKey {
	return &ClientKey{key: key}
}

type ClientKey struct {
	key string
}

// SetKey 设置唯一标识
func (this *ClientKey) SetKey(key string) {
	this.key = key
}

// GetKey 获取唯一标识
func (this *ClientKey) GetKey() string {
	return this.key
}
