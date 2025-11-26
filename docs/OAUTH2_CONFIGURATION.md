# OAuth2 Configuration Guide - Go SDK Documentation

This guide explains how to configure OAuth2 authentication providers for auth collections using the BosBase Go SDK.

## Overview

OAuth2 allows users to authenticate with your application using third-party providers like Google, GitHub, Facebook, etc. Before you can use OAuth2 authentication, you need to:

1. **Create an OAuth2 app** in the provider's dashboard
2. **Obtain Client ID and Client Secret** from the provider
3. **Register a redirect URL** (typically: `https://yourdomain.com/api/oauth2-redirect`)
4. **Configure the provider** in your BosBase auth collection using the SDK

## Prerequisites

- An auth collection in your BosBase instance
- OAuth2 app credentials (Client ID and Client Secret) from your chosen provider
- Admin/superuser authentication to configure collections

## Supported Providers

The following OAuth2 providers are supported:

- **google** - Google OAuth2
- **github** - GitHub OAuth2
- **gitlab** - GitLab OAuth2
- **discord** - Discord OAuth2
- **facebook** - Facebook OAuth2
- **microsoft** - Microsoft OAuth2
- **apple** - Apple Sign In
- **twitter** - Twitter OAuth2
- **spotify** - Spotify OAuth2
- **kakao** - Kakao OAuth2
- **twitch** - Twitch OAuth2
- **strava** - Strava OAuth2
- **vk** - VK OAuth2
- **yandex** - Yandex OAuth2
- **patreon** - Patreon OAuth2
- **linkedin** - LinkedIn OAuth2
- **instagram** - Instagram OAuth2
- **vimeo** - Vimeo OAuth2
- **digitalocean** - DigitalOcean OAuth2
- **bitbucket** - Bitbucket OAuth2
- **dropbox** - Dropbox OAuth2
- **planningcenter** - Planning Center OAuth2
- **notion** - Notion OAuth2
- **linear** - Linear OAuth2
- **oidc**, **oidc2**, **oidc3** - OpenID Connect (OIDC) providers

## Basic Usage

### 1. Enable OAuth2 for a Collection

First, enable OAuth2 authentication for your auth collection:

```go
package main

import (
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("https://your-instance.com")
    defer client.Close()
    
    // Authenticate as admin
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get collection
    collection, err := client.Collections.GetOne("users", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Enable OAuth2
    oauth2Config := map[string]interface{}{
        "enabled": true,
    }
    
    // Update collection with OAuth2 enabled
    collection["oauth2"] = oauth2Config
    _, err = client.Collections.Update("users", &bosbase.CrudMutateOptions{
        Body: collection,
    })
}
```

### 2. Add an OAuth2 Provider

Add a provider configuration to your collection. You'll need the URLs and credentials from your OAuth2 app:

```go
// Get collection
collection, err := client.Collections.GetOne("users", nil)
if err != nil {
    log.Fatal(err)
}

// Configure OAuth2 providers
oauth2, _ := collection["oauth2"].(map[string]interface{})
providers, _ := oauth2["providers"].([]interface{})

// Add Google OAuth2 provider
newProvider := map[string]interface{}{
    "name":        "google",
    "clientId":    "your-google-client-id",
    "clientSecret": "your-google-client-secret",
    "authURL":     "https://accounts.google.com/o/oauth2/v2/auth",
    "tokenURL":    "https://oauth2.googleapis.com/token",
    "userInfoURL": "https://www.googleapis.com/oauth2/v2/userinfo",
    "displayName": "Google",
    "pkce":        true, // Optional: enable PKCE if supported
}

providers = append(providers, newProvider)
oauth2["providers"] = providers
collection["oauth2"] = oauth2

_, err = client.Collections.Update("users", &bosbase.CrudMutateOptions{
    Body: collection,
})
```

### 3. Authenticate with OAuth2

```go
// List available OAuth2 providers
methods, err := client.Collection("users").ListAuthMethods("oauth2", nil, nil)
if err != nil {
    log.Fatal(err)
}

// Authenticate with OAuth2
// Note: OAuth2 flow typically requires browser redirect
// This is a simplified example
authData, err := client.Collection("users").AuthWithOAuth2("google", nil, nil, nil, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Authenticated: %v\n", authData)
```

## Related Documentation

- [Authentication](./AUTHENTICATION.md) - Authentication methods
- [Collections](./COLLECTIONS.md) - Collection configuration

