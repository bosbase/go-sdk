# AI Development Guide - Go SDK Documentation

This guide provides a comprehensive, fast reference for AI systems to quickly develop applications using the BosBase Go SDK. All examples are production-ready and follow best practices.

## Table of Contents

1. [Authentication](#authentication)
2. [Initialize Collections](#initialize-collections)
3. [Define Collection Fields](#define-collection-fields)
4. [Add Data to Collections](#add-data-to-collections)
5. [Modify Collection Data](#modify-collection-data)
6. [Delete Data from Collections](#delete-data-from-collections)
7. [Query Collection Contents](#query-collection-contents)
8. [Add and Delete Fields from Collections](#add-and-delete-fields-from-collections)
9. [Query Collection Field Information](#query-collection-field-information)
10. [Upload Files](#upload-files)
11. [Query Logs](#query-logs)

## Authentication

### Initialize Client

```go
package main

import (
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
}
```

### Password Authentication

```go
// Authenticate with email/username and password
authData, err := client.Collection("users").AuthWithPassword(
    "user@example.com",
    "password123",
    "", "", nil, nil, nil,
)
if err != nil {
    log.Fatal(err)
}

// Auth data is automatically stored
fmt.Printf("Valid: %v\n", client.AuthStore.IsValid())
fmt.Printf("Token: %s\n", client.AuthStore.Token())
record := client.AuthStore.Record()
fmt.Printf("Record: %v\n", record)
```

### Check Authentication Status

```go
if client.AuthStore.IsValid() {
    record := client.AuthStore.Record()
    email, _ := record["email"].(string)
    fmt.Printf("Authenticated as: %s\n", email)
} else {
    fmt.Println("Not authenticated")
}
```

### Logout

```go
client.AuthStore.Clear()
```

## Initialize Collections

### Create Base Collection

```go
collection, err := client.Collections.CreateBase("posts", map[string]interface{}{
    "fields": []map[string]interface{}{
        {
            "name":     "title",
            "type":     "text",
            "required": true,
        },
        {
            "name": "content",
            "type": "editor",
        },
    },
}, nil, nil, nil)
```

### Create Auth Collection

```go
collection, err := client.Collections.CreateAuth("users", map[string]interface{}{
    "fields": []map[string]interface{}{
        {
            "name":     "name",
            "type":     "text",
            "required": true,
        },
    },
}, nil, nil, nil)
```

## Define Collection Fields

See [Collections](./COLLECTIONS.md) for complete field type documentation.

## Add Data to Collections

```go
record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":   "My Post",
        "content": "Post content",
    },
})
```

## Modify Collection Data

```go
_, err := client.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Updated Title",
    },
})
```

## Delete Data from Collections

```go
err := client.Collection("posts").Delete("RECORD_ID", nil)
```

## Query Collection Contents

```go
result, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  "published = true",
    Sort:    "-created",
})
```

## Add and Delete Fields from Collections

```go
// Get collection
collection, err := client.Collections.GetOne("posts", nil)
if err != nil {
    log.Fatal(err)
}

// Get fields
fields, _ := collection["fields"].([]interface{})

// Add new field
newField := map[string]interface{}{
    "name": "views",
    "type": "number",
}
fields = append(fields, newField)

// Update collection
_, err = client.Collections.Update("posts", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

## Query Collection Field Information

```go
// Get schema for a collection
schema, err := client.Collections.GetSchema("posts", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := schema["fields"].([]interface{})
for _, field := range fields {
    f, _ := field.(map[string]interface{})
    name, _ := f["name"].(string)
    fieldType, _ := f["type"].(string)
    fmt.Printf("%s: %s\n", name, fieldType)
}
```

## Upload Files

```go
file, err := os.Open("image.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Post",
    },
    Files: map[string]bosbase.FileParam{
        "image": {
            Filename:    "image.jpg",
            Reader:      file,
            ContentType: "image/jpeg",
        },
    },
})
```

## Query Logs

```go
// Authenticate as superuser first
_, err := client.Collection("_superusers").AuthWithPassword(
    "admin@example.com", "password", "", "", nil, nil, nil)

// Query logs
result, err := client.Logs.GetList(1, 50, "data.status >= 400", "-created", nil, nil)
if err != nil {
    log.Fatal(err)
}

items, _ := result["items"].([]interface{})
for _, item := range items {
    log, _ := item.(map[string]interface{})
    message, _ := log["message"].(string)
    fmt.Println(message)
}
```

## Related Documentation

- [Collections](./COLLECTIONS.md) - Complete collection documentation
- [API Records](./API_RECORDS.md) - CRUD operations
- [Authentication](./AUTHENTICATION.md) - Authentication methods

