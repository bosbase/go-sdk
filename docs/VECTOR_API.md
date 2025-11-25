# Vector Database API - Go SDK Documentation

Vector database operations for semantic search, RAG (Retrieval-Augmented Generation), and AI applications.

> **Note**: Vector operations are currently implemented using sqlite-vec but are designed with abstraction in mind to support future vector database providers.

## Overview

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
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    // Authenticate as superuser (vectors require superuser auth)
    _, err := pb.Admins.AuthWithPassword("admin@example.com", "password", "", "", nil, nil, nil)
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
    ID       string
    Vector   []float64
    Metadata map[string]interface{}
    Content  string
}
```

### VectorSearchOptions
Options for vector similarity search.

```go
type VectorSearchOptions struct {
    QueryVector     []float64
    Limit           *int
    Filter          map[string]interface{}
    MinScore        *float64
    MaxDistance     *float64
    IncludeDistance *bool
    IncludeContent  *bool
}
```

### VectorSearchResult
Result from a similarity search.

```go
type VectorSearchResult struct {
    Document VectorDocument
    Score    float64
    Distance *float64
}
```

## Collection Management

### Create Collection

Create a new vector collection with specified dimension and distance metric.

```go
err := pb.Vectors.CreateCollection("documents", bosbase.VectorCollectionConfig{
    Dimension: 384,      // Vector dimension (default: 384)
    Distance:  "cosine",  // Distance metric: 'cosine' (default), 'l2', 'dot'
}, nil, nil)

