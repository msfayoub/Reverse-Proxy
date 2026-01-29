package admin

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"ReverseProxy/internal/pool"
)

type AdminHandler struct {
	Pool *pool.ServerPool
}

func (h *AdminHandler) Handler() http.Handler {
	if h.Pool == nil {
		panic("Pool in nil")
	}
	mux:= http.NewServeMux()
	mux.HandleFunc("/Backends", h.backendHandler)
	mux.HandleFunc("/status", h.statusHandler)
	return mux
}

func (h *AdminHandler) statusHandler(w http.ResponseWriter, r *http.Request) {
	if (r.Method != http.MethodGet) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	snap:= h.Pool.Snapshot()
	response:= StatusResponse{
		TotalBackends:  len(snap),
		ActiveBackends: h.Pool.CountAlive(),
		Backends:       make([]BackendStatus, 0, len(snap)),
	}
	for _, backend := range snap {
		response.Backends= append(response.Backends, BackendStatus{
			URL:                backend.GetURL().String(),
			Alive:              backend.Alive,
			CurrentConnections: backend.GetConns(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)

}

func (h *AdminHandler) backendHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handleAddBackend(w, r)
	case http.MethodDelete:
		h.handleRemoveBackend(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *AdminHandler) handleAddBackend(w http.ResponseWriter, r *http.Request) {
	var req AddBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid request body"})
		return
	}
	raw := strings.TrimSpace(req.URL)
	if raw == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "url is required"})
		return
	}
	u, err:=url.Parse(raw);
	if err != nil || u.Scheme == "" || u.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid url"})
		return
	}
	h.Pool.AddBackend(&pool.Backend{URL: u, Alive: true})
	w.WriteHeader(http.StatusCreated)

}




func (h *AdminHandler) handleRemoveBackend(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))

	if raw == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "url is required"})
		return
	}

	u, err:=url.Parse(raw);
	if err != nil || u.Scheme == "" || u.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid url"})
		return
	}

	removed := h.Pool.RemoveBackend(u)
	if !removed {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "backend not found"})
		return
	}
	w.WriteHeader(http.StatusOK)
}
