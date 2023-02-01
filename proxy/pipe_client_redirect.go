package proxy

import "sync"

type IRedirect interface {
	Set(oldAddr, newAddr string)
	Get(addr string) string
}

func newRedirect() *redirect {
	return &redirect{m: make(map[string]string)}
}

type redirect struct {
	m  map[string]string
	mu sync.RWMutex
}

func (this *redirect) Set(oldAddr, newAddr string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.m[oldAddr] = newAddr
}

func (this *redirect) Get(addr string) string {
	this.mu.RLock()
	defer this.mu.RUnlock()
	if val, ok := this.m[addr]; ok {
		addr = val
	}
	return addr
}
