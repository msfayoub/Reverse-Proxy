package pool

import (
	"net/url"
	"sync"
)

type Backend struct {
	URL *url.URL  `json:"url"`
	Alive bool     `json:"alive"`
	CurrentConns int `json:"current_connections"`
	mux sync.RWMutex
}


func (b *Backend) SetAlive(alive bool){
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}

func (b *Backend) GetURL() *url.URL{
	return b.URL
}

func (b *Backend) IncConn() {
	b.mux.Lock()
	b.CurrentConns++
	b.mux.Unlock()
}

func (b *Backend) DecConn() {
	b.mux.Lock()
	b.CurrentConns--
	b.mux.Unlock()
}

func (b *Backend) GetConns() int64{
	b.mux.RLock()
	defer b.mux.RUnlock()
	return int64(b.CurrentConns)
}