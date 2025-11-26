# Built-in Users Collection Guide - Go SDK Documentation

This guide explains how to use the built-in `users` collection for authentication, registration, and API rules. **The `users` collection is automatically created when BosBase is initialized and does not need to be created manually.**

## Table of Contents

1. [Overview](#overview)
2. [Users Collection Structure](#users-collection-structure)
3. [User Registration](#user-registration)
4. [User Login/Authentication](#user-loginauthentication)
5. [API Rules and Filters with Users](#api-rules-and-filters-with-users)
6. [Using Users with Other Collections](#using-users-with-other-collections)
7. [Complete Examples](#complete-examples)

## Overview

The `users` collection is a **built-in auth collection** that is automatically created when BosBase starts. It has:

- **Collection ID**: `_pb_users_auth_`
- **Collection Name**: `users`
- **Type**: `auth` (authentication collection)
- **Purpose**: User accounts, authentication, and authorization

**Important**: 
- ✅ **DO NOT** create a new `users` collection manually
- ✅ **DO** use the existing built-in `users` collection
- ✅ The collection already has proper API rules configured
- ✅ It supports password, OAuth2, and OTP authentication

### Getting Users Collection Information

```go
// Get the users collection details
usersCollection, err := client.Collections.GetOne("users", nil)
if err != nil {
    log.Fatal(err)
}

id, _ := usersCollection["id"].(string)
name, _ := usersCollection["name"].(string)
collectionType, _ := usersCollection["type"].(string)

fmt.Printf("Collection ID: %s\n", id)
fmt.Printf("Collection Name: %s\n", name)
fmt.Printf("Collection Type: %s\n", collectionType)
```

## Users Collection Structure

### System Fields (Automatically Created)

These fields are automatically added to all auth collections (including `users`):

| Field Name | Type | Description | Required | Hidden |
|------------|------|-------------|----------|--------|
| `id` | text | Unique record identifier | Yes | No |
| `email` | email | User email address | Yes* | No |
| `username` | text | Username (optional, if enabled) | No* | No |
| `password` | password | Hashed password | Yes* | Yes |
| `tokenKey` | text | Token key for auth tokens | Yes | Yes |
| `emailVisibility` | bool | Whether email is visible to others | No | No |
| `verified` | bool | Whether email is verified | No | No |
| `created` | date | Record creation timestamp | Yes | No |
| `updated` | date | Last update timestamp | Yes | No |

*Required based on authentication method configuration (password auth, username auth, etc.)

### Default API Rules

The `users` collection comes with these default API rules:

```go
{
    "listRule":  "id = @request.auth.id",    // Users can only list themselves
    "viewRule":  "id = @request.auth.id",   // Users can only view themselves
    "createRule": "",                        // Anyone can register (public)
    "updateRule": "id = @request.auth.id", // Users can only update themselves
    "deleteRule": "id = @request.auth.id"  // Users can only delete themselves
}
```

## User Registration

### Create User Account

```go
// Create a new user account
user, err := client.Collection("users").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "email":          "user@example.com",
        "password":        "password123",
        "passwordConfirm": "password123",
        "name":           "John Doe",
    },
})
if err != nil {
    log.Fatal(err)
}

userID, _ := user["id"].(string)
fmt.Printf("Created user: %s\n", userID)
```

## User Login/Authentication

### Password Authentication

```go
// Authenticate with email and password
authData, err := client.Collection("users").AuthWithPassword(
    "user@example.com",
    "password123",
    "", "", nil, nil, nil,
)
if err != nil {
    log.Fatal(err)
}

// Auth data is automatically stored
fmt.Printf("Authenticated: %v\n", client.AuthStore.IsValid())
fmt.Printf("Token: %s\n", client.AuthStore.Token())
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

## Using Users with Other Collections

### Reference Users in Other Collections

```go
// Create a post with author reference
authRecord := client.AuthStore.Record()
userID, _ := authRecord["id"].(string)

post, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":  "My Post",
        "author": userID, // Reference to users collection
    },
})
```

## Complete Examples

### Example: User Registration and Login Flow

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
    
    // Register new user
    user, err := client.Collection("users").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "email":          "newuser@example.com",
            "password":        "securepassword123",
            "passwordConfirm": "securepassword123",
            "name":           "New User",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    userID, _ := user["id"].(string)
    fmt.Printf("Registered user: %s\n", userID)
    
    // Login
    _, err = client.Collection("users").AuthWithPassword(
        "newuser@example.com",
        "securepassword123",
        "", "", nil, nil, nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Successfully authenticated!")
}
```

## Related Documentation

- [Authentication](./AUTHENTICATION.md) - Complete authentication guide
- [Collections](./COLLECTIONS.md) - Collection configuration
- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Access control

