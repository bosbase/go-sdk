# Files Upload and Handling - Go SDK Documentation

## Overview

BosBase allows you to upload and manage files through file fields in your collections. Files are stored with sanitized names and a random suffix for security (e.g., `test_52iwbgds7l.png`).

**Key Features:**
- Upload multiple files per field
- Maximum file size: ~8GB (2^53-1 bytes)
- Automatic filename sanitization and random suffix
- Image thumbnails support
- Protected files with token-based access
- File modifiers for append/prepend/delete operations

**Backend Endpoints:**
- `POST /api/files/token` - Get file access token for protected files
- `GET /api/files/{collection}/{recordId}/{filename}` - Download file

## File Field Configuration

Before uploading files, you must add a file field to your collection:

```go
collection, err := client.Collections.GetOne("example", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})
newField := map[string]interface{}{
    "name":     "documents",
    "type":     "file",
    "maxSelect": 5,        // Maximum number of files (1 for single file)
    "maxSize":  5242880,   // 5MB in bytes (optional, default: 5MB)
    "mimeTypes": []string{"image/jpeg", "image/png", "application/pdf"},
    "thumbs":   []string{"100x100", "300x300"}, // Thumbnail sizes for images
    "protected": false,   // Require token for access
}
fields = append(fields, newField)

_, err = client.Collections.Update("example", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

## Uploading Files

### Basic Upload with Create

When creating a new record, you can upload files directly:

```go
package main

import (
    "log"
    "os"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://localhost:8090")
    defer client.Close()
    
    // Open file
    file, err := os.Open("image.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    // Create record with file upload
    record, err := client.Collection("example").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "title": "Hello world!",
        },
        Files: map[string]bosbase.FileParam{
            "documents": {
                Filename:    "image.jpg",
                Reader:      file,
                ContentType: "image/jpeg",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Created record:", record)
}
```

### Upload with Update

```go
// Update record and upload new files
file, err := os.Open("newfile.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

_, err = client.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Updated title",
    },
    Files: map[string]bosbase.FileParam{
        "documents": {
            Filename:    "newfile.txt",
            Reader:      file,
            ContentType: "text/plain",
        },
    },
})
```

### Append Files (Using + Modifier)

For multiple file fields, use the `+` modifier to append files:

```go
// Append files to existing ones
file, err := os.Open("file4.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

_, err = client.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "documents+": "file4.txt", // Note: In Go SDK, file operations use Files map
    },
    Files: map[string]bosbase.FileParam{
        "documents+": {
            Filename:    "file4.txt",
            Reader:      file,
            ContentType: "text/plain",
        },
    },
})
```

### Upload Multiple Files

```go
files := []struct {
    path        string
    contentType string
}{
    {"file1.txt", "text/plain"},
    {"file2.pdf", "application/pdf"},
    {"file3.jpg", "image/jpeg"},
}

fileParams := make(map[string]bosbase.FileParam)
for i, f := range files {
    file, err := os.Open(f.path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    fileParams[fmt.Sprintf("documents+")] = bosbase.FileParam{
        Filename:    filepath.Base(f.path),
        Reader:      file,
        ContentType: f.contentType,
    }
}

_, err = client.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Files: fileParams,
})
```

## Deleting Files

### Delete All Files

```go
// Delete all files in a field (set to empty array)
_, err := client.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "documents": []string{},
    },
})
```

### Delete Specific Files (Using - Modifier)

```go
// Delete individual files by filename
_, err := client.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "documents-": []string{"file1.pdf", "file2.txt"},
    },
})
```

## File URLs

### Get File URL

Each uploaded file can be accessed via its URL:

```
http://localhost:8090/api/files/COLLECTION_ID_OR_NAME/RECORD_ID/FILENAME
```

**Using SDK:**

```go
record, err := client.Collection("example").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

// Single file field (returns string)
filename, _ := record["documents"].(string)
url := client.Files.GetURL(record, filename, nil)

// Multiple file field (returns array)
filenames, _ := record["documents"].([]interface{})
if len(filenames) > 0 {
    firstFile, _ := filenames[0].(string)
    url := client.Files.GetURL(record, firstFile, nil)
    fmt.Println("File URL:", url)
}
```

### Image Thumbnails

If your file field has thumbnail sizes configured, you can request thumbnails:

```go
record, err := client.Collection("example").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

