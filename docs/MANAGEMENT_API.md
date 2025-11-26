# Management API Documentation - Go SDK Documentation

This document covers the management API capabilities available in the Go SDK, which correspond to the features available in the backend management UI.

> **Note**: All management API operations require superuser authentication (🔐).

## Table of Contents

- [Settings Service](#settings-service)
- [Backup Service](#backup-service)
- [Log Service](#log-service)
- [Cron Service](#cron-service)
- [Health Service](#health-service)
- [Collection Service](#collection-service)

## Settings Service

The Settings Service provides comprehensive management of application settings, matching the capabilities available in the backend management UI.

### Get Application Settings

```go
settings, err := client.Settings.GetApplicationSettings(nil, nil)
if err != nil {
    log.Fatal(err)
}

meta, _ := settings["meta"].(map[string]interface{})
appName, _ := meta["appName"].(string)
fmt.Printf("App name: %s\n", appName)
```

### Update Application Settings

```go
_, err := client.Settings.UpdateApplicationSettings(map[string]interface{}{
    "meta": map[string]interface{}{
        "appName":    "My App",
        "appURL":     "https://example.com",
        "hideControls": false,
    },
    "trustedProxy": map[string]interface{}{
        "headers":      []string{"X-Forwarded-For"},
        "useLeftmostIP": true,
    },
}, nil, nil)
```

### Update Meta Settings

```go
_, err := client.Settings.UpdateMeta(map[string]interface{}{
    "appName":       "My App",
    "appURL":        "https://example.com",
    "senderName":    "My App",
    "senderAddress": "noreply@example.com",
    "hideControls":  false,
}, nil, nil)
```

### Update Trusted Proxy

```go
_, err := client.Settings.UpdateTrustedProxy(map[string]interface{}{
    "headers":       []string{"X-Forwarded-For", "X-Real-IP"},
    "useLeftmostIP": true,
}, nil, nil)
```

### Update Rate Limits

```go
_, err := client.Settings.UpdateRateLimits(map[string]interface{}{
    "enabled": true,
    "rules": []map[string]interface{}{
        {
            "label":       "api/users",
            "duration":    3600,
            "maxRequests": 100,
        },
    },
}, nil, nil)
```

## Backup Service

See [Backups API](./BACKUPS_API.md) for detailed documentation.

## Log Service

See [Logs API](./LOGS_API.md) for detailed documentation.

## Cron Service

See [Crons API](./CRONS_API.md) for detailed documentation.

## Health Service

See [Health API](./HEALTH_API.md) for detailed documentation.

## Collection Service

See [Collection API](./COLLECTION_API.md) for detailed documentation.

## Related Documentation

- [Backups API](./BACKUPS_API.md) - Backup management
- [Logs API](./LOGS_API.md) - Log viewing
- [Crons API](./CRONS_API.md) - Cron job management
- [Health API](./HEALTH_API.md) - Health checks
- [Collection API](./COLLECTION_API.md) - Collection management

