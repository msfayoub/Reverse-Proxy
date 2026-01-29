package main

import (
	"ReverseProxy/health"
	"ReverseProxy/admin"
	"ReverseProxy/internal/balancer"
	"ReverseProxy/internal/pool"
	"ReverseProxy/internal/proxy"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Strategy string   `json:"strategy"`
	Backends []string `json:"backends"`
}

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
	// ---- load config ----
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	// ---- build pool ----
	serverPool := pool.NewServerPool()
	for _, backendURL := range cfg.Backends {
		parsedURL, err := url.Parse(backendURL)
		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
			log.Fatalf("Invalid backend URL: %s", backendURL)
		}
		serverPool.AddBackend(&pool.Backend{URL: parsedURL, Alive: true})
	}

	// ---- build balancer ----
	bal, err := balancer.New(cfg.Strategy, serverPool)
	if err != nil {
		log.Fatal(err)
	}

	// ---- root ctx for shutdown ----
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ---- health checker ----
	checker := &health.Checker{
		Pool:     serverPool,
		Interval: 30 * time.Second,
		Client:   &http.Client{Timeout: 2 * time.Second},
	}
	go checker.Start(rootCtx)

	// ---- proxy server (:8080) ----
	proxyHandler := &proxy.ProxyHandler{
		LB:         bal,
		Transport:  NewDefaultTransport(),
		ReqTimeout: 5 * time.Second,
	}

	proxyMux := http.NewServeMux()
	proxyMux.Handle("/", proxyHandler)

	proxySrv := &http.Server{
		Addr:              ":8080",
		Handler:           proxyMux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// ---- admin server (:8081) ----
	adminHandler := (&admin.AdminHandler{Pool: serverPool}).Handler()

	adminSrv := &http.Server{
		Addr:              ":8081",
		Handler:           adminHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// ---- run both servers concurrently ----
	errCh := make(chan error, 2)

	go func() {
		log.Println("[proxy] started on :8080")
		if err := proxySrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	go func() {
		log.Println("[admin] started on :8081")
		if err := adminSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// ---- wait for shutdown signal or fatal server error ----
	select {
	case <-rootCtx.Done():
		log.Println("[main] shutdown requested")
	case err := <-errCh:
		log.Printf("[main] server error: %v", err)
		stop()
	}

	// ---- graceful shutdown ----
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = adminSrv.Shutdown(ctx)
	_ = proxySrv.Shutdown(ctx)

	log.Println("[main] shutdown complete")
}
