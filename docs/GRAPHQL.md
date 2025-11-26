# GraphQL - Go SDK Documentation

## Overview

Use `client.GraphQL.Query()` to call `/api/graphql` with your current auth token. It returns `{ data, errors, extensions }`.

> **Authentication**: the GraphQL endpoint is **superuser-only**. Authenticate as a superuser before calling GraphQL.

## Authentication

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

## Single-table Query

```go
query := `
  query ActiveUsers($limit: Int!) {
    records(collection: "users", perPage: $limit, filter: "status = true") {
      items { id data }
    }
  }
`

variables := map[string]interface{}{
    "limit": 5,
}

result, err := client.GraphQL.Query(query, variables, nil, nil)
if err != nil {
    log.Fatal(err)
}

data, _ := result["data"].(map[string]interface{})
records, _ := data["records"].(map[string]interface{})
items, _ := records["items"].([]interface{})
```

## Multi-table Join via Expands

```go
query := `
  query PostsWithAuthors {
    records(
      collection: "posts",
      expand: ["author", "author.profile"],
      sort: "-created"
    ) {
      items {
        id
        data  // expanded relations live under data.expand
      }
    }
  }
`

result, err := client.GraphQL.Query(query, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Conditional Query with Variables

```go
query := `
  query FilteredOrders($minTotal: Float!, $state: String!) {
    records(
      collection: "orders",
      filter: "total >= $minTotal && status = $state",
      sort: "created"
    ) {
      items { id data }
    }
  }
`

variables := map[string]interface{}{
    "minTotal": 100.0,
    "state":    "paid",
}

result, err := client.GraphQL.Query(query, variables, nil, nil)
```

## Create a Record

```go
mutation := `
  mutation CreatePost($data: JSON!) {
    createRecord(collection: "posts", data: $data, expand: ["author"]) {
      id
      data
    }
  }
`

data := map[string]interface{}{
    "title":  "Hello",
    "author": "USER_ID",
}

variables := map[string]interface{}{
    "data": data,
}

result, err := client.GraphQL.Query(mutation, variables, nil, nil)
```

## Update a Record

```go
mutation := `
  mutation UpdatePost($id: ID!, $data: JSON!) {
    updateRecord(collection: "posts", id: $id, data: $data) {
      id
      data
    }
  }
`

variables := map[string]interface{}{
    "id": "POST_ID",
    "data": map[string]interface{}{
        "title": "Updated title",
    },
}

result, err := client.GraphQL.Query(mutation, variables, nil, nil)
```

## Delete a Record

```go
mutation := `
  mutation DeletePost($id: ID!) {
    deleteRecord(collection: "posts", id: $id) {
      id
    }
  }
`

variables := map[string]interface{}{
    "id": "POST_ID",
}

result, err := client.GraphQL.Query(mutation, variables, nil, nil)
```

## Related Documentation

- [API Records](./API_RECORDS.md) - REST API for records
- [Collections](./COLLECTIONS.md) - Collection configuration

