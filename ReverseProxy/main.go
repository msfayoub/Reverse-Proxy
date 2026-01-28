package main

import (
	"ReverseProxy/health"
	"ReverseProxy/internal/balancer"
	"ReverseProxy/internal/pool"
	"ReverseProxy/internal/proxy"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {  //I currently use config file to load backends and strategy
    Strategy string   `json:"strategy"`
    Backends []string `json:"backends"`
}

 // Ai generated from best practices
func NewDefaultTransport() *http.Transport {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
	}
}

func main() {
    file, err := os.Open("config.json")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    var cfg Config
    if err := json.NewDecoder(file).Decode(&cfg); err != nil {
        log.Fatal(err)
    }

    serverPool := pool.NewServerPool()
    for _, backendURL := range cfg.Backends {
        parsedURL, err := url.Parse(backendURL)
        if err != nil {
            log.Fatalf("Invalid backend URL: %s", backendURL)
        }
        serverPool.AddBackend(&pool.Backend{URL: parsedURL, Alive: true})
    
    }

    bal, err := balancer.New(cfg.Strategy, serverPool)
    if err != nil {
        log.Fatal(err)
    }

    rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	checker := &health.Checker{
		Pool:     serverPool,
		Interval: 30 * time.Second,
		Client:   &http.Client{Timeout: 2 * time.Second},
	}
	go checker.Start(rootCtx)

    handler := &proxy.ProxyHandler{LB: bal, Transport: NewDefaultTransport(), ReqTimeout: 5 * time.Second}
    http.Handle("/", handler)
    log.Println("Proxy started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}