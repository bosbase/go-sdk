# BosBase Go SDK

Official Go SDK for interacting with the BosBase API. The API surface mirrors the JavaScript/Python SDKs: collections, records, auth flows, realtime subscriptions, pub/sub, GraphQL, vectors, LangChaingo, backups, cache, cron jobs, settings, files, and more.

## Install

```bash
go get github.com/bosbase/go-sdk
```

## Quick start

```go
package main

import (
    "fmt"

    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://127.0.0.1:8090")
    defer client.Close()

    // Authenticate against an auth collection
    _, _ = client.Collection("users").AuthWithPassword("test@example.com", "123456", "", "", nil, nil, nil)

    // List records
    list, _ := client.Collection("example").GetList(&bosbase.CrudListOptions{Page: 1, PerPage: 10})
    fmt.Println(list)
}
```

## Highlights

- Shared `AuthStore` with JWT decoding and listeners
- HTTP hooks (`BeforeSend`/`AfterSend`) and custom headers per request
- CRUD helpers via `CollectionService`/`RecordService`
- Realtime SSE subscriptions and OAuth2 hand-offs
- WebSocket pub/sub with publish/subscribe helpers
- File utilities (`GetFileURL`, token generation, multipart uploads)
- Vector, LangChaingo, LLM document, cache, batch, backup, cron, settings, logs, GraphQL, and health endpoints

All services live under the root `bosbase` package; constructors mirror the JS SDK naming.
