# API Rules and Filters - Go SDK Documentation

## Overview

API Rules are your collection access controls and data filters. They control who can perform actions on your collections and what data they can access.

Each collection has 5 rules, corresponding to specific API actions:
- `listRule` - Controls who can list records
- `viewRule` - Controls who can view individual records
- `createRule` - Controls who can create records
- `updateRule` - Controls who can update records
- `deleteRule` - Controls who can delete records

Auth collections have an additional `manageRule` that allows one user to fully manage another user's data.

## Rule Values

Each rule can be set to:

- **`nil` (locked)** - Only authorized superusers can perform the action (default)
- **Empty string `""`** - Anyone can perform the action (superusers, authenticated users, and guests)
- **Non-empty string** - Only users that satisfy the filter expression can perform the action

## Important Notes

1. **Rules act as filters**: API Rules also act as record filters. For example, setting `listRule` to `status = "active"` will only return active records.
2. **HTTP Status Codes**: 
   - `200` with empty items for unsatisfied `listRule`
   - `400` for unsatisfied `createRule`
   - `404` for unsatisfied `viewRule`, `updateRule`, `deleteRule`
   - `403` for locked rules when not a superuser
3. **Superuser bypass**: API Rules are ignored when the action is performed by an authorized superuser.

## Setting Rules via SDK

