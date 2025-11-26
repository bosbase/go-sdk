# Working with Relations - Go SDK Documentation

## Overview

Relations allow you to link records between collections. BosBase supports both single and multiple relations, and provides powerful features for expanding related records and working with back-relations.

**Key Features:**
- Single and multiple relations
- Expand related records without additional requests
- Nested relation expansion (up to 6 levels)
- Back-relations for reverse lookups
- Field modifiers for append/prepend/remove operations

**Relation Field Types:**
- **Single Relation**: Links to one record (MaxSelect <= 1)
- **Multiple Relation**: Links to multiple records (MaxSelect > 1)

**Backend Behavior:**
- Relations are stored as record IDs or arrays of IDs
- Expand only includes relations the client can view (satisfies View API Rule)
- Back-relations use format: `collectionName_via_fieldName`
- Back-relation expand limited to 1000 records per field

## Setting Up Relations

### Creating a Relation Field

```go
collection, err := client.Collections.GetOne("posts", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})

// Single relation field
newField1 := map[string]interface{}{
    "name":        "user",
    "type":        "relation",
    "collectionId": "users",
    "maxSelect":   1, // Single relation
    "required":    true,
}
fields = append(fields, newField1)

// Multiple relation field
newField2 := map[string]interface{}{
    "name":         "tags",
    "type":         "relation",
    "collectionId": "tags",
    "maxSelect":    10,          // Multiple relation (max 10)
    "minSelect":    1,           // Minimum 1 required
    "cascadeDelete": false,       // Don't delete post when tags deleted
}
fields = append(fields, newField2)

_, err = client.Collections.Update("posts", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

## Creating Records with Relations

### Single Relation

```go
// Create a post with a single user relation
post, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Post",
        "user":  "USER_ID", // Single relation ID
    },
})
```

### Multiple Relations

```go
// Create a post with multiple tags
post, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "My Post",
        "tags":  []string{"TAG_ID1", "TAG_ID2", "TAG_ID3"}, // Array of IDs
    },
})
```

### Mixed Relations

```go
// Create a comment with both single and multiple relations
comment, err := client.Collection("comments").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "message": "Great post!",
        "post":    "POST_ID",        // Single relation
        "user":    "USER_ID",        // Single relation
        "tags":    []string{"TAG1", "TAG2"}, // Multiple relation
    },
})
```

## Updating Relations

### Replace All Relations

```go
// Replace all tags
_, err := client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "tags": []string{"NEW_TAG1", "NEW_TAG2"},
    },
})
```

### Append Relations (Using + Modifier)

```go
// Append tags to existing ones
_, err := client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "tags+": "NEW_TAG_ID", // Append single tag
    },
})

// Append multiple tags
_, err = client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "tags+": []string{"TAG_ID1", "TAG_ID2"}, // Append multiple tags
    },
})
```

### Prepend Relations (Using + Prefix)

```go
// Prepend tags (tags will appear first)
_, err := client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "+tags": "PRIORITY_TAG", // Prepend single tag
    },
})

// Prepend multiple tags
_, err = client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "+tags": []string{"TAG1", "TAG2"}, // Prepend multiple tags
    },
})
```

### Remove Relations (Using - Modifier)

```go
// Remove single tag
_, err := client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "tags-": "TAG_ID_TO_REMOVE",
    },
})

// Remove multiple tags
_, err = client.Collection("posts").Update("POST_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "tags-": []string{"TAG1", "TAG2"},
    },
})
```

## Expanding Relations

The `expand` parameter allows you to fetch related records in a single request, eliminating the need for multiple API calls.

### Basic Expand

```go
// Get comment with expanded user
comment, err := client.Collection("comments").GetOne("COMMENT_ID", &bosbase.CrudViewOptions{
    Expand: "user",
})
if err != nil {
    log.Fatal(err)
}

expand, _ := comment["expand"].(map[string]interface{})
user, _ := expand["user"].(map[string]interface{})
name, _ := user["name"].(string)
fmt.Println("Author:", name)
```

### Expand Multiple Relations

```go
// Expand multiple relations (comma-separated)
comment, err := client.Collection("comments").GetOne("COMMENT_ID", &bosbase.CrudViewOptions{
    Expand: "user,post",
})

expand, _ := comment["expand"].(map[string]interface{})
user, _ := expand["user"].(map[string]interface{})
post, _ := expand["post"].(map[string]interface{})
userName, _ := user["name"].(string)
postTitle, _ := post["title"].(string)
fmt.Printf("Comment by %s on post: %s\n", userName, postTitle)
```

### Nested Expand (Dot Notation)

You can expand nested relations up to 6 levels deep using dot notation:

```go
// Expand post and its tags, and user
comment, err := client.Collection("comments").GetOne("COMMENT_ID", &bosbase.CrudViewOptions{
    Expand: "user,post.tags",
})

