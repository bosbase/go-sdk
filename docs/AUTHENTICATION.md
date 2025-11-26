# Authentication - Go SDK Documentation

## Overview

Authentication in BosBase is stateless and token-based. A client is considered authenticated as long as it sends a valid `Authorization: YOUR_AUTH_TOKEN` header with requests.

**Key Points:**
- **No sessions**: BosBase APIs are fully stateless (tokens are not stored in the database)
- **No logout endpoint**: To "logout", simply clear the token from your local state (`client.AuthStore.Clear()`)
- **Token generation**: Auth tokens are generated through auth collection Web APIs or programmatically
- **Admin users**: `_superusers` collection works like regular auth collections but with full access (API rules are ignored)
- **OAuth2 limitation**: OAuth2 is not supported for `_superusers` collection

## Authentication Methods

BosBase supports multiple authentication methods that can be configured individually for each auth collection:

1. **Password Authentication** - Email/username + password
2. **OTP Authentication** - One-time password via email
3. **OAuth2 Authentication** - Google, GitHub, Microsoft, etc.
4. **Multi-factor Authentication (MFA)** - Requires 2 different auth methods

## Authentication Store

The SDK maintains an `AuthStore` that automatically manages the authentication state:

```go
package main

import (
    "fmt"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    // Check authentication status
    fmt.Printf("Is valid: %v\n", client.AuthStore.IsValid())
    fmt.Printf("Token: %s\n", client.AuthStore.Token())
    fmt.Printf("Record: %v\n", client.AuthStore.Record())
    
    // Clear authentication (logout)
    client.AuthStore.Clear()
}
```

## Password Authentication

Authenticate using email/username and password. The identity field can be configured in the collection options (default is email).

**Backend Endpoint:** `POST /api/collections/{collection}/auth-with-password`

