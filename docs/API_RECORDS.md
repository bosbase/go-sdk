# API Records - Go SDK Documentation

## Overview

The Records API provides comprehensive CRUD (Create, Read, Update, Delete) operations for collection records, along with powerful search, filtering, and authentication capabilities.

**Key Features:**
- Paginated list and search with filtering and sorting
- Single record retrieval with expand support
- Create, update, and delete operations
- Batch operations for multiple records
- Authentication methods (password, OAuth2, OTP)
- Email verification and password reset
- Relation expansion up to 6 levels deep
- Field selection and excerpt modifiers

**Backend Endpoints:**
- `GET /api/collections/{collection}/records` - List records
- `GET /api/collections/{collection}/records/{id}` - View record
- `POST /api/collections/{collection}/records` - Create record
- `PATCH /api/collections/{collection}/records/{id}` - Update record
- `DELETE /api/collections/{collection}/records/{id}` - Delete record
- `POST /api/batch` - Batch operations

## CRUD Operations

### List/Search Records

Returns a paginated records list with support for sorting, filtering, and expansion.

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
    
    // Basic list with pagination
    result, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 50,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    page, _ := result["page"].(float64)
    perPage, _ := result["perPage"].(float64)
    totalItems, _ := result["totalItems"].(float64)
    totalPages, _ := result["totalPages"].(float64)
    items, _ := result["items"].([]interface{})
    
    fmt.Printf("Page: %.0f, PerPage: %.0f, TotalItems: %.0f, TotalPages: %.0f\n",
        page, perPage, totalItems, totalPages)
    fmt.Printf("Items: %d\n", len(items))
}
```

#### Advanced List with Filtering and Sorting

```go
// Filter and sort
result, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 50,
    Filter:  `created >= "2022-01-01 00:00:00" && status = "published"`,
    Sort:    "-created,title", // DESC by created, ASC by title
    Expand:  "author,categories",
})

// Filter with operators
result2, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 50,
    Filter:  `title ~ "javascript" && views > 100`,
    Sort:    "-views",
})
```

#### Get Full List

Fetch all records at once (useful for small collections):

```go
// Get all records
allPosts, err := client.Collection("posts").GetFullList(500, &bosbase.CrudListOptions{
    Sort:   "-created",
    Filter: `status = "published"`,
})

// With batch size for large collections
allPosts, err := client.Collection("posts").GetFullList(200, &bosbase.CrudListOptions{
    Sort: "-created",
})
```

#### Get First Matching Record

Get only the first record that matches a filter:

```go
post, err := client.Collection("posts").GetFirstListItem(
    `slug = "my-post-slug"`,
    &bosbase.CrudViewOptions{
        Expand: "author,categories.tags",
    },
)
```

### View Record

Retrieve a single record by ID:

```go
// Basic retrieval
record, err := client.Collection("posts").GetOne("RECORD_ID", nil)

// With expanded relations
record, err := client.Collection("posts").GetOne("RECORD_ID", &bosbase.CrudViewOptions{
    Expand: "author,categories,tags",
})

// Nested expand
record, err := client.Collection("comments").GetOne("COMMENT_ID", &bosbase.CrudViewOptions{
    Expand: "post.author,user",
})

// Field selection
record, err := client.Collection("posts").GetOne("RECORD_ID", &bosbase.CrudViewOptions{
    Fields: "id,title,content,author.name",
})
```

### Create Record

Create a new record:

```go
// Simple create
record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":   "My First Post",
        "content": "Lorem ipsum...",
        "status":  "draft",
    },
})

// Create with relations
record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":      "My Post",
        "author":     "AUTHOR_ID",           // Single relation
        "categories": []string{"cat1", "cat2"}, // Multiple relation
    },
})

// Create with file upload
import (
    "os"
)

file, err := os.Open("image.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Post",
    },
    Files: map[string]bosbase.FileParam{
        "image": {
            Filename:    "image.jpg",
            Reader:      file,
            ContentType: "image/jpeg",
        },
    },
})

// Create with expand to get related data immediately
record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":  "My Post",
        "author": "AUTHOR_ID",
    },
    Expand: "author",
})
```

### Update Record

Update an existing record:

```go
// Simple update
record, err := client.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":  "Updated Title",
        "status": "published",
    },
})

// Update with relations
_, err = client.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "categories+": "NEW_CATEGORY_ID", // Append
        "tags-":      "OLD_TAG_ID",      // Remove
    },
})

// Update with file upload
file, err := os.Open("newimage.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

record, err := client.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Updated Title",
    },
    Files: map[string]bosbase.FileParam{
        "image": {
            Filename:    "newimage.jpg",
            Reader:      file,
            ContentType: "image/jpeg",
        },
    },
})

