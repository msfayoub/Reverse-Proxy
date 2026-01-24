package balancer

import (
	"net/url"
	"ReverseProxy/internal/pool"
	"fmt"
)

type LoadBalancer interface {
	GetNextValidPeer() *pool.Backend
	AddBackend(backend *pool.Backend)
	SetBackendStatus(uri *url.URL, alive bool)
}

func New(strategy string, pool *pool.ServerPool) (LoadBalancer, error){
	switch strategy {
    case "round-robin":
        return NewRoundRobinBalancer(pool), nil
    case "least-conn":
        return NewLeastConnBalancer(pool), nil
    default:
        return nil, fmt.Errorf("unknown strategy: %s", strategy)
    }
}
