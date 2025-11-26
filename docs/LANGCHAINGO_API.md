# LangChaingo API - Go SDK Documentation

## Overview

BosBase exposes the `/api/langchaingo` endpoints so you can run LangChainGo powered workflows without leaving the platform. The Go SDK wraps these endpoints with the `client.LangChaingo` service.

The service exposes four high-level methods:

| Method | HTTP Endpoint | Description |
| --- | --- | --- |
| `client.LangChaingo.Completions()` | `POST /api/langchaingo/completions` | Runs a chat/completion call using the configured LLM provider. |
| `client.LangChaingo.RAG()` | `POST /api/langchaingo/rag` | Runs a retrieval-augmented generation pass over an `llmDocuments` collection. |
| `client.LangChaingo.QueryDocuments()` | `POST /api/langchaingo/documents/query` | Asks an OpenAI-backed chain to answer questions over `llmDocuments` and optionally return matched sources. |
| `client.LangChaingo.SQL()` | `POST /api/langchaingo/sql` | Lets OpenAI draft and execute SQL against your BosBase database, then returns the results. |

Each method accepts an optional `model` block:

```go
type LangChaingoModelConfig struct {
    Provider string // "openai" | "ollama" | string
    Model    string
    APIKey   string
    BaseURL  string
}
```

If you omit the `model` section, BosBase defaults to `provider: "openai"` and `model: "gpt-4o-mini"` with credentials read from the server environment. Passing an `apiKey` lets you override server defaults on a per-request basis.

## Text + Chat Completions

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
    
    model := &bosbase.LangChaingoModelConfig{
        Provider: "openai",
        Model:    "gpt-4o-mini",
    }
    
    req := bosbase.LangChaingoCompletionRequest{
        Model: model,
        Messages: []bosbase.LangChaingoCompletionMessage{
            {Role: "system", Content: "Answer in one sentence."},
            {Role: "user", Content: "Explain Rayleigh scattering."},
        },
        Temperature: 0.2,
    }
    
    completion, err := client.LangChaingo.Completions(req, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(completion.Content)
}
```

The completion response mirrors the LangChainGo `ContentResponse` shape, so you can inspect the `FunctionCall`, `ToolCalls`, or `GenerationInfo` fields when you need more than plain text.

## Retrieval-Augmented Generation (RAG)

Pair the LangChaingo endpoints with the `/api/llm-documents` store to build RAG workflows. The backend automatically uses the chromem-go collection configured for the target LLM collection.

```go
filters := &bosbase.LangChaingoRAGFilters{
    Where: map[string]interface{}{
        "topic": "physics",
    },
}

req := bosbase.LangChaingoRAGRequest{
    Collection:    "knowledge-base",
    Question:      "Why is the sky blue?",
    TopK:          4,
    ReturnSources: true,
    Filters:       filters,
}

answer, err := client.LangChaingo.RAG(req, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(answer.Answer)
for _, source := range answer.Sources {
    fmt.Printf("Score: %.3f, Title: %s\n", source.Score, source.Metadata["title"])
}
```

Set `PromptTemplate` when you want to control how the retrieved context is stuffed into the answer prompt:

```go
req := bosbase.LangChaingoRAGRequest{
    Collection:    "knowledge-base",
    Question:      "Summarize the explanation below in 2 sentences.",
    PromptTemplate: "Context:\n{{.context}}\n\nQuestion: {{.question}}\nSummary:",
}

answer, err := client.LangChaingo.RAG(req, nil, nil)
```

## LLM Document Queries

> **Note**: This interface is only available to superusers.

When you want to pose a question to a specific `llmDocuments` collection and have LangChaingo+OpenAI synthesize an answer, use `QueryDocuments`. It mirrors the RAG arguments but takes a `Query` field:

```go
req := bosbase.LangChaingoRAGRequest{
    Collection:    "knowledge-base",
    Query:         "List three bullet points about Rayleigh scattering.",
    TopK:          3,
    ReturnSources: true,
}

response, err := client.LangChaingo.QueryDocuments(req, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(response.Answer)
fmt.Println(response.Sources)
```

## SQL Generation + Execution

> **Important Notes**:
> - This interface is only available to superusers. Requests authenticated with regular `users` tokens return a `401 Unauthorized`.
> - It is recommended to execute query statements (SELECT) only.
> - **Do not use this interface for adding or modifying table structures.** Collection interfaces should be used instead for managing database schema.
> - Directly using this interface for initializing table structures and adding or modifying database tables will cause errors that prevent the automatic generation of APIs.

Superuser tokens (`_superusers` records) can ask LangChaingo to have OpenAI propose a SQL statement, execute it, and return both the generated SQL and execution output.

```go
req := bosbase.LangChaingoSQLRequest{
    Query:  "Add a demo project row if it doesn't exist, then list the 5 most recent projects.",
    Tables: []string{"projects"}, // optional hint to limit which tables the model sees
    TopK:   5,
}

result, err := client.LangChaingo.SQL(req, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.SQL)    // Generated SQL
fmt.Println(result.Answer) // Model's summary of the execution
fmt.Println(result.Columns, result.Rows)
```

Use `Tables` to restrict which table definitions and sample rows are passed to the model, and `TopK` to control how many rows the model should target when building queries. You can also pass the optional `Model` block described above to override the default OpenAI model or key for this call.

## Related Documentation

- [LLM Documents](./LLM_DOCUMENTS.md) - LLM document store
- [Vector API](./VECTOR_API.md) - Vector embeddings