### Basic Usage

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
    
    // Authenticate with email and password
    authData, err := client.Collection("users").AuthWithPassword(
        "test@example.com",
        "password123",
        "", // expand
        "", // fields
        nil, // body
        nil, // query
        nil, // headers
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Auth data is automatically stored in client.AuthStore
    fmt.Printf("Is valid: %v\n", client.AuthStore.IsValid())
    fmt.Printf("Token: %s\n", client.AuthStore.Token())
    
    record, _ := authData["record"].(map[string]interface{})
    recordID, _ := record["id"].(string)
    fmt.Printf("User record ID: %s\n", recordID)
}
```

### Response Format

```go
{
    "token": "eyJhbGciOiJIUzI1NiJ9...",
    "record": {
        "id": "record_id",
        "email": "test@example.com",
        // ... other user fields
    }
}
```

### Error Handling with MFA

```go
authData, err := client.Collection("users").AuthWithPassword(
    "test@example.com", "pass123", "", "", nil, nil, nil)
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        if resp, ok := clientErr.Response.(map[string]interface{}); ok {
            if mfaID, ok := resp["mfaId"].(string); ok {
                // Handle MFA flow (see Multi-factor Authentication section)
                fmt.Printf("MFA required: %s\n", mfaID)
            }
        }
    } else {
        log.Printf("Authentication failed: %v\n", err)
    }
}
```

## OTP Authentication

One-time password authentication via email.

**Backend Endpoints:**
- `POST /api/collections/{collection}/request-otp` - Request OTP
- `POST /api/collections/{collection}/auth-with-otp` - Authenticate with OTP

### Request OTP

```go
// Send OTP to user's email
result, err := client.Collection("users").RequestOTP("test@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
otpID, _ := result["otpId"].(string)
fmt.Printf("OTP ID: %s\n", otpID)
```

### Authenticate with OTP

```go
// Step 1: Request OTP
result, err := client.Collection("users").RequestOTP("test@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
otpID, _ := result["otpId"].(string)

// Step 2: User enters OTP from email
// Step 3: Authenticate with OTP
authData, err := client.Collection("users").AuthWithOTP(
    otpID,
    "123456", // OTP code from email
    "",       // expand
    "",       // fields
    nil,      // body
    nil,      // query
    nil,      // headers
)
if err != nil {
    log.Fatal(err)
}
```

## OAuth2 Authentication

**Backend Endpoint:** `POST /api/collections/{collection}/auth-with-oauth2`

### Manual Code Exchange

```go
// Get auth methods
authMethods, err := client.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

oauth2, _ := authMethods["oauth2"].(map[string]interface{})
providers, _ := oauth2["providers"].([]interface{})

var provider map[string]interface{}
for _, p := range providers {
    if pMap, ok := p.(map[string]interface{}); ok {
        if name, _ := pMap["name"].(string); name == "google" {
            provider = pMap
            break
        }
    }
}

// Exchange code for token (after OAuth2 redirect)
authData, err := client.Collection("users").AuthWithOAuth2Code(
    provider["name"].(string),
    code,
    fmt.Sprint(provider["codeVerifier"]),
    redirectURL,
    nil, // createData
    nil, // body
    nil, // query
    nil, // headers
    "",  // expand
    "",  // fields
)
```

## Multi-Factor Authentication (MFA)

Requires 2 different auth methods.

```go
var mfaID string

// First auth method (password)
_, err := client.Collection("users").AuthWithPassword(
    "test@example.com", "pass123", "", "", nil, nil, nil)
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        if resp, ok := clientErr.Response.(map[string]interface{}); ok {
            if id, ok := resp["mfaId"].(string); ok {
                mfaID = id
                
                // Second auth method (OTP)
                otpResult, err := client.Collection("users").RequestOTP("test@example.com", nil, nil, nil)
                if err != nil {
                    log.Fatal(err)
                }
                otpID, _ := otpResult["otpId"].(string)
                
                authData, err := client.Collection("users").AuthWithOTP(
                    otpID,
                    "123456",
                    "", // expand
                    "", // fields
                    map[string]interface{}{
                        "mfaId": mfaID,
                    }, // body
                    nil, // query
                    nil, // headers
                )
                if err != nil {
                    log.Fatal(err)
                }
            }
        }
    }
}
```

## User Impersonation

Superusers can impersonate other users.

**Backend Endpoint:** `POST /api/collections/{collection}/impersonate/{id}`

```go
// Authenticate as superuser
_, err := client.Collection("_superusers").AuthWithPassword(
    "admin@example.com", "adminpass", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Impersonate a user
impersonateClient, err := client.Collection("users").Impersonate(
    "USER_RECORD_ID",
    3600, // Optional: token duration in seconds
    "",   // expand
    "",   // fields
    nil,  // body
    nil,  // query
    nil,  // headers
)
if err != nil {
    log.Fatal(err)
}

// Use impersonate client
data, err := impersonateClient.Collection("posts").GetFullList(500, nil)
if err != nil {
    log.Fatal(err)
}
```

## Auth Token Verification

Verify token by calling `AuthRefresh()`.

**Backend Endpoint:** `POST /api/collections/{collection}/auth-refresh`

```go
authData, err := client.Collection("users").AuthRefresh("", "", nil, nil, nil)
if err != nil {
    log.Printf("Token verification failed: %v\n", err)
    client.AuthStore.Clear()
} else {
    fmt.Println("Token is valid")
}
```

## List Available Auth Methods

**Backend Endpoint:** `GET /api/collections/{collection}/auth-methods`

```go
authMethods, err := client.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

password, _ := authMethods["password"].(map[string]interface{})
oauth2, _ := authMethods["oauth2"].(map[string]interface{})
mfa, _ := authMethods["mfa"].(map[string]interface{})

fmt.Printf("Password enabled: %v\n", password["enabled"])
fmt.Printf("OAuth2 enabled: %v\n", oauth2["enabled"])
fmt.Printf("MFA enabled: %v\n", mfa["enabled"])
```

## Complete Examples

### Example 1: Complete Authentication Flow with Error Handling

```go
package main

import (
    "fmt"
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func authenticateUser(client *bosbase.BosBase, email, password string) (map[string]interface{}, error) {
    // Try password authentication
    authData, err := client.Collection("users").AuthWithPassword(
        email, password, "", "", nil, nil, nil)
    if err != nil {
        // Check if MFA is required
        if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
            if resp, ok := clientErr.Response.(map[string]interface{}); ok {
                if mfaID, ok := resp["mfaId"].(string); ok {
                    fmt.Println("MFA required, proceeding with second factor...")
                    return handleMFA(client, email, mfaID)
                }
            }
        }
        
        // Handle other errors
        if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
            switch clientErr.Status {
            case 400:
                return nil, fmt.Errorf("invalid credentials")
            case 403:
                return nil, fmt.Errorf("password authentication is not enabled for this collection")
            default:
                return nil, err
            }
        }
        return nil, err
    }
    
    fmt.Printf("Successfully authenticated: %v\n", email)
    return authData, nil
}

func handleMFA(client *bosbase.BosBase, email, mfaID string) (map[string]interface{}, error) {
    // Request OTP for second factor
    otpResult, err := client.Collection("users").RequestOTP(email, nil, nil, nil)
    if err != nil {
        return nil, err
    }
    
    otpID, _ := otpResult["otpId"].(string)
    
    // In a real app, show a modal/form for the user to enter OTP
    // For this example, we'll simulate getting the OTP
    userEnteredOTP := getUserOTPInput() // Your UI function
    
    // Authenticate with OTP and MFA ID
    authData, err := client.Collection("users").AuthWithOTP(
        otpID,
        userEnteredOTP,
        "", // expand
        "", // fields
        map[string]interface{}{
            "mfaId": mfaID,
        }, // body
        nil, // query
        nil, // headers
    )
    if err != nil {
        if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
            if clientErr.Status == 429 {
                return nil, fmt.Errorf("too many OTP attempts, please request a new OTP")
            }
        }
        return nil, fmt.Errorf("invalid OTP code")
    }
    
    fmt.Println("MFA authentication successful")
    return authData, nil
}

