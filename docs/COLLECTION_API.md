# Collection API - Go SDK Documentation

## Overview

The Collection API provides endpoints for managing collections (Base, Auth, and View types). All operations require superuser authentication and allow you to create, read, update, and delete collections along with their schemas and configurations.

**Key Features:**
- List and search collections
- View collection details
- Create collections (base, auth, view)
- Update collection schemas and rules
- Delete collections
- Truncate collections (delete all records)
- Import collections in bulk
- Get collection scaffolds (templates)

**Backend Endpoints:**
- `GET /api/collections` - List collections
- `GET /api/collections/{collection}` - View collection
- `POST /api/collections` - Create collection
- `PATCH /api/collections/{collection}` - Update collection
- `DELETE /api/collections/{collection}` - Delete collection
- `DELETE /api/collections/{collection}/truncate` - Truncate collection
- `PUT /api/collections/import` - Import collections
- `GET /api/collections/meta/scaffolds` - Get scaffolds

**Note**: All Collection API operations require superuser authentication.

## Authentication

All Collection API operations require superuser authentication:

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

## List Collections

Returns a paginated list of collections with support for filtering and sorting.

```go
// Basic list
result, err := client.Collections.GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 30,
})
if err != nil {
    log.Fatal(err)
}

page, _ := result["page"].(float64)
perPage, _ := result["perPage"].(float64)
totalItems, _ := result["totalItems"].(float64)
items, _ := result["items"].([]interface{})

fmt.Printf("Page: %.0f, PerPage: %.0f, Total: %.0f\n", page, perPage, totalItems)
```

### Advanced Filtering and Sorting

```go
// Filter by type
authCollections, err := client.Collections.GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 100,
    Filter:  `type = "auth"`,
})

// Filter by name pattern
matchingCollections, err := client.Collections.GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 100,
    Filter:  `name ~ "user"`,
})

// Sort by creation date
sortedCollections, err := client.Collections.GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 100,
    Sort:    "-created",
})
```

### Get Full List

```go
// Get all collections at once
allCollections, err := client.Collections.GetFullList(500, &bosbase.CrudListOptions{
    Sort:   "name",
    Filter: "system = false",
})
```

## View Collection

Retrieve a single collection by ID or name:

```go
// By name
collection, err := client.Collections.GetOne("posts", nil)
if err != nil {
    log.Fatal(err)
}

// By ID
collection, err = client.Collections.GetOne("_pbc_2287844090", nil)
```

## Create Collection

Create a new collection with schema fields and configuration.

### Create Base Collection

```go
baseCollection, err := client.Collections.Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "name": "posts",
        "type": "base",
        "fields": []map[string]interface{}{
            {
                "name":     "title",
                "type":     "text",
                "required": true,
                "min":      10,
                "max":      255,
            },
            {
                "name":     "content",
                "type":     "editor",
                "required": false,
            },
            {
                "name":     "published",
                "type":     "bool",
                "required": false,
            },
        },
        "listRule":  `@request.auth.id != ""`,
        "viewRule":  `@request.auth.id != "" || published = true`,
        "createRule": `@request.auth.id != ""`,
        "updateRule": `author = @request.auth.id`,
        "deleteRule": `author = @request.auth.id`,
    },
})
```

### Create from Scaffold

Use predefined scaffolds as a starting point:

```go
// Create base collection from scaffold
baseCollection, err := client.Collections.CreateBase("my_posts", map[string]interface{}{
    "fields": []map[string]interface{}{
        {
            "name":     "title",
            "type":     "text",
            "required": true,
        },
    },
}, nil, nil, nil)

// Create auth collection from scaffold
authCollection, err := client.Collections.CreateAuth("my_users", map[string]interface{}{
    "passwordAuth": map[string]interface{}{
        "enabled":       true,
        "identityFields": []string{"email"},
    },
}, nil, nil, nil)

// Create view collection from scaffold
viewCollection, err := client.Collections.CreateView("my_view", 
    "SELECT id, title FROM posts", map[string]interface{}{
        "listRule": `@request.auth.id != ""`,
    }, nil, nil, nil)
```

## Update Collection

Update an existing collection's schema, fields, or rules:

```go
// Update collection name and rules
_, err := client.Collections.Update("posts", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "name":     "articles",
        "listRule": `@request.auth.id != "" || status = "public"`,
        "viewRule": `@request.auth.id != "" || status = "public"`,
    },
})

// Add new field
collection, err := client.Collections.GetOne("posts", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})
newField := map[string]interface{}{
    "name": "tags",
    "type": "select",
    "options": map[string]interface{}{
        "values": []string{"tech", "science", "art"},
    },
}
fields = append(fields, newField)

_, err = client.Collections.Update("posts", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

## Delete Collection

Delete a collection and all its records:

```go
err := client.Collections.DeleteCollection("posts", nil)
if err != nil {
    log.Fatal(err)
}
```

## Truncate Collection

Delete all records in a collection without deleting the collection itself:

```go
err := client.Collections.Truncate("posts", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Import Collections

Import multiple collections at once:

```go
collections := []map[string]interface{}{
    {
        "name": "posts",
        "type": "base",
        "fields": []map[string]interface{}{
            {"name": "title", "type": "text", "required": true},
        },
    },
    {
        "name": "users",
        "type": "auth",
        "fields": []map[string]interface{}{
            {"name": "name", "type": "text", "required": true},
        },
    },
}

err := client.Collections.ImportCollections(collections, false, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Get Scaffolds

Get available collection scaffolds (templates):

```go
scaffolds, err := client.Collections.GetScaffolds(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Scaffolds contain templates for base, auth, and view collections
baseScaffold, _ := scaffolds["base"].(map[string]interface{})
authScaffold, _ := scaffolds["auth"].(map[string]interface{})
viewScaffold, _ := scaffolds["view"].(map[string]interface{})
```

## Related Documentation

- [Collections](./COLLECTIONS.md) - Collection and field configuration
- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Understanding API rules

