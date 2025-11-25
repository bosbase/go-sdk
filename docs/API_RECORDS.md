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
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://127.0.0.1:8090")
    
    // Basic list with pagination
    result, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 50,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Page:", result["page"])
    fmt.Println("Per Page:", result["perPage"])
    fmt.Println("Total Items:", result["totalItems"])
    fmt.Println("Total Pages:", result["totalPages"])
    if items, ok := result["items"].([]interface{}); ok {
        fmt.Println("Items:", len(items))
    }
}
```

#### Advanced List with Filtering and Sorting

```go
// Filter and sort
result, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 50,
    Filter:  `created >= "2022-01-01 00:00:00" && status = "published"`,
    Sort:    "-created,title", // DESC by created, ASC by title
    Expand:  "author,categories",
})

// Filter with operators
result2, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
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
allPosts, err := pb.Collection("posts").GetFullList(200, &bosbase.CrudListOptions{
    Sort:   "-created",
    Filter: `status = "published"`,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Retrieved %d posts\n", len(allPosts))
```

#### Get First Matching Record

Get only the first record that matches a filter:

```go
post, err := pb.Collection("posts").GetFirstListItem(
    `slug = "my-post-slug"`,
    &bosbase.CrudViewOptions{
        Expand: "author,categories.tags",
    },
)
if err != nil {
    log.Fatal(err)
}
```

### View Record

Retrieve a single record by ID:

```go
// Basic retrieval
record, err := pb.Collection("posts").GetOne("RECORD_ID", nil)

// With expanded relations
record, err := pb.Collection("posts").GetOne("RECORD_ID", &bosbase.CrudViewOptions{
    Expand: "author,categories,tags",
})

// Nested expand
record, err := pb.Collection("comments").GetOne("COMMENT_ID", &bosbase.CrudViewOptions{
    Expand: "post.author,user",
})

// Field selection
record, err := pb.Collection("posts").GetOne("RECORD_ID", &bosbase.CrudViewOptions{
    Fields: "id,title,content,author.name",
})
```

### Create Record

Create a new record:

```go
// Simple create
record, err := pb.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":   "My First Post",
        "content": "Lorem ipsum...",
        "status":  "draft",
    },
})

// Create with relations
record, err := pb.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":      "My Post",
        "author":     "AUTHOR_ID",              // Single relation
        "categories": []string{"cat1", "cat2"},   // Multiple relation
    },
})

// Create with file upload
file, _ := os.Open("image.jpg")
defer file.Close()

record, err := pb.Collection("posts").Create(&bosbase.CrudMutateOptions{
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
record, err := pb.Collection("posts").Create(&bosbase.CrudMutateOptions{
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
record, err := pb.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":  "Updated Title",
        "status": "published",
    },
})

// Update with relations (using field modifiers)
record, err := pb.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "categories+": "NEW_CATEGORY_ID", // Append
        "tags-":       "OLD_TAG_ID",       // Remove
    },
})

// Update with file upload
file, _ := os.Open("new_image.jpg")
defer file.Close()

record, err := pb.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Updated Title",
    },
    Files: map[string]bosbase.FileParam{
        "image": {
            Filename:    "new_image.jpg",
            Reader:      file,
            ContentType: "image/jpeg",
        },
    },
})

