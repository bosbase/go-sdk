# Logs API - Go SDK Documentation

## Overview

The Logs API provides endpoints for viewing and analyzing application logs. All operations require superuser authentication and allow you to query request logs, filter by various criteria, and get aggregated statistics.

**Key Features:**
- List and paginate logs
- View individual log entries
- Filter logs by status, URL, method, IP, etc.
- Sort logs by various fields
- Get hourly aggregated statistics
- Filter statistics by criteria

**Backend Endpoints:**
- `GET /api/logs` - List logs
- `GET /api/logs/{id}` - View log
- `GET /api/logs/stats` - Get statistics

**Note**: All Logs API operations require superuser authentication.

## Authentication

All Logs API operations require superuser authentication:

```go
package main

import (
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://127.0.0.1:8090")
    defer client.Close()
    
    // Authenticate as superuser
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

## List Logs

Returns a paginated list of logs with support for filtering and sorting.

### Basic Usage

```go
// Basic list
result, err := client.Logs.GetList(1, 30, "", "", nil, nil)
if err != nil {
    log.Fatal(err)
}

page, _ := result["page"].(float64)
perPage, _ := result["perPage"].(float64)
totalItems, _ := result["totalItems"].(float64)
items, _ := result["items"].([]interface{})

fmt.Printf("Page: %.0f, PerPage: %.0f, Total: %.0f\n", page, perPage, totalItems)
fmt.Printf("Items: %d\n", len(items))
```

### Filtering Logs

```go
// Filter by HTTP status code
errorLogs, err := client.Logs.GetList(1, 50, "data.status >= 400", "", nil, nil)

// Filter by method
getLogs, err := client.Logs.GetList(1, 50, `data.method = "GET"`, "", nil, nil)

// Filter by URL pattern
apiLogs, err := client.Logs.GetList(1, 50, `data.url ~ "/api/"`, "", nil, nil)

// Filter by execution time (slow requests)
slowLogs, err := client.Logs.GetList(1, 50, "data.execTime > 1.0", "", nil, nil)

// Complex filter
complexFilter, err := client.Logs.GetList(1, 50, 
    `data.status >= 400 && data.method = "POST" && data.execTime > 0.5`, "", nil, nil)
```

### Sorting Logs

```go
// Sort by creation date (newest first)
recent, err := client.Logs.GetList(1, 50, "", "-created", nil, nil)

// Sort by execution time (slowest first)
slowest, err := client.Logs.GetList(1, 50, "", "-data.execTime", nil, nil)
```

## View Log

Retrieve a single log entry by ID:

```go
// Get specific log
log, err := client.Logs.GetOne("ai5z3aoed6809au", nil, nil)
if err != nil {
    log.Fatal(err)
}

message, _ := log["message"].(string)
data, _ := log["data"].(map[string]interface{})
status, _ := data["status"].(float64)
execTime, _ := data["execTime"].(float64)

fmt.Printf("Message: %s\n", message)
fmt.Printf("Status: %.0f\n", status)
fmt.Printf("Execution Time: %f\n", execTime)
```

## Logs Statistics

Get hourly aggregated statistics for logs:

### Basic Usage

```go
// Get all statistics
stats, err := client.Logs.GetStats(nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, stat := range stats {
    total, _ := stat["total"].(float64)
    date, _ := stat["date"].(string)
    fmt.Printf("Date: %s, Total: %.0f\n", date, total)
}
```

### Filtered Statistics

```go
// Statistics for errors only
query := map[string]interface{}{
    "filter": "data.status >= 400",
}
errorStats, err := client.Logs.GetStats(query, nil)
```

## Complete Examples

### Example 1: Error Monitoring

```go
func getErrorMetrics(client *bosbase.BosBase) error {
    // Get error logs
    clientErrors, err := client.Logs.GetList(1, 100, 
        `data.status >= 400 && data.status < 500`, "-created", nil, nil)
    if err != nil {
        return err
    }
    
    serverErrors, err := client.Logs.GetList(1, 100, 
        "data.status >= 500", "-created", nil, nil)
    if err != nil {
        return err
    }
    
    // Get hourly statistics
    query := map[string]interface{}{
        "filter": "data.status >= 400",
    }
    errorStats, err := client.Logs.GetStats(query, nil)
    if err != nil {
        return err
    }
    
    clientItems, _ := clientErrors["items"].([]interface{})
    serverItems, _ := serverErrors["items"].([]interface{})
    
    fmt.Printf("Client errors: %d\n", len(clientItems))
    fmt.Printf("Server errors: %d\n", len(serverItems))
    fmt.Printf("Error stats entries: %d\n", len(errorStats))
    
    return nil
}
```

## Best Practices

1. **Use Filters**: Always use filters to narrow down results, especially for large log datasets
2. **Paginate**: Use pagination instead of fetching all logs at once
3. **Efficient Sorting**: Use `-rowid` for default sorting (most efficient)
4. **Filter Statistics**: Always filter statistics for meaningful insights
5. **Monitor Errors**: Regularly check for 4xx/5xx errors

## Related Documentation

- [Authentication](./AUTHENTICATION.md) - User authentication
- [Crons API](./CRONS_API.md) - Cron job management