// Minimal example (uses defaults)
err = pb.Vectors.CreateCollection("documents", bosbase.VectorCollectionConfig{}, nil, nil)
```

**Parameters:**
- `name` (string): Collection name
- `config` (VectorCollectionConfig):
  - `Dimension` (int, optional): Vector dimension. Default: 384
  - `Distance` (string, optional): Distance metric. Default: 'cosine'
  - Options: 'cosine', 'l2', 'dot'

### List Collections

Get all available vector collections.

```go
collections, err := pb.Vectors.ListCollections(nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, collection := range collections {
    fmt.Printf("%s: %d vectors\n", collection.Name, *collection.Count)
}
```

**Response:**
```go
[]VectorCollectionInfo{
    {
        Name:      string,
        Count:     *int,
        Dimension: *int,
    },
}
```

### Update Collection

Update a vector collection configuration (distance metric and options).
Note: Collection name and dimension cannot be changed after creation.

```go
err := pb.Vectors.UpdateCollection("documents", bosbase.VectorCollectionConfig{
    Distance: "l2", // Change from cosine to L2
}, nil, nil)

// Update with options
err = pb.Vectors.UpdateCollection("documents", bosbase.VectorCollectionConfig{
    Distance: "inner_product",
    Options: map[string]interface{}{
        "customOption": "value",
    },
}, nil, nil)
```

### Delete Collection

Delete a vector collection and all its data.

```go
err := pb.Vectors.DeleteCollection("documents", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

**⚠️ Warning**: This permanently deletes the collection and all vectors in it!

## Document Operations

### Insert Document

Insert a single vector document.

```go
// With custom ID
result, err := pb.Vectors.Insert(bosbase.VectorDocument{
    ID:     "doc_001",
    Vector: []float64{0.1, 0.2, 0.3, 0.4},
    Metadata: map[string]interface{}{
        "category": "tech",
        "tags":     []string{"AI", "ML"},
    },
    Content: "Document about machine learning",
}, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Inserted:", result.ID)

// Without ID (auto-generated)
result2, err := pb.Vectors.Insert(bosbase.VectorDocument{
    Vector:  []float64{0.5, 0.6, 0.7, 0.8},
    Content: "Another document",
}, "documents", nil, nil)
```

**Response:**
```go
VectorInsertResponse{
    ID:      string,
    Success: bool,
}
```

### Batch Insert

Insert multiple vector documents efficiently.

```go
limit := 10
skipDuplicates := true
result, err := pb.Vectors.BatchInsert(bosbase.VectorBatchInsertOptions{
    Documents: []bosbase.VectorDocument{
        {
            Vector:   []float64{0.1, 0.2, 0.3},
            Metadata: map[string]interface{}{"cat": "A"},
            Content:  "Doc A",
        },
        {
            Vector:   []float64{0.4, 0.5, 0.6},
            Metadata: map[string]interface{}{"cat": "B"},
            Content:  "Doc B",
        },
        {
            Vector:   []float64{0.7, 0.8, 0.9},
            Metadata: map[string]interface{}{"cat": "A"},
            Content:  "Doc C",
        },
    },
    SkipDuplicates: &skipDuplicates,
}, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Inserted: %d\n", result.InsertedCount)
fmt.Printf("Failed: %d\n", result.FailedCount)
fmt.Println("IDs:", result.IDs)
```

**Response:**
```go
VectorBatchInsertResponse{
    InsertedCount: int,
    FailedCount:   int,
    IDs:           []string,
    Errors:        []string,
}
```

### Get Document

Retrieve a vector document by ID.

```go
doc, err := pb.Vectors.Get("doc_001", "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Vector:", doc.Vector)
fmt.Println("Content:", doc.Content)
fmt.Println("Metadata:", doc.Metadata)
```

### Update Document

Update an existing vector document.

```go
// Update all fields
_, err := pb.Vectors.Update("doc_001", bosbase.VectorDocument{
    Vector: []float64{0.9, 0.8, 0.7, 0.6},
    Metadata: map[string]interface{}{
        "updated": true,
    },
    Content: "Updated content",
}, "documents", nil, nil)

// Partial update (only metadata and content)
_, err = pb.Vectors.Update("doc_001", bosbase.VectorDocument{
    Metadata: map[string]interface{}{
        "category": "updated",
    },
    Content: "New content",
}, "documents", nil, nil)
```

### Delete Document

Delete a vector document.

```go
err := pb.Vectors.Delete("doc_001", "documents", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### List Documents

List all documents in a collection with pagination.

```go
// Get first page
page := 1
perPage := 100
result, err := pb.Vectors.List("documents", &page, &perPage, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Page %v of %v\n", result["page"], result["totalPages"])
if items, ok := result["items"].([]interface{}); ok {
    for _, item := range items {
        if doc, ok := item.(map[string]interface{}); ok {
            fmt.Println(doc["id"], doc["content"])
        }
    }
}
```

**Response:**
```go
map[string]interface{}{
    "page":       int,
    "perPage":    int,
    "totalItems": int,
    "totalPages": int,
    "items":      []VectorDocument,
}
```

## Vector Search

### Basic Search

Perform similarity search on vectors.

```go
limit := 10
results, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
    QueryVector: []float64{0.1, 0.2, 0.3, 0.4},
    Limit:       &limit,
}, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, result := range results.Results {
    fmt.Printf("Score: %.2f - %s\n", result.Score, result.Document.Content)
}
```

### Advanced Search

```go
limit := 20
minScore := 0.7
maxDistance := 0.3
includeDistance := true
includeContent := true

results, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
    QueryVector:     []float64{0.1, 0.2, 0.3, 0.4},
    Limit:           &limit,
    MinScore:        &minScore,
    MaxDistance:     &maxDistance,
    IncludeDistance: &includeDistance,
    IncludeContent:  &includeContent,
    Filter: map[string]interface{}{
        "category": "tech",
    },
}, "documents", nil, nil)
if err != nil {
    log.Fatal(err)
}

if results.TotalMatches != nil {
    fmt.Printf("Found %d matches", *results.TotalMatches)
}
if results.QueryTime != nil {
    fmt.Printf(" in %dms\n", *results.QueryTime)
}

