package proxy

import (
	"ReverseProxy/internal/balancer"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"syscall"
	"time"
)

type ProxyHandler struct {
	LB         balancer.LoadBalancer // This will chose the backend to forward the request to
	Transport  http.RoundTripper     // Controls how HTTP requests are sent (timeouts, pooling)
	ReqTimeout time.Duration         // Maximum time allowed for a request
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.LB == nil {
		http.Error(w, "load balancer not configured", http.StatusInternalServerError) // If no load balancer is set, return HTTP 500.
		return
	}

	backend := h.LB.GetNextValidPeer()
	if backend == nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable) // If none, return HTTP 503.
		return
	}

	// If a timeout is set, wrap the request context with a timeout.
	req := r
	if h.ReqTimeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), h.ReqTimeout)
		defer cancel()
		req = r.Clone(ctx)
	}

	// Robust connection counter management: finish() is executed only once
	var done uint32
	finish := func() {
		if atomic.CompareAndSwapUint32(&done, 0, 1) { // ensure this runs only once
			backend.DecConn()
		}
	}
	backend.IncConn()

	target := backend.GetURL()
	rp := httputil.NewSingleHostReverseProxy(target)

	// Shared transport (timeouts + keep-alive)
	if h.Transport != nil {
		rp.Transport = h.Transport
	}

	// On success (response received), call finish() immediately
	rp.ModifyResponse = func(resp *http.Response) error {
		finish() // decrement connection count as soon as response is received
		return nil
	}

	// finish() + handling
	rp.ErrorHandler = func(rw http.ResponseWriter, req2 *http.Request, err error) {
		finish()

		// If the client cancels (disconnects), this is not a backend failure
		if errors.Is(err, context.Canceled) || errors.Is(req2.Context().Err(), context.Canceled) {
			return
		}

		// Timeout global -> 504
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(req2.Context().Err(), context.DeadlineExceeded) {
			http.Error(rw, "gateway timeout", http.StatusGatewayTimeout) // 504
			return
		}

		// Connexion refusée -> backend DOWN
		if isConnRefused(err) {
			h.LB.SetBackendStatus(target, false)
		}

		// Erreur générique -> 502
		http.Error(rw, "bad gateway", http.StatusBadGateway)
	}

	// Forward
	rp.ServeHTTP(w, req)
}


// isConnRefused detects "connection refused" even if the error is wrapped
func isConnRefused(err error) bool {
	if err == nil {
		return false
	}

	// Unwrap *url.Error
	var uerr *url.Error
	if errors.As(err, &uerr) && uerr.Err != nil {
		err = uerr.Err
	}

	// Unwrap *net.OpError
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Err != nil {
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) {
			return true
		}
	}

	return errors.Is(err, syscall.ECONNREFUSED)
}
