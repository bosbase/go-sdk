# Pub/Sub API - Go SDK Documentation

## Overview

BosBase exposes a lightweight WebSocket-based publish/subscribe channel so SDK users can push and receive custom messages. The Go backend uses the `ws` transport and persists each published payload in the `_pubsub_messages` table so every node in a cluster can replay and fan-out messages to its local subscribers.

- **Endpoint**: `/api/pubsub` (WebSocket)
- **Auth**: the SDK automatically forwards `authStore.token` as a `token` query parameter; cookie-based auth also works. Anonymous clients may subscribe, but publishing requires an authenticated token.
- **Reliability**: automatic reconnect with topic re-subscription; messages are stored in the database and broadcasted to all connected nodes.

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://127.0.0.1:8090")
    defer client.Close()
    
    // Subscribe to a topic
    unsubscribe, err := client.PubSub.Subscribe("chat/general", func(msg map[string]interface{}) {
        topic, _ := msg["topic"].(string)
        data, _ := msg["data"].(map[string]interface{})
        fmt.Printf("Message on %s: %v\n", topic, data)
    }, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer unsubscribe()
    
    // Publish to a topic (resolves when the server stores and accepts it)
    ack, err := client.PubSub.Publish("chat/general", map[string]interface{}{
        "text": "Hello team!",
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    created, _ := ack["created"].(string)
    fmt.Printf("Published at %s\n", created)
}
```

## API Surface

- `client.PubSub.Publish(topic, data, body, query, headers)` → `map[string]interface{}` (returns `{ id, topic, created }`)
- `client.PubSub.Subscribe(topic, handler, query, headers)` → `func()` (unsubscribe function)
- `client.PubSub.Unsubscribe(topic)` - Unsubscribe from a specific topic
- `client.PubSub.Disconnect()` - Explicitly close the socket and clear pending requests
- `client.PubSub.IsConnected()` - Check current WebSocket state

## Notes for Clusters

- Messages are written to `_pubsub_messages` with a timestamp; every running node polls the table and pushes new rows to its connected WebSocket clients.
- Old pub/sub rows are cleaned up automatically after a day to keep the table small.
- If a node restarts, it resumes from the latest message and replays new rows as they are inserted, so connected clients on other nodes stay in sync.

## Complete Example

```go
func setupChatRoom(client *bosbase.BosBase, roomID string) (func(), error) {
    topic := fmt.Sprintf("chat/%s", roomID)
    
    unsubscribe, err := client.PubSub.Subscribe(topic, func(msg map[string]interface{}) {
        data, _ := msg["data"].(map[string]interface{})
        text, _ := data["text"].(string)
        user, _ := data["user"].(string)
        fmt.Printf("[%s] %s: %s\n", roomID, user, text)
    }, nil, nil)
    
    return unsubscribe, err
}

// Usage
unsubscribe, _ := setupChatRoom(client, "general")
defer unsubscribe()

// Publish a message
_, err := client.PubSub.Publish("chat/general", map[string]interface{}{
    "text": "Hello everyone!",
    "user": "user123",
}, nil, nil, nil)
```

## Related Documentation

- [Realtime API](./REALTIME.md) - Real-time record subscriptions
- [Authentication](./AUTHENTICATION.md) - User authentication