### Go SDK

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
    
    // Authenticate as admin
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create collection with rules
    collection, err := client.Collections.CreateBase("articles", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "title", "type": "text", "required": true},
            {"name": "status", "type": "select", "options": map[string]interface{}{
                "values": []string{"draft", "published"},
            }, "maxSelect": 1},
            {"name": "author", "type": "relation", "options": map[string]interface{}{
                "collectionId": "users",
            }, "maxSelect": 1},
        },
        "listRule":  `@request.auth.id != "" || status = "published"`,
        "viewRule":  `@request.auth.id != "" || status = "published"`,
        "createRule": `@request.auth.id != ""`,
        "updateRule": `author = @request.auth.id || @request.auth.role = "admin"`,
        "deleteRule": `author = @request.auth.id || @request.auth.role = "admin"`,
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created collection:", collection)
    
    // Update rules
    _, err = client.Collections.Update("articles", &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "listRule": `@request.auth.id != "" && (status = "published" || status = "draft")`,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Remove rule (set to empty string for public access)
    _, err = client.Collections.Update("articles", &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "listRule": "", // Anyone can list
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Lock rule (set to nil for superuser only)
    _, err = client.Collections.Update("articles", &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "deleteRule": nil, // Only superusers can delete
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Filter Syntax

The syntax follows: `OPERAND OPERATOR OPERAND`

### Operators

**Comparison Operators:**
- `=` - Equal
- `!=` - NOT equal
- `>` - Greater than
- `>=` - Greater than or equal
- `<` - Less than
- `<=` - Less than or equal

**String Operators:**
- `~` - Like/Contains (auto-wraps right operand in `%` for wildcard match)
- `!~` - NOT Like/Contains

**Array Operators (Any/At least one of):**
- `?=` - Any Equal
- `?!=` - Any NOT equal
- `?>` - Any Greater than
- `?>=` - Any Greater than or equal
- `?<` - Any Less than
- `?<=` - Any Less than or equal
- `?~` - Any Like/Contains
- `?!~` - Any NOT Like/Contains

**Logical Operators:**
- `&&` - AND
- `||` - OR
- `()` - Grouping
- `//` - Single line comments

## Special Identifiers

### @request.*

Access current request data:

**@request.context** - The context where the rule is used
```go
listRule: `@request.context != "oauth2"`
```

**@request.method** - HTTP request method
```go
updateRule: `@request.method = "PATCH"`
```

**@request.headers.*** - Request headers (normalized to lowercase, `-` replaced with `_`)
```go
listRule: `@request.headers.x_token = "test"`
```

**@request.query.*** - Query parameters
```go
listRule: `@request.query.page = "1"`
```

**@request.auth.*** - Current authenticated user
```go
listRule: `@request.auth.id != ""`
viewRule: `@request.auth.email = "admin@example.com"`
updateRule: `@request.auth.role = "admin"`
```

**@request.body.*** - Submitted body parameters
```go
createRule: `@request.body.title != ""`
updateRule: `@request.body.status:isset = false`  // Prevent changing status
```

### @collection.*

Target other collections that aren't directly related:

```go
// Check if user has access to related collection
listRule: `@request.auth.id != "" && @collection.news.categoryId ?= categoryId && @collection.news.author ?= @request.auth.id`

// Using aliases for multiple joins
listRule: `@request.auth.id != "" && @collection.courseRegistrations.user ?= id && @collection.courseRegistrations:auth.user ?= @request.auth.id && @collection.courseRegistrations.courseGroup ?= @collection.courseRegistrations:auth.courseGroup`
```

### @ Macros (Datetime)

All macros are UTC-based:

- `@now` - Current datetime as string
- `@second` - Current second (0-59)
- `@minute` - Current minute (0-59)
- `@hour` - Current hour (0-23)
- `@weekday` - Current weekday (0-6)
- `@day` - Current day
- `@month` - Current month
- `@year` - Current year
- `@yesterday` - Yesterday datetime
- `@tomorrow` - Tomorrow datetime
- `@todayStart` - Beginning of current day
- `@todayEnd` - End of current day
- `@monthStart` - Beginning of current month
- `@monthEnd` - End of current month
- `@yearStart` - Beginning of current year
- `@yearEnd` - End of current year

**Example:**
```go
listRule: `@request.body.publicDate >= @now`
listRule: `created >= @todayStart && created <= @todayEnd`
```

## Field Modifiers

### :isset

Check if a field was submitted in the request (only for `@request.*` fields):

```go
// Prevent changing role field
updateRule: `@request.body.role:isset = false`

// Require email field
createRule: `@request.body.email:isset = true`
```

### :length

Check the number of items in an array field (multiple file, select, relation):

```go
// Check submitted array length
createRule: `@request.body.tags:length > 1 && @request.body.tags:length <= 5`

// Check existing record array length
listRule: `categories:length = 2`
listRule: `documents:length >= 1`
```

### :each

Apply condition on each item in an array field:

```go
// Check if all submitted select options contain "create"
createRule: `@request.body.permissions:each ~ "create"`

// Check if all existing field values have "pb_" prefix
listRule: `tags:each ~ "pb_%"`
```

### :lower

Perform case-insensitive string comparisons:

```go
// Case-insensitive comparison
listRule: `@request.body.title:lower = "test"`
updateRule: `status:lower ~ "active"`
```

## geoDistance Function

Calculate Haversine distance between two geographic points in kilometers:

```go
// Offices within 25km of location
listRule: `geoDistance(address.lon, address.lat, 23.32, 42.69) < 25`

// Using request data
listRule: `geoDistance(location.lon, location.lat, @request.query.lon, @request.query.lat) < @request.query.radius`
```

## Common Rule Examples

### Allow Only Authenticated Users

```go
listRule:  `@request.auth.id != ""`
viewRule:  `@request.auth.id != ""`
createRule: `@request.auth.id != ""`
updateRule: `@request.auth.id != ""`
deleteRule: `@request.auth.id != ""`
```

### Owner-Based Access

```go
viewRule:  `@request.auth.id != "" && author = @request.auth.id`
updateRule: `@request.auth.id != "" && author = @request.auth.id`
deleteRule: `@request.auth.id != "" && author = @request.auth.id`
```

### Role-Based Access

```go
// Assuming users have a "role" select field
listRule: `@request.auth.id != "" && @request.auth.role = "admin"`
updateRule: `@request.auth.role = "admin" || author = @request.auth.id`
```

### Public with Authentication

```go
// Public can view published, authenticated can view all
listRule: `@request.auth.id != "" || status = "published"`
viewRule: `@request.auth.id != "" || status = "published"`
```

### Filtered Results

```go
// Only show active records
listRule: `status = "active"`

// Only show records from last 30 days
listRule: `created >= @yesterday`

// Only show records matching user's organization
listRule: `@request.auth.id != "" && organization = @request.auth.organization`
```

### Complex Rules

```go
// Multiple conditions
listRule: `@request.auth.id != "" && (status = "published" || status = "draft") && author = @request.auth.id`

// Nested relations
listRule: `@request.auth.id != "" && author.role = "staff"`

// Back relations
listRule: `@request.auth.id != "" && comments_via_author.id != ""`
```

## Using Filters in Queries

Filters can also be used in regular queries (not just rules):

```go
// List with filter
result, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  `status = "published" && created >= @todayStart`,
})

// Complex filter
result, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  `(title ~ "test" || description ~ "test") && status = "published"`,
})

// Using relation filters
result, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  `author.role = "admin" && categories.id ?= "CAT_ID"`,
})

// Geo distance filter
result, err := client.Collection("offices").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  `geoDistance(location.lon, location.lat, 23.32, 42.69) < 25`,
})
```

## Complete Example

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
    
    // Authenticate as admin
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create users collection with role field
    users, err := client.Collections.CreateAuth("users", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "name", "type": "text", "required": true},
            {"name": "role", "type": "select", "options": map[string]interface{}{
                "values": []string{"user", "staff", "admin"},
            }, "maxSelect": 1},
        },
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    usersID, _ := users["id"].(string)
    
    // Create articles collection with comprehensive rules
    articles, err := client.Collections.CreateBase("articles", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "title", "type": "text", "required": true},
            {"name": "content", "type": "editor", "required": true},
            {"name": "status", "type": "select", "options": map[string]interface{}{
                "values": []string{"draft", "published", "archived"},
            }, "maxSelect": 1},
            {"name": "author", "type": "relation", "options": map[string]interface{}{
                "collectionId": usersID,
            }, "maxSelect": 1, "required": true},
            {"name": "categories", "type": "relation", "options": map[string]interface{}{
                "collectionId": "categories",
            }, "maxSelect": 5},
            {"name": "published_at", "type": "date"},
        },
        // Public can see published, authenticated can see their own or published
        "listRule": `@request.auth.id != "" && (author = @request.auth.id || status = "published") || status = "published"`,
        // Same logic for viewing
        "viewRule": `@request.auth.id != "" && (author = @request.auth.id || status = "published") || status = "published"`,
        // Only authenticated users can create
        "createRule": `@request.auth.id != ""`,
        // Owners or admins can update, but prevent changing status after publishing
        "updateRule": `@request.auth.id != "" && (author = @request.auth.id || @request.auth.role = "admin") && (@request.body.status:isset = false || status != "published")`,
        // Only owners or admins can delete
        "deleteRule": `@request.auth.id != "" && (author = @request.auth.id || @request.auth.role = "admin")`,
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Created collections:", users, articles)
    
    // Authenticate as regular user
    _, err = client.Collection("users").AuthWithPassword(
        "user@example.com", "password123", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // User can create article
    authRecord := client.AuthStore.Record()
    userID, _ := authRecord["id"].(string)
    
    article, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "title":   "My Article",
            "content": "<p>Content</p>",
            "status":  "draft",
            "author":  userID,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    articleID, _ := article["id"].(string)
    
    // User can update their own article
    _, err = client.Collection("articles").Update(articleID, &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "title": "Updated Title",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // User can list their own articles or published ones
    myArticles, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 20,
        Filter:  fmt.Sprintf(`author = "%s"`, userID),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("My articles:", myArticles)
    
    // User can also query with additional filters
    published, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
        Page:    1,
        PerPage: 20,
        Filter:  `status = "published" && created >= @todayStart`,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Published articles:", published)
}
```

