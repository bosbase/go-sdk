# Authentication - Go SDK Documentation

## Overview

Authentication in BosBase is stateless and token-based. A client is considered authenticated as long as it sends a valid `Authorization: YOUR_AUTH_TOKEN` header with requests.

**Key Points:**
- **No sessions**: BosBase APIs are fully stateless (tokens are not stored in the database)
- **No logout endpoint**: To "logout", simply clear the token from your local state (`pb.AuthStore.Clear()`)
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
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    // Check authentication status
    fmt.Println("Is Valid:", pb.AuthStore.IsValid())
    fmt.Println("Token:", pb.AuthStore.Token())
    fmt.Println("Record:", pb.AuthStore.Record())
    
    // Clear authentication (logout)
    pb.AuthStore.Clear()
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
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    // Authenticate with email and password
    authData, err := pb.Collection("users").AuthWithPassword(
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
    
    // Auth data is automatically stored in pb.AuthStore
    fmt.Println("Is Valid:", pb.AuthStore.IsValid())
    fmt.Println("Token:", pb.AuthStore.Token())
    if record := pb.AuthStore.Record(); record != nil {
        fmt.Println("User ID:", record["id"])
    }
}
```

### Response Format

```go
// authData is a map[string]interface{} containing:
// {
//   "token": "eyJhbGciOiJIUzI1NiJ9...",
//   "record": {
//     "id": "record_id",
//     "email": "test@example.com",
//     // ... other user fields
//   }
// }
```

### Error Handling with MFA

```go
authData, err := pb.Collection("users").AuthWithPassword(
    "test@example.com",
    "pass123",
    "", "", nil, nil, nil,
)
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        if resp, ok := clientErr.Response.(map[string]interface{}); ok {
            if mfaID, ok := resp["mfaId"].(string); ok {
                // Handle MFA flow (see Multi-factor Authentication section)
                handleMFA(mfaID)
            }
        }
    }
    log.Fatal(err)
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
result, err := pb.Collection("users").RequestOTP(
    "test@example.com",
    nil, // body
    nil, // query
    nil, // headers
)
if err != nil {
    log.Fatal(err)
}
otpID := result["otpId"].(string)
fmt.Println("OTP ID:", otpID)
```

### Authenticate with OTP

```go
// Step 1: Request OTP
result, err := pb.Collection("users").RequestOTP("test@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
otpID := result["otpId"].(string)

// Step 2: User enters OTP from email
otpCode := "123456" // OTP code from email

// Step 3: Authenticate with OTP
authData, err := pb.Collection("users").AuthWithOTP(
    otpID,
    otpCode,
    "", // expand
    "", // fields
    nil, // body
    nil, // query
    nil, // headers
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
methods, err := pb.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

// Find provider
oauth2, _ := methods["oauth2"].(map[string]interface{})
providers, _ := oauth2["providers"].([]interface{})
var provider map[string]interface{}
for _, p := range providers {
    if pm, ok := p.(map[string]interface{}); ok {
        if name, _ := pm["name"].(string); name == "google" {
            provider = pm
            break
        }
    }
}

// Exchange code for token (after OAuth2 redirect)
code := "AUTHORIZATION_CODE"
codeVerifier := provider["codeVerifier"].(string)
redirectURL := "https://yourapp.com/callback"

authData, err := pb.Collection("users").AuthWithOAuth2Code(
    "google",
    code,
    codeVerifier,
    redirectURL,
    nil, // createData
    nil, // body
    nil, // query
    nil, // headers
    "",  // expand
    "",  // fields
)
if err != nil {
    log.Fatal(err)
}
```

## Multi-Factor Authentication (MFA)

Requires 2 different auth methods.

```go
var mfaID string

// First auth method (password)
_, err := pb.Collection("users").AuthWithPassword(
    "test@example.com",
    "pass123",
    "", "", nil, nil, nil,
)
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        if resp, ok := clientErr.Response.(map[string]interface{}); ok {
            if id, ok := resp["mfaId"].(string); ok {
                mfaID = id
                
                // Second auth method (OTP)
                otpResult, err := pb.Collection("users").RequestOTP("test@example.com", nil, nil, nil)
                if err != nil {
                    log.Fatal(err)
                }
                otpID := otpResult["otpId"].(string)
                
                // Authenticate with OTP and MFA ID
                body := map[string]interface{}{"mfaId": mfaID}
                _, err = pb.Collection("users").AuthWithOTP(
                    otpID,
                    "123456",
                    "", "", body, nil, nil,
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
_, err := pb.Admins.AuthWithPassword("admin@example.com", "adminpass", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Impersonate a user
impersonateClient, err := pb.Collection("users").Impersonate(
    "USER_RECORD_ID",
    3600, // token duration in seconds
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
data, err := impersonateClient.Collection("posts").GetFullList(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Auth Token Verification

Verify token by calling `AuthRefresh()`.

**Backend Endpoint:** `POST /api/collections/{collection}/auth-refresh`

```go
authData, err := pb.Collection("users").AuthRefresh("", "", nil, nil, nil)
if err != nil {
    fmt.Println("Token verification failed:", err)
    pb.AuthStore.Clear()
} else {
    fmt.Println("Token is valid")
}
```

## List Available Auth Methods

**Backend Endpoint:** `GET /api/collections/{collection}/auth-methods`

```go
methods, err := pb.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

password, _ := methods["password"].(map[string]interface{})
oauth2, _ := methods["oauth2"].(map[string]interface{})
mfa, _ := methods["mfa"].(map[string]interface{})

fmt.Println("Password enabled:", password["enabled"])
fmt.Println("OAuth2 enabled:", oauth2["enabled"])
fmt.Println("MFA enabled:", mfa["enabled"])
```

## Complete Examples

### Example 1: Complete Authentication Flow with Error Handling

```go
package main

import (
    "fmt"
    "log"
    "github.com/your-org/bosbase-go-sdk"
)

func authenticateUser(pb *bosbase.BosBase, email, password string) (map[string]interface{}, error) {
    // Try password authentication
    authData, err := pb.Collection("users").AuthWithPassword(
        email, password, "", "", nil, nil, nil,
    )
    if err != nil {
        // Check if MFA is required
        if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
            if clientErr.Status == 401 {
                if resp, ok := clientErr.Response.(map[string]interface{}); ok {
                    if mfaID, ok := resp["mfaId"].(string); ok {
                        fmt.Println("MFA required, proceeding with second factor...")
                        return handleMFA(pb, email, mfaID)
                    }
                }
            }
        }
        return nil, err
    }
    
    fmt.Println("Successfully authenticated:", email)
    return authData, nil
}

func handleMFA(pb *bosbase.BosBase, email, mfaID string) (map[string]interface{}, error) {
    // Request OTP for second factor
    otpResult, err := pb.Collection("users").RequestOTP(email, nil, nil, nil)
    if err != nil {
        return nil, err
    }
    otpID := otpResult["otpId"].(string)
    
    // In a real app, get OTP from user input
    userEnteredOTP := getUserOTPInput() // Your function to get user input
    
    // Authenticate with OTP and MFA ID
    body := map[string]interface{}{"mfaId": mfaID}
    authData, err := pb.Collection("users").AuthWithOTP(
        otpID,
        userEnteredOTP,
        "", "", body, nil, nil,
    )
    if err != nil {
        return nil, fmt.Errorf("MFA authentication failed: %v", err)
    }
    
    fmt.Println("MFA authentication successful")
    return authData, nil
}

func getUserOTPInput() string {
    // In a real app, this would get input from user
    return "123456"
}

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    authData, err := authenticateUser(pb, "user@example.com", "password123")
    if err != nil {
        log.Fatal("Authentication failed:", err)
    }
    
    fmt.Println("User is authenticated:", pb.AuthStore.Record())
}
```

### Example 2: Token Management and Refresh

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/your-org/bosbase-go-sdk"
)

func checkAuth(pb *bosbase.BosBase) bool {
    if pb.AuthStore.IsValid() {
        record := pb.AuthStore.Record()
        if record != nil {
            fmt.Println("User is authenticated:", record["email"])
        }
        
        // Verify token is still valid and refresh if needed
        _, err := pb.Collection("users").AuthRefresh("", "", nil, nil, nil)
        if err != nil {
            fmt.Println("Token expired or invalid, clearing auth")
            pb.AuthStore.Clear()
            return false
        }
        fmt.Println("Token refreshed successfully")
        return true
    }
    return false
}

func setupAutoRefresh(pb *bosbase.BosBase) {
    if !pb.AuthStore.IsValid() {
        return
    }
    
    // In a real app, you would parse the JWT token to get expiration time
    // For now, we'll just refresh periodically
    ticker := time.NewTicker(30 * time.Minute)
    go func() {
        for range ticker.C {
            if pb.AuthStore.IsValid() {
                _, err := pb.Collection("users").AuthRefresh("", "", nil, nil, nil)
                if err != nil {
                    fmt.Println("Auto-refresh failed:", err)
                    pb.AuthStore.Clear()
                    ticker.Stop()
                } else {
                    fmt.Println("Token auto-refreshed")
                }
            } else {
                ticker.Stop()
            }
        }
    }()
}

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    if checkAuth(pb) {
        setupAutoRefresh(pb)
    }
}
```

### Example 3: Admin Impersonation for Support

```go
package main

import (
    "fmt"
    "log"
    "github.com/your-org/bosbase-go-sdk"
)

func impersonateUserForSupport(pb *bosbase.BosBase, userID string) error {
    // Authenticate as admin
    _, err := pb.Admins.AuthWithPassword("admin@example.com", "adminpassword", "", "", nil, nil, nil)
    if err != nil {
        return err
    }
    
    // Impersonate the user (1 hour token)
    userClient, err := pb.Collection("users").Impersonate(userID, 3600, "", "", nil, nil, nil)
    if err != nil {
        return err
    }
    
    record := userClient.AuthStore.Record()
    if record != nil {
        fmt.Println("Impersonating user:", record["email"])
    }
    
    // Use the impersonated client to test user experience
    userRecords, err := userClient.Collection("posts").GetFullList(nil, nil, nil)
    if err != nil {
        return err
    }
    
    fmt.Printf("User can see %d posts\n", len(userRecords))
    
    return nil
}

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    err := impersonateUserForSupport(pb, "user_record_id")
    if err != nil {
        log.Fatal("Impersonation failed:", err)
    }
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
_, err := pb.Collection("users").AuthRefresh("", "", nil, nil, nil)
if err != nil {
    // Token expired, require re-authentication
    pb.AuthStore.Clear()
    // Redirect to login
}
```

### MFA Required
If authentication returns 401 with mfaId:
```go
if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
    if clientErr.Status == 401 {
        if resp, ok := clientErr.Response.(map[string]interface{}); ok {
            if mfaID, ok := resp["mfaId"].(string); ok {
                // Proceed with second authentication factor
            }
        }
    }
}
```

