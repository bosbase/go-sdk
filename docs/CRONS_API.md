# Crons API - Go SDK Documentation

## Overview

The Crons API provides endpoints for viewing and manually triggering scheduled cron jobs. All operations require superuser authentication and allow you to list registered cron jobs and execute them on-demand.

**Key Features:**
- List all registered cron jobs
- View cron job schedules (cron expressions)
- Manually trigger cron jobs
- Built-in system jobs for maintenance tasks

**Backend Endpoints:**
- `GET /api/crons` - List cron jobs
- `POST /api/crons/{jobId}` - Run cron job

**Note**: All Crons API operations require superuser authentication.

## Authentication

All Crons API operations require superuser authentication:

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

## List Cron Jobs

Returns a list of all registered cron jobs with their IDs and schedule expressions.

### Basic Usage

```go
// Get all cron jobs
jobs, err := client.Crons.GetFullList(nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, job := range jobs {
    id, _ := job["id"].(string)
    expression, _ := job["expression"].(string)
    fmt.Printf("Job: %s, Schedule: %s\n", id, expression)
}
```

### Built-in System Jobs

The following cron jobs are typically registered by default:

| Job ID | Expression | Description | Schedule |
|--------|-----------|-------------|----------|
| `__pbLogsCleanup__` | `0 */6 * * *` | Cleans up old log entries | Every 6 hours |
| `__pbDBOptimize__` | `0 0 * * *` | Optimizes database | Daily at midnight |
| `__pbMFACleanup__` | `0 * * * *` | Cleans up expired MFA records | Every hour |
| `__pbOTPCleanup__` | `0 * * * *` | Cleans up expired OTP codes | Every hour |

## Run Cron Job

Manually trigger a cron job to execute immediately.

### Basic Usage

```go
// Run a specific cron job
err := client.Crons.Run("__pbLogsCleanup__", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### Use Cases

```go
// Trigger logs cleanup manually
func cleanupLogsNow(client *bosbase.BosBase) error {
    err := client.Crons.Run("__pbLogsCleanup__", nil, nil, nil)
    if err != nil {
        return err
    }
    fmt.Println("Logs cleanup triggered")
    return nil
}

// Trigger database optimization
func optimizeDatabase(client *bosbase.BosBase) error {
    err := client.Crons.Run("__pbDBOptimize__", nil, nil, nil)
    if err != nil {
        return err
    }
    fmt.Println("Database optimization triggered")
    return nil
}
```

## Cron Expression Format

Cron expressions use the standard 5-field format:

```
* * * * *
│ │ │ │ │
│ │ │ │ └─── Day of week (0-7, 0 or 7 is Sunday)
│ │ │ └───── Month (1-12)
│ │ └─────── Day of month (1-31)
│ └───────── Hour (0-23)
└─────────── Minute (0-59)
```

### Common Patterns

| Expression | Description |
|------------|-------------|
| `0 * * * *` | Every hour at minute 0 |
| `0 */6 * * *` | Every 6 hours |
| `0 0 * * *` | Daily at midnight |
| `0 0 * * 0` | Weekly on Sunday at midnight |
| `0 0 1 * *` | Monthly on the 1st at midnight |
| `*/30 * * * *` | Every 30 minutes |
| `0 9 * * 1-5` | Weekdays at 9 AM |

## Complete Examples

### Example 1: Cron Job Monitor

```go
type CronMonitor struct {
    client *bosbase.BosBase
}

func (cm *CronMonitor) ListAllJobs() ([]map[string]interface{}, error) {
    jobs, err := cm.client.Crons.GetFullList(nil, nil)
    if err != nil {
        return nil, err
    }
    
    fmt.Printf("Found %d cron jobs:\n", len(jobs))
    for _, job := range jobs {
        id, _ := job["id"].(string)
        expression, _ := job["expression"].(string)
        fmt.Printf("  - %s: %s\n", id, expression)
    }
    
    return jobs, nil
}

func (cm *CronMonitor) RunJob(jobID string) error {
    err := cm.client.Crons.Run(jobID, nil, nil, nil)
    if err != nil {
        return fmt.Errorf("failed to run %s: %v", jobID, err)
    }
    fmt.Printf("Successfully triggered: %s\n", jobID)
    return nil
}

// Usage
monitor := &CronMonitor{client: client}
jobs, _ := monitor.ListAllJobs()
monitor.RunJob("__pbLogsCleanup__")
```

## Best Practices

1. **Check Job Existence**: Verify a cron job exists before trying to run it
2. **Error Handling**: Always handle errors when running cron jobs
3. **Rate Limiting**: Don't trigger cron jobs too frequently manually
4. **Monitoring**: Regularly check that expected cron jobs are registered
5. **Logging**: Log when cron jobs are manually triggered for auditing

## Limitations

- **Superuser Only**: All operations require superuser authentication
- **Read-Only API**: The SDK API only allows listing and running jobs; adding/removing jobs must be done via backend hooks
- **Asynchronous Execution**: Running a cron job triggers it asynchronously; the API returns immediately
- **No Status**: The API doesn't provide execution status or history

## Related Documentation

- [Logs API](./LOGS_API.md) - Log viewing and analysis
- [Backups API](./BACKUPS_API.md) - Backup management