filename, _ := record["avatar"].(string) // Image file

// Get thumbnail with specific size
thumbUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "100x300", // Width x Height
})
```

**Thumbnail Formats:**

- `WxH` (e.g., `100x300`) - Crop to WxH viewbox from center
- `WxHt` (e.g., `100x300t`) - Crop to WxH viewbox from top
- `WxHb` (e.g., `100x300b`) - Crop to WxH viewbox from bottom
- `WxHf` (e.g., `100x300f`) - Fit inside WxH viewbox (no cropping)
- `0xH` (e.g., `0x300`) - Resize to H height, preserve aspect ratio
- `Wx0` (e.g., `100x0`) - Resize to W width, preserve aspect ratio

**Supported Image Formats:**
- JPEG (`.jpg`, `.jpeg`)
- PNG (`.png`)
- GIF (`.gif` - first frame only)
- WebP (`.webp` - stored as PNG)

**Example:**

```go
record, err := client.Collection("products").GetOne("PRODUCT_ID", nil)
if err != nil {
    log.Fatal(err)
}

image, _ := record["image"].(string)

// Different thumbnail sizes
thumbSmall := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "100x100",
})
thumbMedium := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "300x300f",
})
thumbLarge := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "800x600",
})
```

### Force Download

To force browser download instead of preview:

```go
url := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Download: true, // Force download
})
```

## Protected Files

By default, all files are publicly accessible if you know the full URL. For sensitive files, you can mark the field as "Protected" in the collection settings.

### Setting Up Protected Files

```go
collection, err := client.Collections.GetOne("example", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})
for i, field := range fields {
    if f, ok := field.(map[string]interface{}); ok {
        if name, _ := f["name"].(string); name == "documents" {
            f["protected"] = true
            fields[i] = f
        }
    }
}

