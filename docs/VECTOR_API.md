# Vector Database API - Go SDK Documentation

## Overview

Vector database operations for semantic search, RAG (Retrieval-Augmented Generation), and AI applications.

> **Note**: Vector operations are currently implemented using sqlite-vec but are designed with abstraction in mind to support future vector database providers.

The Vector API provides a unified interface for working with vector embeddings, enabling you to:
- Store and search vector embeddings
- Perform similarity search
- Build RAG applications
- Create recommendation systems
- Enable semantic search capabilities

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
    
    // Authenticate as superuser (vectors require superuser auth)
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Types

### VectorDocument

A vector document with embedding, metadata, and optional content.

```go
type VectorDocument struct {
    ID       string                 // Unique identifier (auto-generated if not provided)
    Vector   []float64              // The vector embedding
    Metadata map[string]interface{} // Optional metadata (key-value pairs)
    Content  string                 // Optional text content
}
```

### VectorSearchOptions

Options for vector similarity search.

```go
type VectorSearchOptions struct {
    QueryVector     []float64              // Query vector to search for
    Limit           *int                    // Max results (default: 10, max: 100)
    Filter          map[string]interface{}  // Optional metadata filter
    MinScore        *float64               // Minimum similarity score threshold
    MaxDistance     *float64               // Maximum distance threshold
    IncludeDistance *bool                  // Include distance in results
    IncludeContent  *bool                  // Include full document content
}
```

## Collection Management

### Create Collection

Create a new vector collection with specified dimension and distance metric.

```go
config := bosbase.VectorCollectionConfig{
    Dimension: 384,      // Vector dimension (default: 384)
    Distance:  "cosine",  // Distance metric: "cosine" (default), "l2", "dot"
}

_, err := client.Vectors.CreateCollection("documents", config, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Minimal example (uses defaults)
_, err = client.Vectors.CreateCollection("documents", bosbase.VectorCollectionConfig{}, nil, nil)
```

**Parameters:**
- `name` (string): Collection name
- `config` (VectorCollectionConfig, optional):
  - `Dimension` (int, optional): Vector dimension. Default: 384
  - `Distance` (string, optional): Distance metric. Default: "cosine"
  - Options: "cosine", "l2", "dot"

### List Collections

Get all available vector collections.

```go
collections, err := client.Vectors.ListCollections(nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, collection := range collections {
    name := collection.Name
    count := collection.Count
    fmt.Printf("%s: %d vectors\n", name, *count)
}
```

### Update Collection

Update a vector collection configuration (distance metric and options).
Note: Collection name and dimension cannot be changed after creation.

```go
config := bosbase.VectorCollectionConfig{
    Distance: "l2", // Change from cosine to L2
}

_, err := client.Vectors.UpdateCollection("documents", config, nil, nil)
```

### Delete Collection

Delete a vector collection and all its data.

```go
err := client.Vectors.DeleteCollection("documents", nil, nil)
if err != nil {
    log.Fatal(err)
}
```

**⚠️ Warning**: This permanently deletes the collection and all vectors in it!

## Insert Vectors

### Insert Single Document

```go
doc := bosbase.VectorDocument{
    ID:      "doc1",
    Vector:  []float64{0.1, 0.2, 0.3, /* ... */},
    Content: "Document content",
    Metadata: map[string]interface{}{
        "title": "My Document",
        "category": "tech",
    },
}

result, err := client.Vectors.Insert(doc, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Inserted document with ID: %s\n", result.ID)
```

### Batch Insert

```go
docs := []bosbase.VectorDocument{
    {
        ID:     "doc1",
        Vector: []float64{0.1, 0.2, 0.3},
        Content: "First document",
    },
    {
        ID:     "doc2",
        Vector: []float64{0.4, 0.5, 0.6},
        Content: "Second document",
    },
}

opts := bosbase.VectorBatchInsertOptions{
    Documents: docs,
}

result, err := client.Vectors.BatchInsert(opts, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Inserted: %d, Failed: %d\n", result.InsertedCount, result.FailedCount)
```

## Search Vectors

### Basic Search

```go
queryVector := []float64{0.1, 0.2, 0.3, /* ... */}
limit := 10

opts := bosbase.VectorSearchOptions{
    QueryVector: queryVector,
    Limit:       &limit,
}

results, err := client.Vectors.Search(opts, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, result := range results.Results {
    fmt.Printf("Score: %.3f, ID: %s\n", result.Score, result.Document.ID)
    fmt.Printf("Content: %s\n", result.Document.Content)
}
```

### Search with Filters

```go
limit := 5
minScore := 0.7

opts := bosbase.VectorSearchOptions{
    QueryVector: queryVector,
    Limit:       &limit,
    Filter: map[string]interface{}{
        "category": "tech",
    },
    MinScore: &minScore,
}

results, err := client.Vectors.Search(opts, "documents", nil, nil)
```

## Update and Delete

### Update Document

```go
doc := bosbase.VectorDocument{
    Vector:  newVector,
    Content: "Updated content",
    Metadata: map[string]interface{}{
        "title": "Updated Title",
    },
}

_, err := client.Vectors.Update("doc1", doc, "documents", nil, nil)
```

### Delete Document

```go
err := client.Vectors.Delete("doc1", "documents", nil, nil)
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
    config := bosbase.VectorCollectionConfig{
        Dimension: 384,
        Distance:  "cosine",
    }
    _, err = client.Vectors.CreateCollection("documents", config, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Insert document
    doc := bosbase.VectorDocument{
        ID:      "doc1",
        Vector:  []float64{0.1, 0.2, 0.3}, // Example vector
        Content: "Sample document content",
        Metadata: map[string]interface{}{
            "title": "Sample Document",
        },
    }
    
    _, err = client.Vectors.Insert(doc, "documents", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Search
    limit := 10
    opts := bosbase.VectorSearchOptions{
        QueryVector: []float64{0.1, 0.2, 0.3},
        Limit:       &limit,
    }
    
    results, err := client.Vectors.Search(opts, "documents", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found %d results\n", len(results.Results))
    for _, result := range results.Results {
        fmt.Printf("Score: %.3f - %s\n", result.Score, result.Document.Content)
    }
}
```

## Related Documentation

- [LLM Documents](./LLM_DOCUMENTS.md) - LLM document store
- [LangChaingo API](./LANGCHAINGO_API.md) - LangChainGo workflows

