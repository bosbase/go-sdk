# LLM Document API - Go SDK Documentation

## Overview

The `LLMDocumentService` wraps the `/api/llm-documents` endpoints that are backed by the embedded chromem-go vector store (persisted in rqlite). Each document contains text content, optional metadata and an embedding vector that can be queried with semantic search.

## Getting Started

```go
package main

import (
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    // Authenticate as superuser
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a logical namespace for your documents
    metadata := map[string]string{
        "domain": "internal",
    }
    err = client.LLMDocuments.CreateCollection("knowledge-base", metadata, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Insert Documents

```go
doc := bosbase.LLMDocument{
    Content: "Leaves are green because chlorophyll absorbs red and blue light.",
    Metadata: map[string]interface{}{
        "topic": "biology",
    },
}

result, err := client.LLMDocuments.Insert("knowledge-base", doc, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Insert with custom ID
doc2 := bosbase.LLMDocument{
    ID:      "sky",
    Content: "The sky is blue because of Rayleigh scattering.",
    Metadata: map[string]interface{}{
        "topic": "physics",
    },
}

_, err = client.LLMDocuments.Insert("knowledge-base", doc2, nil, nil)
```

## Query Documents

```go
query := bosbase.LLMDocumentQuery{
    QueryText: "Why is the sky blue?",
    Limit:     3,
    Where: map[string]interface{}{
        "topic": "physics",
    },
}

result, err := client.LLMDocuments.Query("knowledge-base", query, nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, match := range result.Results {
    fmt.Printf("ID: %s, Similarity: %.3f\n", match.ID, match.Similarity)
    fmt.Printf("Content: %s\n", match.Content)
}
```

## Manage Documents

### Update a Document

```go
update := bosbase.LLMDocumentUpdate{
    Metadata: map[string]interface{}{
        "topic":   "physics",
        "reviewed": "true",
    },
}

_, err := client.LLMDocuments.Update("knowledge-base", "sky", update, nil, nil)
```

### List Documents with Pagination

```go
query := map[string]interface{}{
    "page":    1,
    "perPage": 25,
}

page, err := client.LLMDocuments.List("knowledge-base", query, nil)
if err != nil {
    log.Fatal(err)
}

items, _ := page["items"].([]interface{})
totalItems, _ := page["totalItems"].(float64)

fmt.Printf("Total items: %.0f\n", totalItems)
for _, item := range items {
    doc, _ := item.(map[string]interface{})
    id, _ := doc["id"].(string)
    content, _ := doc["content"].(string)
    fmt.Printf("%s: %s\n", id, content)
}
```

### Delete Document

```go
err := client.LLMDocuments.Delete("knowledge-base", "sky", nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Collection Management

### List Collections

```go
collections, err := client.LLMDocuments.ListCollections(nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, collection := range collections {
    name, _ := collection["name"].(string)
    fmt.Printf("Collection: %s\n", name)
}
```

### Delete Collection

```go
err := client.LLMDocuments.DeleteCollection("knowledge-base", nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    // Authenticate
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create collection
    metadata := map[string]string{
        "domain": "internal",
    }
    err = client.LLMDocuments.CreateCollection("knowledge-base", metadata, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Insert documents
    doc1 := bosbase.LLMDocument{
        Content: "The sky is blue because of Rayleigh scattering.",
        Metadata: map[string]interface{}{
            "topic": "physics",
        },
    }
    
    _, err = client.LLMDocuments.Insert("knowledge-base", doc1, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Query documents
    query := bosbase.LLMDocumentQuery{
        QueryText: "Why is the sky blue?",
        Limit:     3,
    }
    
    result, err := client.LLMDocuments.Query("knowledge-base", query, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, match := range result.Results {
        fmt.Printf("Similarity: %.3f\n", match.Similarity)
        fmt.Printf("Content: %s\n", match.Content)
    }
}
```

## HTTP Endpoints

| Method | Path | Purpose |
| --- | --- | --- |
| `GET /api/llm-documents/collections` | List collections |
| `POST /api/llm-documents/collections/{name}` | Create collection |
| `DELETE /api/llm-documents/collections/{name}` | Delete collection |
| `GET /api/llm-documents/{collection}` | List documents |
| `POST /api/llm-documents/{collection}` | Insert document |
| `GET /api/llm-documents/{collection}/{id}` | Fetch document |
| `PATCH /api/llm-documents/{collection}/{id}` | Update document |
| `DELETE /api/llm-documents/{collection}/{id}` | Delete document |
| `POST /api/llm-documents/{collection}/documents/query` | Query by semantic similarity |

## Related Documentation

- [LangChaingo API](./LANGCHAINGO_API.md) - LangChainGo workflows
- [Vector API](./VECTOR_API.md) - Vector embeddings