// Update with expand
record, err := pb.Collection("posts").Update("RECORD_ID", &bosbase.CrudMutateOptions{
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
err := pb.Collection("posts").Delete("RECORD_ID", nil)
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
filter := `status = "published"`

// Not equal
filter := `status != "draft"`

// Greater than / Less than
filter := `views > 100`
filter := `created < "2023-01-01"`

// Greater/Less than or equal
filter := `age >= 18`
filter := `price <= 99.99`
```

### String Operators

```go
// Contains (like)
filter := `title ~ "javascript"`
// Equivalent to: title LIKE "%javascript%"

// Not contains
filter := `title !~ "deprecated"`

// Exact match (case-sensitive)
filter := `email = "user@example.com"`
```

### Array Operators (for multiple relations/files)

```go
// Any of / At least one
filter := `tags.id ?= "TAG_ID"`         // Any tag matches
filter := `tags.name ?~ "important"`     // Any tag name contains "important"

// All must match
filter := `tags.id = "TAG_ID" && tags.id = "TAG_ID2"`
```

### Logical Operators

```go
// AND
filter := `status = "published" && views > 100`

// OR
filter := `status = "published" || status = "featured"`

// Parentheses for grouping
filter := `(status = "published" || featured = true) && views > 50`
```

### Special Identifiers

```go
// Request context (only in API rules, not client filters)
// @request.auth.id, @request.query.*, etc.

// Collection joins
filter := `@collection.users.email = "test@example.com"`

// Record fields
filter := `author.id = @request.auth.id`
```

### Comments

```go
// Single-line comments are supported
filter := `status = "published" // Only published posts`
```

## Sorting

Sort records using the `sort` parameter:

```go
// Single field (ASC)
sort := "created"

// Single field (DESC)
sort := "-created"

// Multiple fields
sort := "-created,title"  // DESC by created, then ASC by title

// Supported fields
sort := "@random"         // Random order
sort := "@rowid"          // Internal row ID
sort := "id"              // Record ID
sort := "fieldName"       // Any collection field

// Relation field sorting
sort := "author.name"     // Sort by related author's name
```

## Field Selection

Control which fields are returned:

```go
// Specific fields
fields := "id,title,content"

// All fields at level
fields := "*"

// Nested field selection
fields := "*,author.name,author.email"

// Excerpt modifier for text fields
fields := "*,content:excerpt(200,true)"
// Returns first 200 characters with ellipsis if truncated

// Combined
fields := "*,content:excerpt(200),author.name,author.email"
```

## Expanding Relations

Expand related records without additional API calls:

```go
// Single relation
expand := "author"

// Multiple relations
expand := "author,categories,tags"

// Nested relations (up to 6 levels)
expand := "author.profile,categories.tags"

// Back-relations
expand := "comments_via_post.user"
```

See [Relations Documentation](./RELATIONS.md) for detailed information.

## Pagination Options

```go
// Skip total count (faster queries)
result, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:      1,
    PerPage:   50,
    SkipTotal: true, // totalItems and totalPages will be -1
    Filter:    `status = "published"`,
})

// Get Full List with batch processing
allPosts, err := pb.Collection("posts").GetFullList(200, &bosbase.CrudListOptions{
    Sort: "-created",
})
// Processes in batches of 200 to avoid memory issues
```

## Batch Operations

Execute multiple operations in a single transaction:

```go
// Create a batch
batch := pb.CreateBatch()

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
methods, err := pb.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

password, _ := methods["password"].(map[string]interface{})
oauth2, _ := methods["oauth2"].(map[string]interface{})
otp, _ := methods["otp"].(map[string]interface{})
mfa, _ := methods["mfa"].(map[string]interface{})

