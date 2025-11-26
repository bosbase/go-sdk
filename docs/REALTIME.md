# Realtime API - Go SDK Documentation

## Overview

The Realtime API enables real-time updates for collection records using **Server-Sent Events (SSE)**. It allows you to subscribe to changes in collections or specific records and receive instant notifications when records are created, updated, or deleted.

**Key Features:**
- Real-time notifications for record changes
- Collection-level and record-level subscriptions
- Automatic connection management and reconnection
- Authorization support
- Subscription options (expand, custom headers, query params)
- Event-driven architecture

**Backend Endpoints:**
- `GET /api/realtime` - Establish SSE connection
- `POST /api/realtime` - Set subscriptions

## How It Works

1. **Connection**: The SDK establishes an SSE connection to `/api/realtime`
2. **Client ID**: Server sends `PB_CONNECT` event with a unique `clientId`
3. **Subscriptions**: Client submits subscription topics via POST request
4. **Events**: Server sends events when matching records change
5. **Reconnection**: SDK automatically reconnects on connection loss

## Basic Usage

### Subscribe to Collection Changes

Subscribe to all changes in a collection:

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
    
    // Subscribe to all changes in the 'posts' collection
    unsubscribe, err := client.Collection("posts").Subscribe("*", func(e map[string]interface{}) {
        action, _ := e["action"].(string)
        record, _ := e["record"].(map[string]interface{})
        fmt.Printf("Action: %s\n", action)
        fmt.Printf("Record: %v\n", record)
    }, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Later, unsubscribe
    defer unsubscribe()
}
```

### Subscribe to Specific Record

Subscribe to changes for a single record:

```go
// Subscribe to changes for a specific post
unsubscribe, err := client.Collection("posts").Subscribe("RECORD_ID", func(e map[string]interface{}) {
    record, _ := e["record"].(map[string]interface{})
    action, _ := e["action"].(string)
    fmt.Printf("Record changed: %v\n", record)
    fmt.Printf("Action: %s\n", action)
}, nil, nil)
if err != nil {
    log.Fatal(err)
}
defer unsubscribe()
```

### Multiple Subscriptions

You can subscribe multiple times to the same or different topics:

```go
// Subscribe to multiple records
unsubscribe1, _ := client.Collection("posts").Subscribe("RECORD_ID_1", handleChange, nil, nil)
unsubscribe2, _ := client.Collection("posts").Subscribe("RECORD_ID_2", handleChange, nil, nil)
unsubscribe3, _ := client.Collection("posts").Subscribe("*", handleAllChanges, nil, nil)

func handleChange(e map[string]interface{}) {
    fmt.Println("Change event:", e)
}

func handleAllChanges(e map[string]interface{}) {
    fmt.Println("Collection-wide change:", e)
}

// Unsubscribe individually
defer unsubscribe1()
defer unsubscribe2()
defer unsubscribe3()
```

## Event Structure

Each event received contains:

```go
{
    "action": "create" | "update" | "delete",  // Action type
    "record": {                                 // Record data
        "id": "RECORD_ID",
        "collectionId": "COLLECTION_ID",
        "collectionName": "collection_name",
        "created": "2023-01-01 00:00:00.000Z",
        "updated": "2023-01-01 00:00:00.000Z",
        // ... other fields
    }
}
```

### PB_CONNECT Event

When the connection is established, you receive a `PB_CONNECT` event:

```go
unsubscribe, err := client.Realtime.Subscribe("PB_CONNECT", func(e map[string]interface{}) {
    clientID, _ := e["clientId"].(string)
    fmt.Printf("Connected! Client ID: %s\n", clientID)
}, nil, nil)
if err != nil {
    log.Fatal(err)
}
defer unsubscribe()
```

## Subscription Topics

### Collection-Level Subscription

Subscribe to all changes in a collection:

```go
// Wildcard subscription - all records in collection
unsubscribe, err := client.Collection("posts").Subscribe("*", handler, nil, nil)
```

**Access Control**: Uses the collection's `ListRule` to determine if the subscriber has access to receive events.

### Record-Level Subscription

Subscribe to changes for a specific record:

```go
// Specific record subscription
unsubscribe, err := client.Collection("posts").Subscribe("RECORD_ID", handler, nil, nil)
```

**Access Control**: Uses the collection's `ViewRule` to determine if the subscriber has access to receive events.

## Subscription Options

You can pass additional options when subscribing:

```go
query := map[string]interface{}{
    "filter": "status = \"published\"",
    "expand": "author",
}
headers := map[string]string{
    "X-Custom-Header": "value",
}