expand, _ := comment["expand"].(map[string]interface{})
post, _ := expand["post"].(map[string]interface{})
postExpand, _ := post["expand"].(map[string]interface{})
tags, _ := postExpand["tags"].([]interface{})
fmt.Printf("Post has %d tags\n", len(tags))
```

### Expand with List Requests

```go
// List comments with expanded users
result, err := client.Collection("comments").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Expand:  "user",
})

items, _ := result["items"].([]interface{})
for _, item := range items {
    comment, _ := item.(map[string]interface{})
    message, _ := comment["message"].(string)
    
    expand, _ := comment["expand"].(map[string]interface{})
    user, _ := expand["user"].(map[string]interface{})
    userName, _ := user["name"].(string)
    
    fmt.Printf("%s: %s\n", userName, message)
}
```

## Back-Relations

Back-relations allow you to query and expand records that reference the current record through a relation field.

### Back-Relation Syntax

The format is: `collectionName_via_fieldName`

- `collectionName`: The collection that contains the relation field
- `fieldName`: The name of the relation field that points to your record

### Example: Posts with Comments

```go
// Get a post and expand all comments that reference it
post, err := client.Collection("posts").GetOne("POST_ID", &bosbase.CrudViewOptions{
    Expand: "comments_via_post",
})

expand, _ := post["expand"].(map[string]interface{})
comments, _ := expand["comments_via_post"].([]interface{})
fmt.Printf("Post has %d comments\n", len(comments))
```

### Back-Relation with Nested Expand

```go
// Get post with comments, and expand each comment's user
post, err := client.Collection("posts").GetOne("POST_ID", &bosbase.CrudViewOptions{
    Expand: "comments_via_post.user",
})

expand, _ := post["expand"].(map[string]interface{})
comments, _ := expand["comments_via_post"].([]interface{})
for _, comment := range comments {
    c, _ := comment.(map[string]interface{})
    message, _ := c["message"].(string)
    
    commentExpand, _ := c["expand"].(map[string]interface{})
    user, _ := commentExpand["user"].(map[string]interface{})
    userName, _ := user["name"].(string)
    
    fmt.Printf("%s: %s\n", userName, message)
}
```

### Filtering with Back-Relations

```go
// List posts that have at least one comment containing "hello"
posts, err := client.Collection("posts").GetList(&bosbase.CrudListOptions{
    Page:    1,
    PerPage: 20,
    Filter:  `comments_via_post.message ?~ "hello"`,
    Expand:  "comments_via_post.user",
})
```

## Complete Examples

### Example 1: Blog Post with Author and Tags

```go
// Create a blog post with relations
post, err := client.Collection("posts").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title":   "Getting Started with BosBase",
        "content": "Lorem ipsum...",
        "author":  "AUTHOR_ID",           // Single relation
        "tags":    []string{"tag1", "tag2", "tag3"}, // Multiple relation
    },
})

// Retrieve with all relations expanded
fullPost, err := client.Collection("posts").GetOne(post["id"].(string), &bosbase.CrudViewOptions{
    Expand: "author,tags",
})

expand, _ := fullPost["expand"].(map[string]interface{})
author, _ := expand["author"].(map[string]interface{})
authorName, _ := author["name"].(string)
fmt.Printf("Author: %s\n", authorName)

tags, _ := expand["tags"].([]interface{})
fmt.Println("Tags:")
for _, tag := range tags {
    t, _ := tag.(map[string]interface{})
    tagName, _ := t["name"].(string)
    fmt.Printf("  - %s\n", tagName)
}
```

### Example 2: Dynamic Tag Management

```go
type PostManager struct {
    client *bosbase.BosBase
}

func (pm *PostManager) AddTag(postID, tagID string) error {
    _, err := pm.client.Collection("posts").Update(postID, &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "tags+": tagID,
        },
    })
    return err
}

func (pm *PostManager) RemoveTag(postID, tagID string) error {
    _, err := pm.client.Collection("posts").Update(postID, &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "tags-": tagID,
        },
    })
    return err
}

func (pm *PostManager) GetPostWithTags(postID string) (map[string]interface{}, error) {
    return pm.client.Collection("posts").GetOne(postID, &bosbase.CrudViewOptions{
        Expand: "tags",
    })
}

// Usage
manager := &PostManager{client: client}
err := manager.AddTag("POST_ID", "NEW_TAG_ID")
post, _ := manager.GetPostWithTags("POST_ID")
```

## Best Practices

1. **Use Expand Wisely**: Only expand relations you actually need to reduce response size and improve performance.
2. **Handle Missing Expands**: Always check if expand data exists before accessing:
   ```go
   if expand, ok := record["expand"].(map[string]interface{}); ok {
       if user, ok := expand["user"].(map[string]interface{}); ok {
           name, _ := user["name"].(string)
           fmt.Println(name)
       }
   }
   ```
3. **Pagination for Large Back-Relations**: If you expect more than 1000 related records, fetch them separately with pagination.
4. **Error Handling**: Handle cases where related records might not be accessible due to API rules.

## Related Documentation

- [Collections](./COLLECTIONS.md) - Collection and field configuration
- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Filtering and querying related records

