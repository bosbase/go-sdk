# Collections - Go SDK Documentation

## Overview

**Collections** represent your application data. Under the hood they are backed by plain SQLite tables that are generated automatically with the collection **name** and **fields** (columns).

A single entry of a collection is called a **record** (a single row in the SQL table).

## Collection Types

### Base Collection

Default collection type for storing any application data.

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
    
    // Create base collection
    collection, err := client.Collections.CreateBase("articles", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "title", "type": "text", "required": true},
            {"name": "description", "type": "text"},
        },
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created collection:", collection)
}
```

### View Collection

Read-only collection populated from a SQL SELECT statement.

```go
view, err := client.Collections.CreateView("post_stats", 
    `SELECT posts.id, posts.name, count(comments.id) as totalComments 
     FROM posts LEFT JOIN comments on comments.postId = posts.id 
     GROUP BY posts.id`,
    nil, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### Auth Collection

Base collection with authentication fields (email, password, etc.).

```go
users, err := client.Collections.CreateAuth("users", map[string]interface{}{
    "fields": []map[string]interface{}{
        {"name": "name", "type": "text", "required": true},
    },
}, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Collections API

### List Collections

```go
// Get paginated list
result, err := client.Collections.GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 50,
})

// Get full list
all, err := client.Collections.GetFullList(500, nil)
```

### Get Collection

```go
collection, err := client.Collections.GetOne("articles", nil)
if err != nil {
    log.Fatal(err)
}
```

### Create Collection

```go
// Using scaffolds
base, err := client.Collections.CreateBase("articles", nil, nil, nil, nil)
auth, err := client.Collections.CreateAuth("users", nil, nil, nil, nil)
view, err := client.Collections.CreateView("stats", "SELECT * FROM posts", nil, nil, nil, nil)

// Manual
collection, err := client.Collections.Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "type": "base",
        "name": "articles",
        "fields": []map[string]interface{}{
            {"name": "title", "type": "text", "required": true},
            {
                "name":     "created",
                "type":     "autodate",
                "required": false,
                "onCreate": true,
                "onUpdate": false,
            },
            {
                "name":     "updated",
                "type":     "autodate",
                "required": false,
                "onCreate": true,
                "onUpdate": true,
            },
        },
    },
})
```

### Update Collection

```go
// Update collection rules
_, err := client.Collections.Update("articles", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "listRule": "published = true",
    },
})

// Update collection name
_, err := client.Collections.Update("articles", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "name": "posts",
    },
})
```

### Add Fields to Collection

To add a new field to an existing collection, fetch the collection, add the field to the fields array, and update:

```go
// Get existing collection
collection, err := client.Collections.GetOne("articles", nil)
if err != nil {
    log.Fatal(err)
}

// Get fields
fields, _ := collection["fields"].([]interface{})

// Add new field
newField := map[string]interface{}{
    "name":    "views",
    "type":    "number",
    "min":     0,
    "onlyInt": true,
}
fields = append(fields, newField)

// Update collection with new field
_, err = client.Collections.Update("articles", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

### Delete Fields from Collection

To delete a field, fetch the collection, remove the field from the fields array, and update:

```go
// Get existing collection
collection, err := client.Collections.GetOne("articles", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})
var filteredFields []interface{}

// Remove field by filtering it out
for _, field := range fields {
    if f, ok := field.(map[string]interface{}); ok {
        if name, _ := f["name"].(string); name != "oldFieldName" {
            filteredFields = append(filteredFields, field)
        }
    }
}

// Update collection without the deleted field
_, err = client.Collections.Update("articles", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": filteredFields,
    },
})
```

### Modify Fields in Collection

To modify an existing field (e.g., change its type, add options, etc.), fetch the collection, update the field object, and save:

```go
// Get existing collection
collection, err := client.Collections.GetOne("articles", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})

// Find and modify a field
for i, field := range fields {
    if f, ok := field.(map[string]interface{}); ok {
        if name, _ := f["name"].(string); name == "title" {
            f["max"] = 200
            f["required"] = true
            fields[i] = f
        }
    }
}

// Save changes
_, err = client.Collections.Update("articles", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

### Delete Collection

```go
err := client.Collections.DeleteCollection("articles", nil)
if err != nil {
    log.Fatal(err)
}
```

## Records API

### List Records

```go
result, err := client.Collection("articles").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  "published = true",
    Sort:    "-created",
    Expand:  "author",
})
```

### Get Record

```go
record, err := client.Collection("articles").GetOne("RECORD_ID", &bosbase.CrudViewOptions{
    Expand: "author,category",
})
```

### Create Record

```go
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Article",
        "views+": 1, // Field modifier
    },
})
```

### Update Record

```go
_, err := client.Collection("articles").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":  "Updated",
        "views+": 1,
        "tags+":  "new-tag",
    },
})
```

### Delete Record

```go
err := client.Collection("articles").Delete("RECORD_ID", nil)
```

## Field Types

### BoolField

```go
// Field definition
{"name": "published", "type": "bool", "required": true}

