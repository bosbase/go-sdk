# Register Existing SQL Tables with the Go SDK

Use the SQL table helpers to expose existing tables (or run SQL to create them) and automatically generate REST collections. Both calls are **superuser-only**.

- `RegisterSqlTables(tables []string, ...)` – map existing tables to collections without running SQL.
- `ImportSqlTables(tables []SqlTableDefinition, ...)` – optionally run SQL to create tables first, then register them. Returns `{ created, skipped }`.

## Requirements

- Authenticate with a `_superusers` token.
- Each table must contain a `TEXT` primary key column named `id`.
- Missing audit columns (`created`, `updated`, `createdBy`, `updatedBy`) are automatically added so the default API rules can be applied.
- Non-system columns are mapped by best effort (text, number, bool, date/time, JSON).

## Basic Usage

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
    
    // Must be a superuser token
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    collections, err := client.Collections.RegisterSqlTables(
        []string{"projects", "accounts"},
        nil, nil, nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    for _, collection := range collections {
        name, _ := collection["name"].(string)
        fmt.Printf("Registered: %s\n", name)
    }
}
```

## With Request Options

You can pass standard request options (headers, query params, etc.).

```go
collections, err := client.Collections.RegisterSqlTables(
    []string{"legacy_orders"},
    map[string]interface{}{
        "q": 1, // adds ?q=1
    },
    map[string]string{
        "x-trace-id": "reg-123",
    },
)
```

## Create-or-register flow

`ImportSqlTables()` accepts `SqlTableDefinition { name: string; sql?: string }` items, runs the SQL (if provided), and registers collections. Existing collection names are reported under `skipped`.

```go
tables := []bosbase.SqlTableDefinition{
    {
        Name: "legacy_orders",
        SQL: `
            CREATE TABLE IF NOT EXISTS legacy_orders (
                id TEXT PRIMARY KEY,
                customer_email TEXT NOT NULL
            );
        `,
    },
    {
        Name: "reporting_view", // assumes table already exists
    },
}

result, err := client.Collections.ImportSqlTables(tables, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Process created collections
created, _ := result["created"].([]interface{})
for _, c := range created {
    collection, _ := c.(map[string]interface{})
    name, _ := collection["name"].(string)
    fmt.Printf("Created: %s\n", name)
}

// Process skipped collections
skipped, _ := result["skipped"].([]interface{})
for _, s := range skipped {
    name, _ := s.(string)
    fmt.Printf("Skipped: %s\n", name)
}
```

## What It Does

- Creates BosBase collection metadata for the provided tables.
- Generates REST endpoints for CRUD against those tables.
- Applies the standard default API rules (authenticated create; update/delete scoped to the creator).
- Ensures audit columns exist (`created`, `updated`, `createdBy`, `updatedBy`) and leaves all other existing SQL schema and data untouched; no further field mutations or table syncs are performed.
- Marks created collections with `externalTable: true` so you can distinguish them from regular BosBase-managed tables.

## Troubleshooting

- 400 error: ensure `id` exists as `TEXT PRIMARY KEY` and the table name is not system-reserved (no leading `_`).
- 401/403: confirm you are authenticated as a superuser.
- Default audit fields (`created`, `updated`, `createdBy`, `updatedBy`) are auto-added if they're missing so the default owner rules validate successfully.

## Related Documentation

- [Collection API](./COLLECTION_API.md) - Collection management
- [Collections](./COLLECTIONS.md) - Collection configuration

