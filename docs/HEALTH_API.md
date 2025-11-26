# Health API - Go SDK Documentation

## Overview

The Health API provides a simple endpoint to check the health status of the server. It returns basic health information and, when authenticated as a superuser, provides additional diagnostic information about the server state.

**Key Features:**
- No authentication required for basic health check
- Superuser authentication provides additional diagnostic data
- Lightweight endpoint for monitoring and health checks

**Backend Endpoints:**
- `GET /api/health` - Check health status

**Note**: The health endpoint is publicly accessible, but superuser authentication provides additional information.

## Authentication

Basic health checks do not require authentication:

```go
package main

import (
    "fmt"
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://127.0.0.1:8090")
    defer client.Close()
    
    // Basic health check (no auth required)
    health, err := client.Health.Check(nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Health:", health)
}
```

For additional diagnostic information, authenticate as a superuser:

```go
// Authenticate as superuser for extended health data
_, err := client.Collection("_superusers").AuthWithPassword(
    "admin@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

health, err := client.Health.Check(nil, nil)
if err != nil {
    log.Fatal(err)
}

data, _ := health["data"].(map[string]interface{})
canBackup, _ := data["canBackup"].(bool)
realIP, _ := data["realIP"].(string)
fmt.Printf("Can backup: %v, Real IP: %s\n", canBackup, realIP)
```

## Health Check Response Structure

### Basic Response (Guest/Regular User)

```go
{
    "code": 200,
    "message": "API is healthy.",
    "data": {}
}
```

### Superuser Response

```go
{
    "code": 200,
    "message": "API is healthy.",
    "data": {
        "canBackup": bool,           // Whether backup operations are allowed
        "realIP": string,            // Real IP address of the client
        "requireS3": bool,           // Whether S3 storage is required
        "possibleProxyHeader": string // Detected proxy header (if behind reverse proxy)
    }
}
```

## Check Health Status

Returns the health status of the API server.

### Basic Usage

```go
// Simple health check
health, err := client.Health.Check(nil, nil)
if err != nil {
    log.Fatal(err)
}

code, _ := health["code"].(float64)
message, _ := health["message"].(string)
fmt.Printf("Status: %.0f - %s\n", code, message)
```

### With Superuser Authentication

```go
// Authenticate as superuser first
_, err := client.Collection("_superusers").AuthWithPassword(
    "admin@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Get extended health information
health, err := client.Health.Check(nil, nil)
if err != nil {
    log.Fatal(err)
}

data, _ := health["data"].(map[string]interface{})
canBackup, _ := data["canBackup"].(bool)
realIP, _ := data["realIP"].(string)
requireS3, _ := data["requireS3"].(bool)
proxyHeader, _ := data["possibleProxyHeader"].(string)

fmt.Printf("Can backup: %v\n", canBackup)
fmt.Printf("Real IP: %s\n", realIP)
fmt.Printf("Require S3: %v\n", requireS3)
fmt.Printf("Proxy header: %s\n", proxyHeader)
```

## Use Cases

### 1. Basic Health Monitoring

```go
func checkServerHealth(client *bosbase.BosBase) bool {
    health, err := client.Health.Check(nil, nil)
    if err != nil {
        return false
    }
    
    code, _ := health["code"].(float64)
    message, _ := health["message"].(string)
    
    return code == 200 && message == "API is healthy."
}

// Use in monitoring
for {
    if !checkServerHealth(client) {
        fmt.Println("Server health check failed!")
    }
    time.Sleep(60 * time.Second)
}
```

### 2. Backup Readiness Check

```go
func canPerformBackup(client *bosbase.BosBase) (bool, error) {
    // Authenticate as superuser
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        return false, err
    }
    
    health, err := client.Health.Check(nil, nil)
    if err != nil {
        return false, err
    }
    
    data, _ := health["data"].(map[string]interface{})
    canBackup, _ := data["canBackup"].(bool)
    
    return canBackup, nil
}

// Use before creating backups
if canBackup, _ := canPerformBackup(client); canBackup {
    err := client.Backups.Create("backup.zip", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Best Practices

1. **Monitoring**: Use health checks for regular monitoring (e.g., every 30-60 seconds)
2. **Load Balancers**: Configure load balancers to use the health endpoint for health checks
3. **Pre-flight Checks**: Check `canBackup` before initiating backup operations
4. **Error Handling**: Always handle errors gracefully as the server may be down
5. **Rate Limiting**: Don't poll the health endpoint too frequently (avoid spamming)

## Related Documentation

- [Backups API](./BACKUPS_API.md) - Using `canBackup` to check backup readiness
- [Authentication](./AUTHENTICATION.md) - Superuser authentication

