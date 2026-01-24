package balancer

import (
    "net/url"
    "ReverseProxy/internal/pool"
    "sync"
)

type LeastConnBalancer struct {
    pool *pool.ServerPool
    mu     sync.Mutex
}

func NewLeastConnBalancer(pool *pool.ServerPool) *LeastConnBalancer {
    return &LeastConnBalancer{pool: pool}
}

func (b *LeastConnBalancer) GetNextValidPeer() *pool.Backend {
    b.mu.Lock()
    defer b.mu.Unlock()
    backends := b.pool.GetBackends()
    var selected *pool.Backend
    for _, backend := range backends {
        if backend.Alive && (selected == nil || backend.CurrentConns < selected.CurrentConns) {
            selected = backend
        }
    }
    return selected
}

func (b *LeastConnBalancer) AddBackend(backend *pool.Backend) {
    b.pool.AddBackend(backend)
}

func (b *LeastConnBalancer) SetBackendStatus(uri *url.URL, alive bool) {
    b.pool.SetBackendStatus(uri, alive)
}