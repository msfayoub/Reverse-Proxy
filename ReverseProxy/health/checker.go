package health

import (
	"ReverseProxy/internal/pool"
	"context"
	"log"
	"net/http"
	"time"
)

type Checker struct {
	Pool     *pool.ServerPool
	Interval time.Duration
	Client   *http.Client
}

func (c *Checker) Start(ctx context.Context) {
	if c.Pool == nil {
		panic("health.Checker: Pool is nil")  //The panic built-in function stops normal execution of the current goroutine.
	}
	if c.Interval <= 0 {
		c.Interval = 30 * time.Second //  Default to 30 seconds
	}
	if c.Client == nil {
		c.Client = &http.Client{Timeout: 2 * time.Second}
	}

	c.checkAll(ctx)  //First check immediately

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done(): 
			return
		case <-ticker.C: 
			c.checkAll(ctx)
		}
	}
}

func (c *Checker) checkAll(ctx context.Context) {
	backends := c.Pool.Snapshot()

	for _, b := range backends {
		u := b.GetURL() 
		alive := c.ping(ctx, u)
		prev := b.IsAlive()
		if alive != prev {
			c.Pool.SetBackendStatus(u, alive)
			if alive {
				log.Printf("[health] BACKEND UP: %s", u.String())
			} else {
				log.Printf("[health] BACKEND DOWN: %s", u.String())
			}
		}
	}
}

func (c *Checker) ping(parent context.Context, target stringerURL) bool {
	ctx, cancel := context.WithTimeout(parent, c.Client.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return false
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return false 
	}
	defer resp.Body.Close() 

	return resp.StatusCode < 500
}

type stringerURL interface {
	String() string
}
