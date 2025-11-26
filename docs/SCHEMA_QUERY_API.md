# Schema Query API - Go SDK Documentation

## Overview

The Schema Query API provides lightweight interfaces to retrieve collection field information without fetching full collection definitions. This is particularly useful for AI systems that need to understand the structure of collections and the overall system architecture.

**Key Features:**
- Get schema for a single collection by name or ID
- Get schemas for all collections in the system
- Lightweight response with only essential field information
- Support for all collection types (base, auth, view)
- Fast and efficient queries

**Backend Endpoints:**
- `GET /api/collections/{collection}/schema` - Get single collection schema
- `GET /api/collections/schemas` - Get all collection schemas

**Note**: All Schema Query API operations require superuser authentication.

## Authentication

All Schema Query API operations require superuser authentication:

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

## Get Single Collection Schema

Retrieves the schema (fields and types) for a single collection by name or ID.

### Basic Usage

```go
// Get schema for a collection by name
schema, err := client.Collections.GetSchema("demo1", nil)
if err != nil {
    log.Fatal(err)
}

name, _ := schema["name"].(string)
schemaType, _ := schema["type"].(string)
fields, _ := schema["fields"].([]interface{})

fmt.Printf("Collection: %s (type: %s)\n", name, schemaType)
fmt.Printf("Fields: %d\n", len(fields))

// Iterate through fields
for _, field := range fields {
    f, _ := field.(map[string]interface{})
    fieldName, _ := f["name"].(string)
    fieldType, _ := f["type"].(string)
    required, _ := f["required"].(bool)
    
    requiredStr := ""
    if required {
        requiredStr = " (required)"
    }
    fmt.Printf("  - %s: %s%s\n", fieldName, fieldType, requiredStr)
}
```

### Using Collection ID

```go
// Get schema for a collection by ID
schema, err := client.Collections.GetSchema("_pbc_base_123", nil)
```

## Get All Collection Schemas

Retrieves schemas for all collections in the system.

### Basic Usage

```go
// Get all collection schemas
allSchemas, err := client.Collections.GetAllSchemas(nil)
if err != nil {
    log.Fatal(err)
}

collections, _ := allSchemas["collections"].([]interface{})

for _, collection := range collections {
    c, _ := collection.(map[string]interface{})
    name, _ := c["name"].(string)
    schemaType, _ := c["type"].(string)
    fields, _ := c["fields"].([]interface{})
    
    fmt.Printf("%s (%s): %d fields\n", name, schemaType, len(fields))
}
```

## Use Cases

### AI System Understanding

```go
func getSystemSchema(client *bosbase.BosBase) (map[string]interface{}, error) {
    allSchemas, err := client.Collections.GetAllSchemas(nil)
    if err != nil {
        return nil, err
    }
    
    // Process schemas for AI understanding
    collections, _ := allSchemas["collections"].([]interface{})
    schemaMap := make(map[string]interface{})
    
    for _, collection := range collections {
        c, _ := collection.(map[string]interface{})
        name, _ := c["name"].(string)
        schemaMap[name] = c
    }
    
    return schemaMap, nil
}
```

## Related Documentation

- [Collections](./COLLECTIONS.md) - Collection and field configuration
- [Collection API](./COLLECTION_API.md) - Collection management