func getUserOTPInput() string {
    // In a real app, this would get input from the user
    return "123456"
}

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    authData, err := authenticateUser(client, "user@example.com", "password123")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("User is authenticated: %v\n", client.AuthStore.Record())
}
```

### Example 2: Token Management and Refresh

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    bosbase "github.com/bosbase/go-sdk"
)

// Check if user is already authenticated
func checkAuth(client *bosbase.BosBase) (bool, error) {
    if client.AuthStore.IsValid() {
        fmt.Printf("User is authenticated: %v\n", client.AuthStore.Record())
        
        // Verify token is still valid and refresh if needed
        _, err := client.Collection("users").AuthRefresh("", "", nil, nil, nil)
        if err != nil {
            fmt.Println("Token expired or invalid, clearing auth")
            client.AuthStore.Clear()
            return false, nil
        }
        
        fmt.Println("Token refreshed successfully")
        return true, nil
    }
    return false, nil
}

// Auto-refresh token before expiration
func setupAutoRefresh(client *bosbase.BosBase) {
    if !client.AuthStore.IsValid() {
        return
    }
    
    // In a real app, you would parse the JWT token to get expiration
    // For this example, we'll refresh every 5 minutes
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for range ticker.C {
            _, err := client.Collection("users").AuthRefresh("", "", nil, nil, nil)
            if err != nil {
                fmt.Printf("Auto-refresh failed: %v\n", err)
                client.AuthStore.Clear()
                ticker.Stop()
            } else {
                fmt.Println("Token auto-refreshed")
            }
        }
    }()
}

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    isAuthenticated, err := checkAuth(client)
    if err != nil {
        log.Fatal(err)
    }
    
    if !isAuthenticated {
        fmt.Println("Not authenticated, redirect to login")
    } else {
        setupAutoRefresh(client)
        // Keep the program running
        select {}
    }
}
```

### Example 3: Admin Impersonation for Support

```go
package main

import (
    "fmt"
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func impersonateUserForSupport(client *bosbase.BosBase, userID string) (map[string]interface{}, error) {
    // Authenticate as admin
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "adminpassword", "", "", nil, nil, nil)
    if err != nil {
        return nil, err
    }
    
    // Impersonate the user (1 hour token)
    userClient, err := client.Collection("users").Impersonate(userID, 3600, "", "", nil, nil, nil)
    if err != nil {
        return nil, err
    }
    
    record := userClient.AuthStore.Record()
    email, _ := record["email"].(string)
    fmt.Printf("Impersonating user: %s\n", email)
    
    // Use the impersonated client to test user experience
    userRecords, err := userClient.Collection("posts").GetFullList(500, nil)
    if err != nil {
        return nil, err
    }
    fmt.Printf("User can see %d posts\n", len(userRecords))
    
    // Check what the user sees
    userView, err := userClient.Collection("posts").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 10,
        Filter:  `published = true`,
    })
    if err != nil {
        return nil, err
    }
    
    items, _ := userView["items"].([]interface{})
    
    return map[string]interface{}{
        "canAccess":  len(items),
        "totalPosts": len(userRecords),
    }, nil
}

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    result, err := impersonateUserForSupport(client, "user_record_id")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("User access check: %v\n", result)
}
```

## Best Practices

1. **Secure Token Storage**: Never expose tokens in client-side code or logs
2. **Token Refresh**: Implement automatic token refresh before expiration
3. **Error Handling**: Always handle MFA requirements and token expiration
4. **OAuth2 Security**: Always validate the `state` parameter in OAuth2 callbacks
5. **API Keys**: Use impersonation tokens for server-to-server communication only
6. **Superuser Tokens**: Never expose superuser impersonation tokens in client code
7. **OTP Security**: Use OTP with MFA for security-critical applications
8. **Rate Limiting**: Be aware of rate limits on authentication endpoints

## Troubleshooting

### Token Expired

If you get 401 errors, check if the token has expired:

```go
_, err := client.Collection("users").AuthRefresh("", "", nil, nil, nil)
if err != nil {
    // Token expired, require re-authentication
    client.AuthStore.Clear()
    // Redirect to login
}
```

### MFA Required

If authentication returns 401 with mfaId:

```go
if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
    if resp, ok := clientErr.Response.(map[string]interface{}); ok {
        if mfaID, ok := resp["mfaId"].(string); ok {
            // Proceed with second authentication factor
        }
    }
}
```

## Related Documentation

- [Collections](./COLLECTIONS.md)
- [API Rules](./API_RULES_AND_FILTERS.md)

