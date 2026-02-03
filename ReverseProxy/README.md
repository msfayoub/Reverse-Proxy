# Concurrent Load-Balancing Reverse Proxy

A Go-based reverse proxy server with health monitoring, dynamic backend management, and configurable load balancing strategies.

## Features

- **Load Balancing**: Round-Robin and Least-Connections strategies
- **Health Monitoring**: Automatic periodic health checks of backend servers
- **Dynamic Management**: Add/remove backends via REST API without restarting
- **Concurrent**: Thread-safe operations using proper synchronization
- **Graceful Shutdown**: Clean shutdown with context cancellation
- **Admin API**: Separate admin interface for management operations

## Project Structure

```
ReverseProxy/
├── main.go                 # Entry point
├── config.json             # Initial configuration
├── admin/
│   ├── handlers.go         # Admin API handlers
│   └── models.go           # Admin API models
├── Backends/
│   ├── backend1.go         # Test backend server 1
│   └── backend2.go         # Test backend server 2
├── health/
│   └── checker.go          # Health monitoring service
└── internal/
    ├── balancer/
    │   ├── interface.go    # LoadBalancer interface
    │   ├── Round-Robin.go  # Round-robin implementation
    │   └── leastconn.go    # Least-connections implementation
    ├── pool/
    │   ├── backend.go      # Backend model
    │   └── pool.go         # Server pool management
    └── proxy/
        └── handler.go      # Main proxy handler
```

## Quick Start

### 1. Start the Backend Servers

Open two separate terminals:

**Terminal 1:**
```bash
start-backend1.bat
```

**Terminal 2:**
```bash
start-backend2.bat
```

### 2. Start the Proxy

Open a third terminal:
```bash
start-proxy.bat
```

The proxy will start on:
- Proxy Server: `http://localhost:8080`
- Admin API: `http://localhost:8081`

### 3. Test the Proxy

Run the test script:
```bash
test-proxy.bat
```

Or test manually:
```bash
# Send requests through the proxy
curl http://localhost:8080/

# Check system status
curl http://localhost:8081/status
```

## Configuration

Edit `config.json` to configure the proxy:

```json
{
  "strategy": "round-robin",
  "backends": [
    "http://localhost:9001",
    "http://localhost:9002"
  ]
}
```

**Available Strategies:**
- `round-robin`: Distributes requests evenly across backends
- `least-conn`: Routes to backend with fewest active connections

## Admin API

### Get Status
```bash
GET http://localhost:8081/status
```

Response:
```json
{
  "total_backends": 2,
  "active_backends": 2,
  "backends": [
    {
      "url": "http://localhost:9001",
      "alive": true,
      "current_connections": 5
    },
    {
      "url": "http://localhost:9002",
      "alive": true,
      "current_connections": 3
    }
  ]
}
```

### Add Backend
```bash
POST http://localhost:8081/Backends
Content-Type: application/json

{
  "url": "http://localhost:9003"
}
```

### Remove Backend
```bash
DELETE http://localhost:8081/Backends?url=http://localhost:9003
```

## How It Works

### 1. Request Flow

1. Client sends request to proxy (`:8080`)
2. Proxy selects healthy backend using load balancing strategy
3. Request is forwarded to selected backend
4. Response is returned to client
5. Connection counter is decremented

### 2. Health Monitoring

- Background goroutine checks all backends every 30 seconds
- Sends GET request to each backend
- Marks backend as DOWN if response status >= 500 or no response
- Logs status changes

### 3. Load Balancing

**Round-Robin:**
- Cycles through backends in order
- Skips backends marked as DOWN
- Uses mutex for thread-safe index tracking

**Least-Connections:**
- Selects backend with fewest active connections
- Only considers backends marked as UP
- Thread-safe connection counting

## Key Features

### Thread Safety
- `sync.RWMutex` protects concurrent access to backend pool
- `sync.Mutex` protects connection counters
- Atomic operations for reliable counting

### Context Management
- Request timeout: 5 seconds (configurable)
- Client cancellation propagation
- Graceful shutdown on SIGTERM/SIGINT

### Error Handling
- Connection refused → Mark backend DOWN immediately
- Timeout → Return 504 Gateway Timeout
- No healthy backends → Return 503 Service Unavailable
- General errors → Return 502 Bad Gateway

## Building and Running

### Build
```bash
go build -o proxy.exe
```

### Run
```bash
./proxy.exe
```

## Testing Scenarios

### Test Load Balancing
```bash
# Send multiple requests and observe alternating backends
for i in {1..6}; do curl http://localhost:8080/; done
```

### Test Health Monitoring
1. Stop one backend server (Ctrl+C)
2. Wait 30 seconds for health check
3. Observe logs showing backend marked DOWN
4. Verify requests only go to healthy backend

### Test Dynamic Backend Management
```bash
# Add backend
curl -X POST http://localhost:8081/Backends \
  -H "Content-Type: application/json" \
  -d '{"url": "http://localhost:9003"}'

# Verify it was added
curl http://localhost:8081/status

# Remove backend
curl -X DELETE "http://localhost:8081/Backends?url=http://localhost:9003"
```

## Requirements Met

✅ Complex interfaces and structural composition  
✅ net/http package usage  
✅ Shared state management with sync primitives  
✅ Background goroutines with tickers  
✅ Context for timeouts and cancellation  
✅ Round-Robin load balancing  
✅ Thread-safe operations with RWMutex  
✅ httputil.ReverseProxy for forwarding  
✅ Periodic health monitoring  
✅ Context propagation for request cancellation  

## Author

Ayoub - S6 Go Project
