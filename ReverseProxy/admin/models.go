package admin

type AddBackendRequest struct {
	URL string `json:"url"`
}

type BackendStatus struct {
	URL                string `json:"url"`
	Alive              bool   `json:"alive"`
	CurrentConnections int64  `json:"current_connections"`
}

type StatusResponse struct {
	TotalBackends  int            `json:"total_backends"`
	ActiveBackends int            `json:"active_backends"`
	Backends       []BackendStatus `json:"backends"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
