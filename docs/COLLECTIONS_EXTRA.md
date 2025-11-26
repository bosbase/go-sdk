# Collections Extra - Go SDK Documentation

This document provides comprehensive documentation for working with Collections and Fields in the BosBase Go SDK. This documentation is designed to be AI-readable and includes practical examples for all operations.

## Table of Contents

- [Overview](#overview)
- [Collection Types](#collection-types)
- [Collections API](#collections-api)
- [Records API](#records-api)
- [Field Types](#field-types)
- [Examples](#examples)

## Overview

**Collections** represent your application data. Under the hood they are backed by plain SQLite tables that are generated automatically with the collection **name** and **fields** (columns).

A single entry of a collection is called a **record** (a single row in the SQL table).

You can manage your **collections** from the Dashboard, or with the Go SDK using the `Collections` service.

Similarly, you can manage your **records** from the Dashboard, or with the Go SDK using the `Collection(name)` method which returns a `RecordService` instance.

## Collection Types

Currently there are 3 collection types: **Base**, **View** and **Auth**.

### Base Collection

**Base collection** is the default collection type and it could be used to store any application data (articles, products, posts, etc.).

```go
package main

import (
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    // Authenticate as admin
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a base collection
    collection, err := client.Collections.CreateBase("articles", map[string]interface{}{
        "fields": []map[string]interface{}{
            {
                "name":     "title",
                "type":     "text",
                "required": true,
                "min":      6,
                "max":      100,
            },
            {
                "name": "description",
                "type": "text",
            },
        },
        "listRule": `@request.auth.id != "" || status = "public"`,
        "viewRule": `@request.auth.id != "" || status = "public"`,
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Created collection:", collection)
}
```

### View Collection

**View collection** is a read-only collection type where the data is populated from a plain SQL `SELECT` statement, allowing users to perform aggregations or any other custom queries.

```go
// Create a view collection
viewCollection, err := client.Collections.CreateView("post_stats", 
    `SELECT posts.id, posts.name, count(comments.id) as totalComments 
     FROM posts 
     LEFT JOIN comments on comments.postId = posts.id 
     GROUP BY posts.id`,
    nil, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

**Note**: View collections don't receive realtime events because they don't have create/update/delete operations.

### Auth Collection

**Auth collection** has everything from the **Base collection** but with some additional special fields to help you manage your app users and also provide various authentication options.

Each Auth collection has the following special system fields: `email`, `emailVisibility`, `verified`, `password` and `tokenKey`. They cannot be renamed or deleted but can be configured using their specific field options.

```go
// Create an auth collection
usersCollection, err := client.Collections.CreateAuth("users", map[string]interface{}{
    "fields": []map[string]interface{}{
        {
            "name":     "name",
            "type":     "text",
            "required": true,
        },
        {
            "name": "role",
            "type": "select",
            "options": map[string]interface{}{
                "values": []string{"employee", "staff", "admin"},
            },
        },
    },
}, nil, nil, nil)
```

You can have as many Auth collections as you want (users, managers, staffs, members, clients, etc.) each with their own set of fields, separate login and records managing endpoints.

## Collections API

### Initialize Client

```go
client := bosbase.New("http://localhost:8090")
defer client.Close()

// Authenticate as superuser (required for collection management)
_, err := client.Collection("_superusers").AuthWithPassword(
    "admin@example.com", "password", "", "", nil, nil, nil)
```

### List Collections

```go
// Get paginated list
result, err := client.Collections.GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 50,
})

// Get all collections
allCollections, err := client.Collections.GetFullList(500, nil)
```

### Get Collection

```go
// By ID or name
collection, err := client.Collections.GetOne("articles", nil)
// or
collection, err = client.Collections.GetOne("COLLECTION_ID", nil)
```

## Records API

### List Records

```go
result, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  "published = true",
    Sort:    "-created",
    Expand:  "author",
})
```

### Get Record

```go
record, err := client.Collection("articles").GetOne("RECORD_ID", &bosbase.CrudViewOptions{
    Expand: "author,category",
})
```

### Create Record

```go
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Article",
        "views+": 1, // Field modifier
    },
})
```

### Update Record

```go
_, err := client.Collection("articles").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":  "Updated",
        "views+": 1,
        "tags+":  "new-tag",
    },
})
```

### Delete Record

```go
err := client.Collection("articles").Delete("RECORD_ID", nil)
```

## Field Types

All field types are documented with examples in the [Collections](./COLLECTIONS.md) documentation.

## Related Documentation

- [Collections](./COLLECTIONS.md) - Complete collection and field documentation
- [API Records](./API_RECORDS.md) - CRUD operations
- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Access control

