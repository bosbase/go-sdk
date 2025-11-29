# SQL API - Go SDK

Execute raw SQL against the BosBase database. This endpoint mirrors the management UI SQL console and is **superuser-only** (`_superusers` token required).

## Execute SQL

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

    // Authenticate as superuser first
    _, err := client.Collection("_superusers").
        AuthWithPassword("admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Read data
    result, err := client.SQL.Execute("SELECT id, email FROM users LIMIT 5", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Columns:", result.Columns)
    fmt.Println("Rows:", result.Rows)

    // Mutations return rowsAffected
    updateRes, err := client.SQL.Execute("UPDATE users SET verified = 1", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Updated %d rows\n", updateRes.RowsAffected)
}
```

## Options

- `query`: attach query parameters to the request (e.g., tracing IDs)
- `headers`: supply custom headers for auditing or routing

The response includes:
- `Columns`: column names (for queries returning rows)
- `Rows`: two-dimensional slice of values stringified by the API
- `RowsAffected`: number of rows touched by write operations

See also: [SQL Table Registration](./SQL_TABLE_REGISTRATION.md) for mapping existing tables to collections.