for _, r := range results.Results {
    fmt.Printf("Score: %.2f", r.Score)
    if r.Distance != nil {
        fmt.Printf(", Distance: %.2f", *r.Distance)
    }
    fmt.Printf("\nContent: %s\n", r.Document.Content)
}
```

**Response:**
```go
VectorSearchResponse{
    Results:      []VectorSearchResult,
    TotalMatches: *int,
    QueryTime:    *int,
}
```

## Common Use Cases

### Semantic Search

```go
// 1. Generate embeddings for your documents
documents := []struct {
    Text string
    ID   string
}{
    {Text: "Introduction to machine learning", ID: "doc1"},
    {Text: "Deep learning fundamentals", ID: "doc2"},
    {Text: "Natural language processing", ID: "doc3"},
}

for _, doc := range documents {
    // Generate embedding using your model
    embedding := generateEmbedding(doc.Text) // Your function
    
    _, err := pb.Vectors.Insert(bosbase.VectorDocument{
        ID:      doc.ID,
        Vector:  embedding,
        Content: doc.Text,
        Metadata: map[string]interface{}{
            "type": "tutorial",
        },
    }, "articles", nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}

// 2. Search
queryEmbedding := generateEmbedding("What is AI?")
limit := 5
minScore := 0.75
results, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
    QueryVector: queryEmbedding,
    Limit:       &limit,
    MinScore:    &minScore,
}, "articles", nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, r := range results.Results {
    fmt.Printf("%.2f: %s\n", r.Score, r.Document.Content)
}
```

### RAG (Retrieval-Augmented Generation)

```go
func retrieveContext(pb *bosbase.BosBase, query string, limit int) ([]string, error) {
    queryEmbedding := generateEmbedding(query)
    
    minScore := 0.75
    includeContent := true
    results, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
        QueryVector:    queryEmbedding,
        Limit:          &limit,
        MinScore:       &minScore,
        IncludeContent: &includeContent,
    }, "knowledge_base", nil, nil)
    if err != nil {
        return nil, err
    }
    
    var contexts []string
    for _, r := range results.Results {
        contexts = append(contexts, r.Document.Content)
    }
    return contexts, nil
}

// Use with your LLM
context, err := retrieveContext(pb, "What are best practices for security?", 5)
if err != nil {
    log.Fatal(err)
}
answer := llmGenerate(context, userQuery) // Your LLM function
```

### Recommendation System

```go
// Store user profile embeddings
_, err := pb.Vectors.Insert(bosbase.VectorDocument{
    ID:     userID,
    Vector: userProfileEmbedding,
    Metadata: map[string]interface{}{
        "preferences": []string{"tech", "science"},
        "demographics": map[string]interface{}{
            "age":      30,
            "location": "US",
        },
    },
}, "users", nil, nil)

// Find similar users
limit := 20
includeDistance := true
similarUsers, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
    QueryVector:     currentUserEmbedding,
    Limit:           &limit,
    IncludeDistance: &includeDistance,
}, "users", nil, nil)

// Generate recommendations based on similar users
recommendations := generateRecommendations(similarUsers)
```

### Multi-modal Search

```go
// Store embeddings from different sources
_, err := pb.Vectors.Insert(bosbase.VectorDocument{
    ID:     "image_001",
    Vector: imageEmbedding,
    Metadata: map[string]interface{}{
        "type": "image",
        "url":  "https://...",
    },
    Content: "Description of the image",
}, "media", nil, nil)

_, err = pb.Vectors.Insert(bosbase.VectorDocument{
    ID:     "video_001",
    Vector: videoEmbedding,
    Metadata: map[string]interface{}{
        "type":     "video",
        "duration": 120,
    },
    Content: "Video transcript",
}, "media", nil, nil)

