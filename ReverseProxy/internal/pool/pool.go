package pool

import (
	"net/url"
	"sync"
)
type ServerPool struct {
	Current uint64 `json:"current"`
	Backends []*Backend `json:"backends"`
	mux sync.RWMutex  // protects concurrent access to Current
}
func NewServerPool() *ServerPool {
    return &ServerPool{}
}

func (p *ServerPool) AddBackend(b *Backend){
	p.mux.Lock()
	p.Backends = append(p.Backends, b)
	p.mux.Unlock()
}

func (p *ServerPool) RemoveBackend(target *url.URL) bool{
	p.mux.Lock()
	defer p.mux.Unlock()
	for i, b := range p.Backends {
		if b.URL.String() == target.String() {
			p.Backends = append(p.Backends[:i], p.Backends[i+1:]...)
			return true
		}
	}
	return false
}

func (p *ServerPool) SetBackendStatus(target *url.URL, alive bool){
	p.mux.Lock()
	defer p.mux.Unlock()
	for i, b := range p.Backends {
		if b.URL.String() == target.String() {
			p.Backends[i].SetAlive(alive)
			return
		}
	}
	return 
}

func (p *ServerPool) GetBackends() []*Backend {
    p.mux.RLock()
    defer p.mux.RUnlock()
    return p.Backends
}


func (p *ServerPool) Snapshot() []*Backend {
	p.mux.RLock()
	defer p.mux.RUnlock()
	backends := make([]*Backend, len(p.Backends))
	copy(backends, p.Backends)
	return backends
}

func (p *ServerPool) CountAlive() int{
	p.mux.RLock()
	defer p.mux.RUnlock()
	count := 0
	for _, b := range p.Backends {
		if b.IsAlive() {
			count++
		}
	}
	return count
}

