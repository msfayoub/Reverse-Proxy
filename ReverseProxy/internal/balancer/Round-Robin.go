package balancer

import (
    "net/url"
    "sync"
    "ReverseProxy/internal/pool"
)

type RoundRobinBalancer struct {
    pool   *pool.ServerPool
    index  int
    mu     sync.Mutex
}

func NewRoundRobinBalancer(pool *pool.ServerPool) *RoundRobinBalancer {
    return &RoundRobinBalancer{pool: pool}
}

func (b *RoundRobinBalancer) GetNextValidPeer() *pool.Backend {
    b.mu.Lock()
    defer b.mu.Unlock()
    backends := b.pool.GetBackends()
    if len(backends) == 0 {
        return nil
    }
    for i := 0; i < len(backends); i++ {
        b.index = (b.index + 1) % len(backends)
        if backends[b.index].Alive {
            return backends[b.index]
        }
    }
    return nil
}

func (b *RoundRobinBalancer) AddBackend(backend *pool.Backend) {
    b.pool.AddBackend(backend)
}

func (b *RoundRobinBalancer) SetBackendStatus(uri *url.URL, alive bool) {
    b.pool.SetBackendStatus(uri, alive)
}