fmt.Println("Password enabled:", password["enabled"])
fmt.Println("OAuth2 enabled:", oauth2["enabled"])
fmt.Println("OTP enabled:", otp["enabled"])
fmt.Println("MFA enabled:", mfa["enabled"])
```

### Auth with Password

```go
authData, err := pb.Collection("users").AuthWithPassword(
    "user@example.com", // username or email
    "password123",
    "", "", nil, nil, nil,
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

// With expand
authData, err = pb.Collection("users").AuthWithPassword(
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
methods, err := pb.Collection("users").ListAuthMethods("", nil, nil)
if err != nil {
    log.Fatal(err)
}

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

// Step 2: After redirect, exchange code for token
authData, err := pb.Collection("users").AuthWithOAuth2Code(
    "google",                    // Provider name
    "AUTHORIZATION_CODE",        // From redirect URL
    provider["codeVerifier"].(string), // From step 1
    "https://yourapp.com/callback",    // Redirect URL
    nil,                         // createData
    nil,                         // body
    nil,                         // query
    nil,                         // headers
    "",                          // expand
    "",                          // fields
)
```

### Auth with OTP (One-Time Password)

```go
// Step 1: Request OTP
otpRequest, err := pb.Collection("users").RequestOTP("user@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
otpID := otpRequest["otpId"].(string)

// Step 2: User enters OTP from email
// Step 3: Authenticate with OTP
authData, err := pb.Collection("users").AuthWithOTP(
    otpID,
    "123456", // OTP from email
    "", "", nil, nil, nil,
)
```

### Auth Refresh

Refresh the current auth token and get updated user data:

```go
// Refresh auth (useful on page reload)
authData, err := pb.Collection("users").AuthRefresh("", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Check if still valid
if pb.AuthStore.IsValid() {
    fmt.Println("User is authenticated")
} else {
    fmt.Println("Token expired or invalid")
}
```

### Email Verification

```go
// Request verification email
err := pb.Collection("users").RequestVerification("user@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Confirm verification (on verification page)
err = pb.Collection("users").ConfirmVerification("VERIFICATION_TOKEN", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### Password Reset

```go
// Request password reset email
err := pb.Collection("users").RequestPasswordReset("user@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Confirm password reset (on reset page)
// Note: This invalidates all previous auth tokens
err = pb.Collection("users").ConfirmPasswordReset(
    "RESET_TOKEN",
    "newpassword123",
    "newpassword123", // Confirm
    nil,             // body
    nil,             // query
    nil,             // headers
)
if err != nil {
    log.Fatal(err)
}
```

### Email Change

```go
// Must be authenticated first
_, err := pb.Collection("users").AuthWithPassword("user@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Request email change
err = pb.Collection("users").RequestEmailChange("newemail@example.com", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Confirm email change (on confirmation page)
// Note: This invalidates all previous auth tokens
err = pb.Collection("users").ConfirmEmailChange(
    "EMAIL_CHANGE_TOKEN",
    "currentpassword",
    nil, // body
    nil, // query
    nil, // headers
)
if err != nil {
    log.Fatal(err)
}
```

### Impersonate (Superuser Only)

Generate a token to authenticate as another user:

```go
// Must be authenticated as superuser
_, err := pb.Admins.AuthWithPassword("admin@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Impersonate a user
impersonateClient, err := pb.Collection("users").Impersonate("USER_ID", 3600, "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Use the impersonated client
posts, err := impersonateClient.Collection("posts").GetFullList(200, nil)
if err != nil {
    log.Fatal(err)
}

// Access the token
fmt.Println("Token:", impersonateClient.AuthStore.Token())
fmt.Println("Record:", impersonateClient.AuthStore.Record())
```

## Complete Examples

### Example 1: Blog Post Search with Filters

```go
func searchPosts(pb *bosbase.BosBase, query, categoryID string, minViews int) ([]interface{}, error) {
    filter := fmt.Sprintf(`title ~ "%s" || content ~ "%s"`, query, query)
    
    if categoryID != "" {
        filter += fmt.Sprintf(` && categories.id ?= "%s"`, categoryID)
    }
    
    if minViews > 0 {
        filter += fmt.Sprintf(` && views >= %d`, minViews)
    }
    
    result, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 20,
        Filter:  filter,
        Sort:    "-created",
        Expand:  "author,categories",
    })
    if err != nil {
        return nil, err
    }
    
    if items, ok := result["items"].([]interface{}); ok {
        return items, nil
    }
    return nil, nil
}
```

### Example 2: User Dashboard with Related Content

```go
func getUserDashboard(pb *bosbase.BosBase, userID string) (map[string]interface{}, error) {
    // Get user's posts
    postsResult, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
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
    commentsResult, err := pb.Collection("comments").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 10,
        Filter:  fmt.Sprintf(`user = "%s"`, userID),
        Sort:    "-created",
        Expand:  "post",
    })
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "posts":    postsResult["items"],
        "comments": commentsResult["items"],
    }, nil
}
```

### Example 3: Advanced Filtering

```go
// Complex filter example
result, err := pb.Collection("posts").GetList(&bosbase.CrudListOptions{
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
func createMultiplePosts(pb *bosbase.BosBase, postsData []map[string]interface{}) ([]interface{}, error) {
    batch := pb.CreateBatch()
    
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
func getAllRecordsPaginated(pb *bosbase.BosBase, collectionName string, options *bosbase.CrudListOptions) ([]interface{}, error) {
    var allRecords []interface{}
    page := 1
    
    for {
        if options == nil {
            options = &bosbase.CrudListOptions{}
        }
        options.Page = page
        options.PerPage = 500
        options.SkipTotal = true // Skip count for performance
        
        result, err := pb.Collection(collectionName).GetList(options)
        if err != nil {
            return nil, err
        }
        
        items, _ := result["items"].([]interface{})
        allRecords = append(allRecords, items...)
        
        if len(items) < 500 {
            break
        }
        page++
    }
    
    return allRecords, nil
}
```

## Error Handling

```go
record, err := pb.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Post",
    },
})
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        switch clientErr.Status {
        case 400:
            // Validation error
            fmt.Println("Validation errors:", clientErr.Response)
        case 403:
            // Permission denied
            fmt.Println("Access denied")
        case 404:
            // Not found
            fmt.Println("Collection or record not found")
        default:
            fmt.Println("Unexpected error:", err)
        }
    } else {
        fmt.Println("Error:", err)
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