// Update with expand
record, err := client.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Updated",
    },
    Expand: "author,categories",
})
```

### Delete Record

Delete a record:

```go
// Simple delete
err := client.Collection("posts").Delete("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

// Note: Returns 204 No Content on success
// Returns error if record doesn't exist or permission denied
```

## Filter Syntax

The filter parameter supports a powerful query syntax:

### Comparison Operators

```go
// Equal
filter: `status = "published"`

// Not equal
filter: `status != "draft"`

// Greater than / Less than
filter: `views > 100`
filter: `created < "2023-01-01"`

// Greater/Less than or equal
filter: `age >= 18`
filter: `price <= 99.99`
```

### String Operators

```go
// Contains (like)
filter: `title ~ "javascript"`
// Equivalent to: title LIKE "%javascript%"

// Not contains
filter: `title !~ "deprecated"`

// Exact match (case-sensitive)
filter: `email = "user@example.com"`
```

### Array Operators (for multiple relations/files)

```go
// Any of / At least one
filter: `tags.id ?= "TAG_ID"`         // Any tag matches
filter: `tags.name ?~ "important"`    // Any tag name contains "important"

// All must match
filter: `tags.id = "TAG_ID" && tags.id = "TAG_ID2"`
```

### Logical Operators

```go
// AND
filter: `status = "published" && views > 100`

// OR
filter: `status = "published" || status = "featured"`

// Parentheses for grouping
filter: `(status = "published" || featured = true) && views > 50`
```

### Special Identifiers

```go
// Request context (only in API rules, not client filters)
// @request.auth.id, @request.query.*, etc.

// Collection joins
filter: `@collection.users.email = "test@example.com"`

// Record fields
filter: `author.id = @request.auth.id`
```

### Comments

```go
// Single-line comments are supported
filter: `status = "published" // Only published posts`
```

## Sorting

Sort records using the `sort` parameter:

```go
// Single field (ASC)
sort: "created"

// Single field (DESC)
sort: "-created"

// Multiple fields
sort: "-created,title" // DESC by created, then ASC by title

// Supported fields
sort: "@random"         // Random order
sort: "@rowid"          // Internal row ID
sort: "id"              // Record ID
sort: "fieldName"       // Any collection field

// Relation field sorting
sort: "author.name"     // Sort by related author's name
```

## Field Selection

Control which fields are returned:

```go
// Specific fields
fields: "id,title,content"

// All fields at level
fields: "*"

// Nested field selection
fields: "*,author.name,author.email"

// Excerpt modifier for text fields
fields: "*,content:excerpt(200,true)"
// Returns first 200 characters with ellipsis if truncated

// Combined
fields: "*,content:excerpt(200),author.name,author.email"
```

## Expanding Relations

Expand related records without additional API calls:

```go
// Single relation
expand: "author"

// Multiple relations
expand: "author,categories,tags"

// Nested relations (up to 6 levels)
expand: "author.profile,categories.tags"

// Back-relations
expand: "comments_via_post.user"
```

See [Relations Documentation](./RELATIONS.md) for detailed information.

## Pagination Options

```go
// Skip total count (faster queries)
result, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:      1,
    PerPage:   50,
    SkipTotal: true, // totalItems and totalPages will be -1
    Filter:    `status = "published"`,
})

// Get Full List with batch processing
allPosts, err := client.Collection("posts").GetFullList(200, &bosbase.CrudListOptions{
    Sort: "-created",
})
// Processes in batches of 200 to avoid memory issues
```

## Batch Operations

Execute multiple operations in a single transaction:

```go
// Create a batch
batch := client.CreateBatch()

// Add operations
batch.Collection("posts").Create(map[string]interface{}{
    "title":  "Post 1",
    "author": "AUTHOR_ID",
})

batch.Collection("posts").Create(map[string]interface{}{
    "title":  "Post 2",
    "author": "AUTHOR_ID",
})

batch.Collection("tags").Update("TAG_ID", map[string]interface{}{
    "name": "Updated Tag",
})

batch.Collection("categories").Delete("CAT_ID")

// Upsert (create or update based on id)
batch.Collection("posts").Upsert(map[string]interface{}{
    "id":    "EXISTING_ID",
    "title": "Updated Post",
})

// Send batch request
results, err := batch.Send()
if err != nil {
    log.Fatal(err)
}

