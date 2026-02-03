# Testing Guide for Reverse Proxy

## Quick Start (Recommended)

### Step 1: Start All Servers
Double-click `start-all.bat` - this will open 3 windows:
- Backend Server 1 (port 9001)
- Backend Server 2 (port 9002)
- Reverse Proxy (ports 8080 & 8081)

Wait 5 seconds for all servers to fully start.

### Step 2: Run Tests
Double-click `test-proxy.bat` to run automated tests.

## Manual Testing Steps

### 1. Test Basic Load Balancing

Open PowerShell and run:
```powershell
# Request 1 - Should go to Backend 1
curl http://localhost:8080/

# Request 2 - Should go to Backend 2
curl http://localhost:8080/

# Request 3 - Should go to Backend 1 again
curl http://localhost:8080/
```

You should see responses alternating between "Backend 1 (port 9001)" and "Backend 2 (port 9002)".

### 2. Test Status Endpoint

```powershell
curl http://localhost:8081/status
```

Expected output:
```json
{
  "total_backends": 2,
  "active_backends": 2,
  "backends": [
    {
      "url": "http://localhost:9001",
      "alive": true,
      "current_connections": 0
    },
    {
      "url": "http://localhost:9002",
      "alive": true,
      "current_connections": 0
    }
  ]
}
```

### 3. Test Adding a Backend

```powershell
curl -Method POST -Uri http://localhost:8081/Backends -Headers @{"Content-Type"="application/json"} -Body '{"url": "http://localhost:9003"}'
```

Check status again:
```powershell
curl http://localhost:8081/status
```

You should now see 3 backends (though the 3rd will be marked as DOWN since it's not running).

### 4. Test Removing a Backend

```powershell
curl -Method DELETE -Uri "http://localhost:8081/Backends?url=http://localhost:9003"
```

Verify removal:
```powershell
curl http://localhost:8081/status
```

### 5. Test Health Monitoring

#### Simulate Backend Failure:
1. Go to the Backend 1 window (port 9001)
2. Press Ctrl+C to stop it
3. Wait 30 seconds for health check to run
4. Check status:
```powershell
curl http://localhost:8081/status
```

You should see Backend 1 marked as `"alive": false`.

#### Send requests:
```powershell
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/
```

All requests should now go to Backend 2 only (since Backend 1 is DOWN).

#### Bring Backend 1 back:
1. Restart Backend 1: `start-backend1.bat`
2. Wait 30 seconds for health check
3. Check status - Backend 1 should be UP again
4. Send requests - load balancing should resume

### 6. Test Least-Connections Strategy

1. Stop the proxy (Ctrl+C in proxy window)
2. Edit `config.json`:
```json
{
  "strategy": "least-conn",
  "backends": [
    "http://localhost:9001",
    "http://localhost:9002"
  ]
}
```
3. Restart proxy: `start-proxy.bat`
4. Send multiple requests:
```powershell
curl http://localhost:8080/
curl http://localhost:8080/
curl http://localhost:8080/
```

With least-connections, requests go to the backend with the fewest active connections.

### 7. Test Error Handling

#### Test with no backends:
1. Stop both backend servers
2. Wait 30 seconds for health check
3. Send request:
```powershell
curl http://localhost:8080/
```

Expected: "503 Service Unavailable"

#### Test connection refused:
1. Remove all backends from config and restart proxy
2. Try adding a non-existent backend:
```powershell
curl -Method POST -Uri http://localhost:8081/Backends -Headers @{"Content-Type"="application/json"} -Body '{"url": "http://localhost:9999"}'
```
3. Send request through proxy:
```powershell
curl http://localhost:8080/
```

The proxy should detect the connection failure and mark the backend as DOWN.

## Expected Behaviors

### Round-Robin Strategy
- Requests distributed evenly across all healthy backends
- Cycles through backends in order: 1 → 2 → 1 → 2 → ...
- Skips backends marked as DOWN

### Least-Connections Strategy
- Routes to backend with fewest active connections
- Better for long-running requests
- Automatically balances load

### Health Monitoring
- Checks every 30 seconds
- Marks backend DOWN if:
  - Connection refused
  - HTTP status >= 500
  - Request timeout (2 seconds)
- Marks backend UP when it responds with status < 500

### Admin API
- Add backends dynamically without restart
- Remove backends dynamically
- View real-time status of all backends
- See current connection counts

## Troubleshooting

### Ports Already in Use
If you get "address already in use" errors:
```powershell
# Check what's using the ports
netstat -ano | findstr :8080
netstat -ano | findstr :8081
netstat -ano | findstr :9001
netstat -ano | findstr :9002

# Kill processes if needed
taskkill /PID <process_id> /F
```

### Backends Not Starting
- Make sure Go is installed: `go version`
- Check for firewall blocking ports
- Try different ports in config.json

### Health Check Not Working
- Wait at least 30 seconds between checks
- Check proxy logs for health check messages
- Verify backend servers are responding

### Load Balancing Not Working
- Ensure both backends are UP: `curl http://localhost:8081/status`
- Check logs in each window
- Verify strategy in config.json is correct

## Demo for Professor

### Quick Demo Script (2 minutes):

1. **Start everything:** Double-click `start-all.bat`

2. **Show load balancing:**
   ```powershell
   curl http://localhost:8080/
   curl http://localhost:8080/
   curl http://localhost:8080/
   ```
   Point out responses alternate between backends.

3. **Show status:**
   ```powershell
   curl http://localhost:8081/status
   ```
   Show JSON with backend health and connection counts.

4. **Show health monitoring:**
   - Stop Backend 1 (Ctrl+C)
   - Wait 30 seconds
   - Show logs: "BACKEND DOWN: http://localhost:9001"
   - Check status: Backend 1 is now `"alive": false`
   - Send requests: All go to Backend 2

5. **Show dynamic management:**
   ```powershell
   curl -Method POST -Uri http://localhost:8081/Backends -Headers @{"Content-Type"="application/json"} -Body '{"url": "http://localhost:9003"}'
   curl http://localhost:8081/status
   ```
   Show new backend added.

This demonstrates all key features: load balancing, health monitoring, admin API, and graceful handling of failures.

## Success Criteria Checklist

✅ Proxy forwards requests to backends  
✅ Load balancing works (Round-Robin)  
✅ Least-Connections strategy works  
✅ Health checks run every 30 seconds  
✅ Backends auto-marked DOWN when unreachable  
✅ Backends auto-marked UP when recovered  
✅ Admin API returns status  
✅ Admin API adds backends  
✅ Admin API removes backends  
✅ Thread-safe operations (no race conditions)  
✅ Context propagation (timeouts work)  
✅ Graceful shutdown (Ctrl+C)  
✅ Connection counting accurate  
✅ Error handling for all scenarios  

All requirements from the assignment are implemented!
