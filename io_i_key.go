package io

func NewIKey(key string) *IKey {
	return &IKey{key: key}
}

type IKey struct {
	key string
}

// SetKey 设置唯一标识
func (this *IKey) SetKey(key string) {
	this.key = key
}

// GetKey 获取唯一标识
func (this *IKey) GetKey() string {
	return this.key
}