// Results is an array matching the order of operations
for i, result := range results {
    if result.Status >= 400 {
        fmt.Printf("Operation %d failed: %v\n", i, result.Body)
    } else {
        fmt.Printf("Operation %d succeeded: %v\n", i, result.Body)
    }
}
```

**Note**: Batch operations must be enabled in Dashboard > Settings > Application.

## Authentication Actions

### List Auth Methods

Get available authentication methods for a collection:

```go
methods, err := client.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

password, _ := methods["password"].(map[string]interface{})
oauth2, _ := methods["oauth2"].(map[string]interface{})
otp, _ := methods["otp"].(map[string]interface{})
mfa, _ := methods["mfa"].(map[string]interface{})

fmt.Printf("Password enabled: %v\n", password["enabled"])
fmt.Printf("OAuth2 enabled: %v\n", oauth2["enabled"])
```

### Auth with Password

```go
authData, err := client.Collection("users").AuthWithPassword(
    "user@example.com", // username or email
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

// Access the returned data
token, _ := authData["token"].(string)
record, _ := authData["record"].(map[string]interface{})

// With expand
authData, err := client.Collection("users").AuthWithPassword(
    "user@example.com",
    "password123",
    "profile", // expand
    "",        // fields
    nil,       // body
    nil,       // query
    nil,       // headers
)
```

### Auth with OAuth2

```go
// Step 1: Get OAuth2 URL (usually done in UI)
methods, err := client.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

oauth2, _ := methods["oauth2"].(map[string]interface{})
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

// Step 2: After redirect, exchange code for token
authData, err := client.Collection("users").AuthWithOAuth2Code(
    "google",                    // Provider name
    "AUTHORIZATION_CODE",        // From redirect URL
    fmt.Sprint(provider["codeVerifier"]), // From step 1
    "https://yourapp.com/callback", // Redirect URL
    map[string]interface{}{      // Optional data for new accounts
        "name": "John Doe",
    },
    nil, // body
    nil, // query
    nil, // headers
    "",  // expand
    "",  // fields
)
```

### Auth with OTP (One-Time Password)

```go
// Step 1: Request OTP
otpRequest, err := client.Collection("users").RequestOTP("user@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
otpID, _ := otpRequest["otpId"].(string)

// Step 2: User enters OTP from email
// Step 3: Authenticate with OTP
authData, err := client.Collection("users").AuthWithOTP(
    otpID,
    "123456", // OTP from email
    "",       // expand
    "",       // fields
    nil,      // body
    nil,      // query
    nil,      // headers
)
```

### Auth Refresh

Refresh the current auth token and get updated user data:

```go
// Refresh auth (useful on app restart)
authData, err := client.Collection("users").AuthRefresh("", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Check if still valid
if client.AuthStore.IsValid() {
    fmt.Println("User is authenticated")
} else {
    fmt.Println("Token expired or invalid")
}
```

### Email Verification

```go
// Request verification email
err := client.Collection("users").RequestVerification("user@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Confirm verification (on verification page)
err = client.Collection("users").ConfirmVerification("VERIFICATION_TOKEN", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### Password Reset

```go
// Request password reset email
err := client.Collection("users").RequestPasswordReset("user@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Confirm password reset (on reset page)
// Note: This invalidates all previous auth tokens
err = client.Collection("users").ConfirmPasswordReset(
    "RESET_TOKEN",
    "newpassword123",
    "newpassword123", // Confirm
    nil,              // body
    nil,              // query
    nil,              // headers
)
```

### Email Change

```go
// Must be authenticated first
_, err := client.Collection("users").AuthWithPassword("user@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Request email change
err = client.Collection("users").RequestEmailChange("newemail@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Confirm email change (on confirmation page)
// Note: This invalidates all previous auth tokens
err = client.Collection("users").ConfirmEmailChange(
    "EMAIL_CHANGE_TOKEN",
    "currentpassword",
    nil, // body
    nil, // query
    nil, // headers
)
```

### Impersonate (Superuser Only)

Generate a token to authenticate as another user:

```go
// Must be authenticated as superuser
_, err := client.Collection("_superusers").AuthWithPassword("admin@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Impersonate a user
impersonateClient, err := client.Collection("users").Impersonate("USER_ID", 3600, "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Use the impersonated client
posts, err := impersonateClient.Collection("posts").GetFullList(500, nil)
if err != nil {
    log.Fatal(err)
}

// Access the token
fmt.Println("Token:", impersonateClient.AuthStore.Token())
```

## Complete Examples

### Example 1: Blog Post Search with Filters

```go
func searchPosts(client *bosbase.BosBase, query, categoryID string, minViews int) ([]interface{}, error) {
    filter := fmt.Sprintf(`title ~ "%s" || content ~ "%s"`, query, query)
    
    if categoryID != "" {
        filter += fmt.Sprintf(` && categories.id ?= "%s"`, categoryID)
    }
    
    if minViews > 0 {
        filter += fmt.Sprintf(` && views >= %d`, minViews)
    }
    
    result, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 20,
        Filter:  filter,
        Sort:    "-created",
        Expand:  "author,categories",
    })
    if err != nil {
        return nil, err
    }
    
    items, _ := result["items"].([]interface{})
    return items, nil
}
```

### Example 2: User Dashboard with Related Content

```go
func getUserDashboard(client *bosbase.BosBase, userID string) (map[string]interface{}, error) {
    // Get user's posts
    postsResult, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 10,
        Filter:  fmt.Sprintf(`author = "%s"`, userID),
        Sort:    "-created",
        Expand:  "categories",
    })
    if err != nil {
        return nil, err
    }
    
    // Get user's comments
    commentsResult, err := client.Collection("comments").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 10,
        Filter:  fmt.Sprintf(`user = "%s"`, userID),
        Sort:    "-created",
        Expand:  "post",
    })
    if err != nil {
        return nil, err
    }
    
    posts, _ := postsResult["items"].([]interface{})
    comments, _ := commentsResult["items"].([]interface{})
    
    return map[string]interface{}{
        "posts":    posts,
        "comments": comments,
    }, nil
}
```

### Example 3: Advanced Filtering

```go
result, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 50,
    Filter: `(status = "published" || featured = true) &&
              created >= "2023-01-01" &&
              (tags.id ?= "important" || categories.id = "news") &&
              views > 100 &&
              author.email != ""`,
    Sort:   "-views,created",
    Expand: "author.profile,tags,categories",
    Fields: "*,content:excerpt(300),author.name,author.email",
})
```

### Example 4: Batch Create Posts

```go
func createMultiplePosts(client *bosbase.BosBase, postsData []map[string]interface{}) ([]interface{}, error) {
    batch := client.CreateBatch()
    
    for _, postData := range postsData {
        batch.Collection("posts").Create(postData)
    }
    
    results, err := batch.Send()
    if err != nil {
        return nil, err
    }
    
    // Check for failures
    var failures []int
    for i, result := range results {
        if result.Status >= 400 {
            failures = append(failures, i)
        }
    }
    
    if len(failures) > 0 {
        fmt.Printf("Some posts failed to create: %v\n", failures)
    }
    
    var bodies []interface{}
    for _, result := range results {
        bodies = append(bodies, result.Body)
    }
    
    return bodies, nil
}
```

### Example 5: Pagination Helper

```go
func getAllRecordsPaginated(client *bosbase.BosBase, collectionName string, options *bosbase.CrudListOptions) ([]interface{}, error) {
    var allRecords []interface{}
    page := 1
    
    for {
        if options == nil {
            options = &bosbase.CrudListOptions{}
        }
        options.Page = page
        options.PerPage = 500
        options.SkipTotal = true // Skip count for performance
        
        result, err := client.Collection(collectionName).GetList(options)
        if err != nil {
            return nil, err
        }
        
        items, _ := result["items"].([]interface{})
        allRecords = append(allRecords, items...)
        
        perPage, _ := result["perPage"].(float64)
        if len(items) < int(perPage) {
            break
        }
        page++
    }
    
    return allRecords, nil
}
```

## Error Handling

```go
record, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Post",
    },
})
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        switch clientErr.Status {
        case 400:
            fmt.Println("Validation errors:", clientErr.Response)
        case 403:
            fmt.Println("Access denied")
        case 404:
            fmt.Println("Collection or record not found")
        default:
            fmt.Printf("Unexpected error: %v\n", err)
        }
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Best Practices

1. **Use Pagination**: Always use pagination for large datasets
2. **Skip Total When Possible**: Use `SkipTotal: true` for better performance when you don't need counts
3. **Batch Operations**: Use batch for multiple operations to reduce round trips
4. **Field Selection**: Only request fields you need to reduce payload size
5. **Expand Wisely**: Only expand relations you actually use
6. **Filter Before Sort**: Apply filters before sorting for better performance
7. **Cache Auth Tokens**: Auth tokens are automatically stored in `AuthStore`, no need to manually cache
8. **Handle Errors**: Always handle authentication and permission errors gracefully

## Related Documentation

- [Collections](./COLLECTIONS.md) - Collection configuration
- [Relations](./RELATIONS.md) - Working with relations
- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Filter syntax details
- [Authentication](./AUTHENTICATION.md) - Detailed authentication guide
- [Files](./FILES.md) - File uploads and handling

