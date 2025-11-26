# Custom Token Binding and Login - Go SDK Documentation

The Go SDK and BosBase service support binding a custom token to an auth record (both `users` and `_superusers`) and signing in with that token. The server stores bindings in the `_token_bindings` table (created automatically on first bind; legacy `_tokenBindings`/`tokenBindings` are auto-renamed). Tokens are stored as hashes so raw values aren't persisted.

## API endpoints
- `POST /api/collections/{collection}/bind-token`
- `POST /api/collections/{collection}/unbind-token`
- `POST /api/collections/{collection}/auth-with-token`

## Binding a token

```go
package main

import (
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://127.0.0.1:8090")
    defer client.Close()
    
    // Bind for a regular user
    _, err := client.Collection("users").BindCustomToken(
        "user@example.com",
        "user-password",
        "my-app-token",
        nil, nil, nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Bind for a superuser
    _, err = client.Collection("_superusers").BindCustomToken(
        "admin@example.com",
        "admin-password",
        "admin-app-token",
        nil, nil, nil,
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

## Unbinding a token

```go
// Stop accepting the token for the user
_, err := client.Collection("users").UnbindCustomToken(
    "user@example.com",
    "user-password",
    "my-app-token",
    nil, nil, nil,
)

// Stop accepting the token for a superuser
_, err = client.Collection("_superusers").UnbindCustomToken(
    "admin@example.com",
    "admin-password",
    "admin-app-token",
    nil, nil, nil,
)
```

## Logging in with a token

```go
// Login with the previously bound token
auth, err := client.Collection("users").AuthWithToken(
    "my-app-token",
    nil, nil, nil,
)
if err != nil {
    log.Fatal(err)
}

// Auth data is automatically stored
fmt.Printf("Token: %s\n", client.AuthStore.Token())
record := client.AuthStore.Record()
fmt.Printf("Record: %v\n", record)

// Superuser token login
superAuth, err := client.Collection("_superusers").AuthWithToken(
    "admin-app-token",
    nil, nil, nil,
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Superuser token: %s\n", client.AuthStore.Token())
```

## Notes

- Binding and unbinding require a valid email and password for the target account.
- The same token value can be used for either `users` or `_superusers` collections; the collection is enforced during login.
- MFA and existing auth rules still apply when authenticating with a token.

## Related Documentation

- [Authentication](./AUTHENTICATION.md) - Authentication methods
- [API Records](./API_RECORDS.md) - Record operations

