# Cache API - Go SDK Documentation

## Overview

BosBase caches combine in-memory FreeCache storage with persistent database copies. Each cache instance is safe to use in single-node or multi-node (cluster) mode: nodes read from FreeCache first, fall back to the database if an item is missing or expired, and then reload FreeCache automatically.

The Go SDK exposes the cache endpoints through `client.Caches`. Typical use cases include:

- Caching AI prompts/responses that must survive restarts.
- Quickly sharing feature flags and configuration between workers.
- Preloading expensive vector search results for short periods.

> **Timeouts & TTLs:** Each cache defines a default TTL (in seconds). Individual entries may provide their own `ttlSeconds`. A value of `0` keeps the entry until it is manually deleted.

## List Available Caches

The `List()` function allows you to query and retrieve all currently available caches, including their names and capacities. This is particularly useful for AI systems to discover existing caches before creating new ones, avoiding duplicate cache creation.

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
    
    // Authenticate as superuser
    _, err := client.Collection("_superusers").AuthWithPassword(
        "root@example.com", "hunter2", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Query all available caches
    caches, err := client.Caches.List(nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Each cache object contains:
    // - name: string - The cache identifier
    // - sizeBytes: number - The cache capacity in bytes
    // - defaultTTLSeconds: number - Default expiration time
    // - readTimeoutMs: number - Read timeout in milliseconds
    // - created: string - Creation timestamp (RFC3339)
    // - updated: string - Last update timestamp (RFC3339)
    
    // Example: Find a cache by name and check its capacity
    for _, cache := range caches {
        name, _ := cache["name"].(string)
        sizeBytes, _ := cache["sizeBytes"].(float64)
        if name == "ai-session" {
            fmt.Printf("Cache \"%s\" has capacity of %.0f bytes\n", name, sizeBytes)
        }
    }
}
```

## Manage Cache Configurations

```go
// List all available caches
caches, err := client.Caches.List(nil, nil)
if err != nil {
    log.Fatal(err)
}

// Find an existing cache by name
var existingCache map[string]interface{}
for _, cache := range caches {
    name, _ := cache["name"].(string)
    if name == "ai-session" {
        existingCache = cache
        break
    }
}

if existingCache != nil {
    sizeBytes, _ := existingCache["sizeBytes"].(float64)
    fmt.Printf("Found cache \"ai-session\" with capacity %.0f bytes\n", sizeBytes)
} else {
    // Create a new cache only if it doesn't exist
    sizeBytes := 64 * 1024 * 1024
    defaultTTL := 300
    readTimeout := 25
    _, err := client.Caches.Create("ai-session", &sizeBytes, &defaultTTL, &readTimeout, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}

// Update limits later (e.g., shrink TTL to 2 minutes)
defaultTTL := 120
_, err = client.Caches.Update("ai-session", map[string]interface{}{
    "defaultTTLSeconds": defaultTTL,
}, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Delete the cache (DB rows + FreeCache)
err = client.Caches.Delete("ai-session", nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Work with Cache Entries

```go
// Store an object in cache. The same payload is serialized into the DB.
ttlSeconds := 90
_, err := client.Caches.SetEntry("ai-session", "dialog:42", map[string]interface{}{
    "prompt":    "describe Saturn",
    "embedding": []float64{0.1, 0.2, 0.3}, // vector
}, &ttlSeconds, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Read from cache. `source` indicates where the hit came from.
entry, err := client.Caches.GetEntry("ai-session", "dialog:42", nil, nil)
if err != nil {
    log.Fatal(err)
}

source, _ := entry["source"].(string)   // "cache" or "database"
expiresAt, _ := entry["expiresAt"].(string) // RFC3339 timestamp or empty
value, _ := entry["value"].(map[string]interface{})

fmt.Printf("Source: %s, Expires at: %s\n", source, expiresAt)

// Renew an entry's TTL without changing its value
renewTTL := 120
renewed, err := client.Caches.RenewEntry("ai-session", "dialog:42", &renewTTL, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

newExpiresAt, _ := renewed["expiresAt"].(string)
fmt.Printf("New expiration time: %s\n", newExpiresAt)

// Delete an entry
err = client.Caches.DeleteEntry("ai-session", "dialog:42", nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### Cluster-Aware Behavior

1. **Write-through persistence** – every `SetEntry` writes to FreeCache and the `_cache_entries` table so other nodes (or a restarted node) can immediately reload values.
2. **Read path** – FreeCache is consulted first. If a lock cannot be acquired within `readTimeoutMs` or if the entry is missing/expired, BosBase queries the database copy and repopulates FreeCache in the background.
3. **Automatic cleanup** – expired entries are ignored and removed from the database when fetched, preventing stale data across nodes.

Use caches whenever you need fast, transient data that must still be recoverable or shareable across BosBase nodes.

## Field Reference

| Field | Description |
|-------|-------------|
| `sizeBytes` | Approximate FreeCache size. Values too small (<512KB) or too large (>512MB) are clamped. |
| `defaultTTLSeconds` | Default expiration for entries. `0` means no expiration. |
| `readTimeoutMs` | Optional lock timeout while reading FreeCache. When exceeded, the value is fetched from the database instead. |

## Complete Example

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
    
    // Authenticate
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Check if cache exists
    caches, _ := client.Caches.List(nil, nil)
    cacheExists := false
    for _, cache := range caches {
        name, _ := cache["name"].(string)
        if name == "my-cache" {
            cacheExists = true
            break
        }
    }
    
    // Create cache if it doesn't exist
    if !cacheExists {
        sizeBytes := 32 * 1024 * 1024 // 32MB
        defaultTTL := 600              // 10 minutes
        readTimeout := 50
        _, err := client.Caches.Create("my-cache", &sizeBytes, &defaultTTL, &readTimeout, nil, nil, nil)
        if err != nil {
            log.Fatal(err)
        }
    }
    
    // Store data
    ttl := 300 // 5 minutes
    _, err = client.Caches.SetEntry("my-cache", "key1", map[string]interface{}{
        "data": "value1",
    }, &ttl, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Retrieve data
    entry, err := client.Caches.GetEntry("my-cache", "key1", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    value, _ := entry["value"].(map[string]interface{})
    fmt.Println("Retrieved:", value)
}
```

## Related Documentation

- [Collections](./COLLECTIONS.md) - Collection and field configuration

