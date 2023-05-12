package io

import "sync"

/*
ClientManage
客户端统一管理
例如串口,需要统一
*/
type ClientManage struct {
	m  map[string]*Client
	mu sync.RWMutex
}

func (this *ClientManage) Close() error {
	this.CloseClientAll()
	return nil
}

// DialClient 连接客户端
func (this *ClientManage) DialClient(dialFunc DialFunc) (*Client, error) {
	c, err := NewDial(dialFunc)
	if err != nil {
		return nil, err
	}
	this.SetClient(c)
	return c, nil
}

// RedialClient 连接客户端直到成功
func (this *ClientManage) RedialClient(dial DialFunc, options ...OptionClient) *Client {
	c := Redial(dial, options...)
	this.SetClient(c)
	return c
}

// SetClient 添加客户端
func (this *ClientManage) SetClient(c *Client) {
	if c == nil {
		return
	}
	this.mu.Lock()
	defer this.mu.Unlock()

	old, ok := this.m[c.GetKey()]
	if ok && old.Pointer() != c.Pointer() {
		old.CloseAll()
	}
	this.m[c.GetKey()] = c
}

// GetClient 获取客户端
func (this *ClientManage) GetClient(key string) *Client {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.m[key]
}

// GetClientAny 获取任意一个客户端
func (this *ClientManage) GetClientAny() *Client {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for _, v := range this.m {
		return v
	}
	return nil
}

// GetClientDo 获取客户端并执行
func (this *ClientManage) GetClientDo(key string, fn func(c *Client) error) (bool, error) {
	c := this.GetClient(key)
	if c != nil {
		return true, fn(c)
	}
	return false, nil
}

// GetClientLen 获取客户端数量
func (this *ClientManage) GetClientLen() int {
	return len(this.m)
}

// GetClientMap 获取客户端map,元数据,注意安全
func (this *ClientManage) GetClientMap() map[string]*Client {
	return this.m
}

// CopyClientMap 复制所有客户端数据
func (this *ClientManage) CopyClientMap() map[string]*Client {
	m := make(map[string]*Client)
	this.RangeClient(func(key string, c *Client) bool {
		m[key] = c
		return true
	})
	return m
}

// RangeClient 遍历客户端
func (this *ClientManage) RangeClient(fn func(key string, c *Client) bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	for i, v := range this.m {
		if !fn(i, v) {
			break
		}
	}
}

// WriteClient 写入客户端数据
func (this *ClientManage) WriteClient(key string, p []byte) (bool, error) {
	return this.GetClientDo(key, func(c *Client) error {
		_, err := c.Write(p)
		return err
	})
}

// WriteClientAll 广播,发送数据给所有连接,加入到连接的队列
func (this *ClientManage) WriteClientAll(p []byte) {
	for _, c := range this.GetClientMap() {
		//写入到队列,避免阻塞
		c.WriteQueue(p)
	}
}

// TryWriteClientAll 广播,发送数据给所有连接,尝试加入到连接的队列
func (this *ClientManage) TryWriteClientAll(p []byte) {
	for _, c := range this.GetClientMap() {
		//写入到队列,避免阻塞,加入不了则丢弃数据
		c.TryWriteQueue(p)
	}
}

// WriteClientAny 写入任意一个客户端数据
func (this *ClientManage) WriteClientAny(p []byte) (int, error) {
	if c := this.GetClientAny(); c != nil {
		return c.Write(p)
	}
	return len(p), nil
}

// WriteReadClient 写入客户端并等待结果返回
func (this *ClientManage) WriteReadClient(key string, p []byte) ([]byte, bool, error) {
	var res []byte
	var err error
	has, err := this.GetClientDo(key, func(c *Client) error {
		res, err = c.WriteRead(p)
		return err
	})
	return res, has, err
}

// Write 给所有客户端发送数据,实现io.Writer接口
func (this *ClientManage) Write(p []byte) (int, error) {
	this.WriteClientAll(p)
	return len(p), nil
}

// CloseClient 关闭客户端,会重试
func (this *ClientManage) CloseClient(key string) error {
	if c := this.GetClient(key); c != nil {
		return c.CloseAll()
	}
	return nil
}

// CloseClientAll 关闭所有客户端
func (this *ClientManage) CloseClientAll() {
	this.mu.Lock()
	defer this.mu.Unlock()
	for _, v := range this.m {
		v.CloseAll()
	}
	this.m = make(map[string]*Client)
}

// SetClientKey 重命名key
func (this *ClientManage) SetClientKey(newClient *Client, newKey string) {
	//判断这个标识符的客户端是否存在,存在则关闭
	if oldClient := this.GetClient(newKey); oldClient != nil {
		//判断指针地址是否一致,不一致则关闭
		if oldClient.Pointer() != newClient.Pointer() {
			oldClient.CloseAll()
		}
	}
	//更新新的客户端
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.m, newClient.GetKey())
	this.m[newKey] = newClient.SetKey(newKey)
}