// Usage
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "published": true,
    },
})
```

### NumberField

```go
// Field definition
{"name": "views", "type": "number", "min": 0}

// Usage with modifier
_, err := client.Collection("articles").Update("ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "views+": 1, // Increment
    },
})
```

### TextField

```go
// Field definition
{"name": "title", "type": "text", "required": true, "min": 6, "max": 100}

// Usage with autogenerate
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "slug:autogenerate": "article-",
    },
})
```

### EmailField

```go
// Field definition
{"name": "email", "type": "email", "required": true}
```

### URLField

```go
// Field definition
{"name": "website", "type": "url"}
```

### EditorField

```go
// Field definition
{"name": "content", "type": "editor", "required": true}

// Usage
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "content": "<p>HTML content</p>",
    },
})
```

### DateField

```go
// Field definition
{"name": "published_at", "type": "date"}

// Usage
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "published_at": "2024-11-10 18:45:27.123Z",
    },
})
```

### AutodateField

**Important Note:** Bosbase does not initialize `created` and `updated` fields by default. To use these fields, you must explicitly add them when initializing the collection. For autodate fields, `onCreate` and `onUpdate` must be direct properties of the field object:

```go
// Create field with proper structure
{
    "name":     "created",
    "type":     "autodate",
    "required": false,
    "onCreate": true,  // Set on record creation (direct property)
    "onUpdate": false, // Don't update on record update (direct property)
}

// For updated field
{
    "name":     "updated",
    "type":     "autodate",
    "required": false,
    "onCreate": true,  // Set on record creation (direct property)
    "onUpdate": true,  // Update on record update (direct property)
}

// The value is automatically set by the backend based on onCreate and onUpdate properties
```

### SelectField

```go
// Single select
{"name": "status", "type": "select", "options": map[string]interface{}{
    "values": []string{"draft", "published"},
}, "maxSelect": 1}

// Usage
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "status": "published",
    },
})

// Multiple select
{"name": "tags", "type": "select", "options": map[string]interface{}{
    "values": []string{"tech", "design"},
}, "maxSelect": 5}

// Usage with modifier
_, err := client.Collection("articles").Update("ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "tags+": "marketing",
    },
})
```

### FileField

```go
// Single file
{"name": "cover", "type": "file", "maxSelect": 1, "mimeTypes": []string{"image/jpeg"}}

// Usage with file upload
import (
    "os"
    "io"
)

file, err := os.Open("image.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Article",
    },
    Files: map[string]bosbase.FileParam{
        "cover": {
            Filename:    "image.jpg",
            Reader:      file,
            ContentType: "image/jpeg",
        },
    },
})
```

### RelationField

```go
// Field definition
{"name": "author", "type": "relation", "options": map[string]interface{}{
    "collectionId": "users",
}, "maxSelect": 1}

// Usage
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "author": "USER_ID",
    },
})

// Get with expanded relation
record, err := client.Collection("articles").GetOne("ID", &bosbase.CrudViewOptions{
    Expand: "author",
})
```

### JSONField

```go
// Field definition
{"name": "metadata", "type": "json"}

// Usage
record, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "metadata": map[string]interface{}{
            "seo": map[string]interface{}{
                "title": "SEO Title",
            },
        },
    },
})
```

### GeoPointField

```go
// Field definition
{"name": "location", "type": "geoPoint"}

// Usage
record, err := client.Collection("places").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "location": map[string]interface{}{
            "lon": 139.6917,
            "lat": 35.6586,
        },
    },
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
    
    // Create collections
    users, err := client.Collections.CreateAuth("users", nil, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    usersID, _ := users["id"].(string)
    articles, err := client.Collections.CreateBase("articles", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "title", "type": "text", "required": true},
            {
                "name":     "author",
                "type":     "relation",
                "options":  map[string]interface{}{"collectionId": usersID},
                "maxSelect": 1,
            },
        },
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Create and authenticate user
    user, err := client.Collection("users").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "email":          "user@example.com",
            "password":        "password123",
            "passwordConfirm": "password123",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    _, err = client.Collection("users").AuthWithPassword(
        "user@example.com", "password123", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    userID, _ := user["id"].(string)
    
    // Create article
    article, err := client.Collection("articles").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "title":  "My Article",
            "author": userID,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Created article:", article)
}
```