_, err = client.Collections.Update("example", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

### Accessing Protected Files

Protected files require authentication and a file token:

```go
// Step 1: Authenticate
_, err := client.Collection("users").AuthWithPassword(
    "user@example.com", "password123", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Step 2: Get file token (valid for ~2 minutes)
fileToken, err := client.Files.GetToken(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Step 3: Get protected file URL with token
record, err := client.Collection("example").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

filename, _ := record["privateDocument"].(string)
url := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Token: fileToken,
})

fmt.Println("Protected file URL:", url)
```

**Important:**
- File tokens are short-lived (~2 minutes)
- Only authenticated users satisfying the collection's `viewRule` can access protected files
- Tokens must be regenerated when they expire

### Complete Protected File Example

```go
func loadProtectedImage(client *bosbase.BosBase, recordID, filename string) (string, error) {
    // Check if authenticated
    if !client.AuthStore.IsValid() {
        return "", fmt.Errorf("not authenticated")
    }
    
    // Get fresh token
    token, err := client.Files.GetToken(nil, nil, nil)
    if err != nil {
        return "", err
    }
    
    // Get file URL
    record, err := client.Collection("documents").GetOne(recordID, nil)
    if err != nil {
        return "", err
    }
    
    url := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
    })
    
    return url, nil
}
```

## Complete Examples

### Example 1: Image Upload with Thumbnails

```go
package main

import (
    "fmt"
    "log"
    "os"
    
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
    
    // Create collection with image field and thumbnails
    _, err = client.Collections.CreateBase("products", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "name", "type": "text", "required": true},
            {
                "name":     "image",
                "type":     "file",
                "maxSelect": 1,
                "mimeTypes": []string{"image/jpeg", "image/png"},
                "thumbs":   []string{"100x100", "300x300", "800x600f"}, // Thumbnail sizes
            },
        },
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Upload product with image
    file, err := os.Open("product.jpg")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    product, err := client.Collection("products").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "name": "My Product",
        },
        Files: map[string]bosbase.FileParam{
            "image": {
                Filename:    "product.jpg",
                Reader:      file,
                ContentType: "image/jpeg",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Display thumbnail in UI
    image, _ := product["image"].(string)
    thumbnailUrl := client.Files.GetURL(product, image, &bosbase.FileURLOptions{
        Thumb: "300x300",
    })
    
    fmt.Println("Thumbnail URL:", thumbnailUrl)
}
```

### Example 2: Multiple File Upload

```go
func uploadMultipleFiles(client *bosbase.BosBase, collectionID, recordID string, filePaths []string) error {
    fileParams := make(map[string]bosbase.FileParam)
    
    for _, path := range filePaths {
        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()
        
        filename := filepath.Base(path)
        contentType := mime.TypeByExtension(filepath.Ext(path))
        if contentType == "" {
            contentType = "application/octet-stream"
        }
        
        fileParams["documents+"] = bosbase.FileParam{
            Filename:    filename,
            Reader:      file,
            ContentType: contentType,
        }
    }
    
    _, err := client.Collection(collectionID).Update(recordID, &bosbase.CrudMutateOptions{
        Files: fileParams,
    })
    
    return err
}
```

### Example 3: File Management

```go
type FileManager struct {
    client     *bosbase.BosBase
    collection string
    recordID   string
}

func (fm *FileManager) Load() (map[string]interface{}, error) {
    return fm.client.Collection(fm.collection).GetOne(fm.recordID, nil)
}

func (fm *FileManager) DeleteFile(filename string) error {
    _, err := fm.client.Collection(fm.collection).Update(fm.recordID, &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "documents-": []string{filename},
        },
    })
    return err
}

func (fm *FileManager) AddFiles(filePaths []string) error {
    fileParams := make(map[string]bosbase.FileParam)
    
    for _, path := range filePaths {
        file, err := os.Open(path)
        if err != nil {
            return err
        }
        defer file.Close()
        
        fileParams["documents+"] = bosbase.FileParam{
            Filename:    filepath.Base(path),
            Reader:      file,
            ContentType: mime.TypeByExtension(filepath.Ext(path)),
        }
    }
    
    _, err := fm.client.Collection(fm.collection).Update(fm.recordID, &bosbase.CrudMutateOptions{
        Files: fileParams,
    })
    return err
}

// Usage
manager := &FileManager{
    client:     client,
    collection: "example",
    recordID:   "RECORD_ID",
}

record, _ := manager.Load()
fmt.Println("Files:", record["documents"])

err := manager.AddFiles([]string{"file1.txt", "file2.pdf"})
if err != nil {
    log.Fatal(err)
}
```

## File Field Modifiers

### Summary

- **No modifier** - Replace all files: `documents: []string{}`
- **`+` suffix** - Append files: Use `Files` map with `"documents+"` key
- **`-` suffix** - Delete files: `documents-: []string{"file1.pdf"}`

## Best Practices

1. **File Size Limits**: Always validate file sizes on the client before upload
2. **MIME Types**: Configure allowed MIME types in collection field settings
3. **Thumbnails**: Pre-generate common thumbnail sizes for better performance
4. **Protected Files**: Use protected files for sensitive documents (ID cards, contracts)
5. **Token Refresh**: Refresh file tokens before they expire for protected files
6. **Error Handling**: Handle 404 errors for missing files and 401 for protected file access
7. **Filename Sanitization**: Files are automatically sanitized, but validate on client side too

## Error Handling

```go
record, err := client.Collection("example").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Test",
    },
    Files: map[string]bosbase.FileParam{
        "documents": {
            Filename:    "test.txt",
            Reader:      strings.NewReader("content"),
            ContentType: "text/plain",
        },
    },
})
if err != nil {
    if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
        switch clientErr.Status {
        case 413:
            fmt.Println("File too large")
        case 400:
            fmt.Println("Invalid file type or field validation failed")
        case 403:
            fmt.Println("Insufficient permissions")
        default:
            fmt.Printf("Upload failed: %v\n", err)
        }
    } else {
        fmt.Printf("Upload failed: %v\n", err)
    }
}
```

## Storage Options

By default, BosBase stores files in `pb_data/storage` on the local filesystem. For production, you can configure S3-compatible storage (AWS S3, MinIO, Wasabi, DigitalOcean Spaces, etc.) from:
**Dashboard > Settings > Files storage**

This is configured server-side and doesn't require SDK changes.

## Related Documentation

- [Collections](./COLLECTIONS.md) - Collection and field configuration
- [Authentication](./AUTHENTICATION.md) - Required for protected files
- [File API](./FILE_API.md) - File download and URL generation