unsubscribe, err := client.Collection("posts").Subscribe("*", handler, query, headers)
```

### Expand Relations

Expand relations in the event data:

```go
query := map[string]interface{}{
    "expand": "author,categories",
}

unsubscribe, err := client.Collection("posts").Subscribe("RECORD_ID", func(e map[string]interface{}) {
    record, _ := e["record"].(map[string]interface{})
    expand, _ := record["expand"].(map[string]interface{})
    author, _ := expand["author"].(map[string]interface{})
    fmt.Println("Author relation expanded:", author)
}, query, nil)
```

### Filter with Query Parameters

Use query parameters for API rule filtering:

```go
query := map[string]interface{}{
    "filter": "status = \"published\"",
}

unsubscribe, err := client.Collection("posts").Subscribe("*", handler, query, nil)
```

## Unsubscribing

### Unsubscribe from Specific Topic

```go
// Remove all subscriptions for a specific record
client.Collection("posts").Unsubscribe("RECORD_ID")

// Remove all wildcard subscriptions for the collection
client.Collection("posts").Unsubscribe("*")
```

### Unsubscribe from All

```go
// Unsubscribe from all subscriptions in the collection
client.Collection("posts").Unsubscribe("")

// Or unsubscribe from everything
client.Realtime.Unsubscribe("")
```

### Unsubscribe Using Returned Function

```go
unsubscribe, err := client.Collection("posts").Subscribe("*", handler, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Later...
unsubscribe()  // Removes this specific subscription
```

## Connection Management

### Connection Status

Check if the realtime connection is established:

```go
// Note: The Go SDK manages connection state internally
// Connection is established automatically when subscribing
```

### Disconnect Handler

Handle disconnection events:

```go
client.Realtime.OnDisconnect = func(activeSubscriptions []string) {
    if len(activeSubscriptions) > 0 {
        fmt.Printf("Connection lost, but subscriptions remain: %v\n", activeSubscriptions)
        // Connection will automatically reconnect
    } else {
        fmt.Println("Intentionally disconnected (no active subscriptions)")
    }
}
```

### Automatic Reconnection

The SDK automatically:
- Reconnects when the connection is lost
- Resubmits all active subscriptions
- Handles network interruptions gracefully
- Closes connection after 5 minutes of inactivity (server-side timeout)

## Authorization

### Authenticated Subscriptions

Subscriptions respect authentication. If you're authenticated, events are filtered based on your permissions:

```go
// Authenticate first
_, err := client.Collection("users").AuthWithPassword(
    "user@example.com", "password123", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Now subscribe - events will respect your permissions
unsubscribe, err := client.Collection("posts").Subscribe("*", handler, nil, nil)
```

### Authorization Rules

- **Collection-level (`*`)**: Uses `ListRule` to determine access
- **Record-level**: Uses `ViewRule` to determine access
- **Superusers**: Can receive all events (if rules allow)
- **Guests**: Only receive events they have permission to see

### Auth State Changes

When authentication state changes, you may need to resubscribe:

```go
// After login/logout, resubscribe to update permissions
_, err := client.Collection("users").AuthWithPassword(
    "user@example.com", "password123", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Re-subscribe to update auth state in realtime connection
unsubscribe, err := client.Collection("posts").Subscribe("*", handler, nil, nil)
```

## Advanced Examples

### Example 1: Real-time Chat

```go
// Subscribe to messages in a chat room
func setupChatRoom(client *bosbase.BosBase, roomID string) (func(), error) {
    query := map[string]interface{}{
        "filter": fmt.Sprintf(`roomId = "%s"`, roomID),
    }
    
    unsubscribe, err := client.Collection("messages").Subscribe("*", func(e map[string]interface{}) {
        action, _ := e["action"].(string)
        record, _ := e["record"].(map[string]interface{})
        
        // Filter for this room only
        recordRoomID, _ := record["roomId"].(string)
        if recordRoomID == roomID {
            if action == "create" {
                displayMessage(record)
            } else if action == "delete" {
                removeMessage(record["id"].(string))
            }
        }
    }, query, nil)
    
    return unsubscribe, err
}

// Usage
unsubscribeChat, err := setupChatRoom(client, "ROOM_ID")
if err != nil {
    log.Fatal(err)
}
defer unsubscribeChat()
```

### Example 2: Real-time Dashboard

```go
// Subscribe to multiple collections
func setupDashboard(client *bosbase.BosBase) error {
    // Posts updates
    query1 := map[string]interface{}{
        "filter": "status = \"published\"",
        "expand": "author",
    }
    _, err := client.Collection("posts").Subscribe("*", func(e map[string]interface{}) {
        action, _ := e["action"].(string)
        record, _ := e["record"].(map[string]interface{})
        
        if action == "create" {
            addPostToFeed(record)
        } else if action == "update" {
            updatePostInFeed(record)
        }
    }, query1, nil)
    if err != nil {
        return err
    }
    
    // Comments updates
    query2 := map[string]interface{}{
        "expand": "user",
    }
    _, err = client.Collection("comments").Subscribe("*", func(e map[string]interface{}) {
        record, _ := e["record"].(map[string]interface{})
        postID, _ := record["postId"].(string)
        updateCommentsCount(postID)
    }, query2, nil)
    
    return err
}
```

### Example 3: User Activity Tracking

```go
// Track changes to a user's own records
func trackUserActivity(client *bosbase.BosBase, userID string) error {
    query := map[string]interface{}{
        "filter": fmt.Sprintf(`author = "%s"`, userID),
    }
    
    _, err := client.Collection("posts").Subscribe("*", func(e map[string]interface{}) {
        action, _ := e["action"].(string)
        record, _ := e["record"].(map[string]interface{})
        title, _ := record["title"].(string)
        
        fmt.Printf("Your post %s: %s\n", action, title)
        
        if action == "update" {
            showNotification("Post updated")
        }
    }, query, nil)
    
    return err
}

// Usage
authRecord := client.AuthStore.Record()
userID, _ := authRecord["id"].(string)
trackUserActivity(client, userID)
```

### Example 4: Connection Monitoring

```go
// Monitor connection state
client.Realtime.OnDisconnect = func(activeSubscriptions []string) {
    if len(activeSubscriptions) > 0 {
        fmt.Println("Connection lost, attempting to reconnect...")
        showConnectionStatus("Reconnecting...")
    }
}

// Monitor connection establishment
_, err := client.Realtime.Subscribe("PB_CONNECT", func(e map[string]interface{}) {
    clientID, _ := e["clientId"].(string)
    fmt.Printf("Connected to realtime: %s\n", clientID)
    showConnectionStatus("Connected")
}, nil, nil)
```

## Error Handling

```go
unsubscribe, err := client.Collection("posts").Subscribe("*", handler, nil, nil)
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        if clientErr.Status == 403 {
            fmt.Println("Permission denied")
        } else if clientErr.Status == 404 {
            fmt.Println("Collection not found")
        } else {
            fmt.Printf("Subscription error: %v\n", err)
        }
    } else {
        fmt.Printf("Subscription error: %v\n", err)
    }
    return
}
defer unsubscribe()
```

## Best Practices

1. **Unsubscribe When Done**: Always unsubscribe when subscriptions are no longer needed
2. **Handle Disconnections**: Implement `OnDisconnect` handler for better UX
3. **Filter Server-Side**: Use query parameters to filter events server-side when possible
4. **Limit Subscriptions**: Don't subscribe to more collections than necessary
5. **Use Record-Level When Possible**: Prefer record-level subscriptions over collection-level when you only need specific records
6. **Monitor Connection**: Track connection state for debugging and user feedback
7. **Handle Errors**: Wrap subscriptions in error handling
8. **Respect Permissions**: Understand that events respect API rules and permissions

## Limitations

- **Maximum Subscriptions**: Up to 1000 subscriptions per client
- **Topic Length**: Maximum 2500 characters per topic
- **Idle Timeout**: Connection closes after 5 minutes of inactivity
- **Network Dependency**: Requires stable network connection

## Troubleshooting

### Connection Not Establishing

```go
// Connection is established automatically when subscribing
// If connection fails, check network connectivity and server status
```

### Events Not Received

1. Check API rules - you may not have permission
2. Verify subscription is active
3. Check network connectivity
4. Review server logs for errors

### Memory Leaks

Always unsubscribe:

```go
// Good
unsubscribe, err := client.Collection("posts").Subscribe("*", handler, nil, nil)
if err != nil {
    log.Fatal(err)
}
// ... later
unsubscribe()

// Bad - no cleanup
_, _ = client.Collection("posts").Subscribe("*", handler, nil, nil)
// Never unsubscribed - memory leak!
```

## Related Documentation

- [API Records](./API_RECORDS.md) - CRUD operations
- [Collections](./COLLECTIONS.md) - Collection configuration
- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Understanding API rules