// Search across all media types
limit := 10
includeContent := true
results, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
    QueryVector:    queryEmbedding,
    Limit:          &limit,
    IncludeContent: &includeContent,
}, "media", nil, nil)
```

## Best Practices

### Vector Dimensions

Choose the right dimension for your use case:

- **OpenAI embeddings**: 1536 (`text-embedding-3-large`)
- **Sentence Transformers**: 384-768
  - `all-MiniLM-L6-v2`: 384
  - `all-mpnet-base-v2`: 768
- **Custom models**: Match your model's output

### Distance Metrics

| Metric | Best For | Notes |
|--------|----------|-------|
| `cosine` | Text embeddings | Works well with normalized vectors |
| `l2` | General similarity | Euclidean distance |
| `dot` | Performance | Requires normalized vectors |

### Performance Tips

1. **Use batch insert** for multiple vectors
2. **Set appropriate limits** to avoid excessive results
3. **Use metadata filtering** to narrow search space
4. **Enable indexes** (automatic with sqlite-vec)

### Security

- All vector endpoints require superuser authentication
- Never expose credentials in client-side code
- Use environment variables for sensitive data

## Error Handling

```go
limit := 10
_, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
    QueryVector: []float64{0.1, 0.2, 0.3},
    Limit:       &limit,
}, "documents", nil, nil)
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        switch clientErr.Status {
        case 404:
            fmt.Println("Collection not found")
        case 400:
            fmt.Println("Invalid request:", clientErr.Response)
        default:
            fmt.Println("Error:", err)
        }
    } else {
        fmt.Println("Error:", err)
    }
}
```

## Examples

### Complete RAG Application

```go
package main

import (
    "fmt"
    "log"
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    // Initialize
    _, err := pb.Admins.AuthWithPassword("admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // 1. Create knowledge base collection
    err = pb.Vectors.CreateCollection("knowledge_base", bosbase.VectorCollectionConfig{
        Dimension: 1536, // OpenAI dimensions
        Distance:  "cosine",
    }, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // 2. Index documents
    documents := []struct {
        ID      string
        Content string
        Source  string
        Topic   string
    }{
        {ID: "doc1", Content: "Content 1", Source: "source1", Topic: "topic1"},
        // ... more documents
    }
    
    for _, doc := range documents {
        // Generate OpenAI embedding (your function)
        embedding := generateOpenAIEmbedding(doc.Content)
        
        _, err := pb.Vectors.Insert(bosbase.VectorDocument{
            ID:      doc.ID,
            Vector:  embedding,
            Content: doc.Content,
            Metadata: map[string]interface{}{
                "source": doc.Source,
                "topic":  doc.Topic,
            },
        }, "knowledge_base", nil, nil)
        if err != nil {
            log.Fatal(err)
        }
    }
    
    // 3. RAG Query
    answer, err := ask(pb, "What is machine learning?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(answer)
}

func ask(pb *bosbase.BosBase, question string) (string, error) {
    // Generate query embedding
    embedding := generateOpenAIEmbedding(question)
    
    // Search for relevant context
    limit := 5
    minScore := 0.8
    includeContent := true
    results, err := pb.Vectors.Search(bosbase.VectorSearchOptions{
        QueryVector:    embedding,
        Limit:          &limit,
        MinScore:       &minScore,
        IncludeContent: &includeContent,
        Filter: map[string]interface{}{
            "topic": "relevant_topic",
        },
    }, "knowledge_base", nil, nil)
    if err != nil {
        return "", err
    }
    
    // Build context
    var context string
    for _, r := range results.Results {
        context += r.Document.Content + "\n\n"
    }
    
    // Generate answer with LLM (your function)
    answer := llmGenerate(context, question)
    return answer, nil
}

// Placeholder functions - implement with your actual embedding/LLM services
func generateOpenAIEmbedding(text string) []float64 {
    // Your OpenAI embedding generation
    return []float64{}
}

func llmGenerate(context, question string) string {
    // Your LLM generation
    return ""
}
```

## References

- [sqlite-vec Documentation](https://alexgarcia.xyz/sqlite-vec)
- [sqlite-vec with rqlite](https://alexgarcia.xyz/sqlite-vec/rqlite.html)
- [Vector Implementation Guide](../../VECTOR_IMPLEMENTATION.md)
- [Vector Setup Guide](../../VECTOR_SETUP_GUIDE.md)

