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
collection, err := pb.Collections.GetOne("example", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})
newField := map[string]interface{}{
    "name":      "documents",
    "type":      "file",
    "maxSelect": 5,        // Maximum number of files (1 for single file)
    "maxSize":   5242880,  // 5MB in bytes (optional, default: 5MB)
    "mimeTypes": []string{"image/jpeg", "image/png", "application/pdf"},
    "thumbs":    []string{"100x100", "300x300"}, // Thumbnail sizes for images
    "protected": false,    // Require token for access
}
fields = append(fields, newField)

_, err = pb.Collections.Update("example", &bosbase.CrudMutateOptions{
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
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://localhost:8090")
    
    // Open file
    file, err := os.Open("file1.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    // Create record with file
    record, err := pb.Collection("example").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "title": "Hello world!",
        },
        Files: map[string]bosbase.FileParam{
            "documents": {
                Filename:    "file1.txt",
                Reader:      file,
                ContentType: "text/plain",
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Created record:", record)
}
```

### Upload with Update

```go
// Update record and upload new files
file, _ := os.Open("file3.txt")
defer file.Close()

record, err := pb.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Updated title",
    },
    Files: map[string]bosbase.FileParam{
        "documents": {
            Filename:    "file3.txt",
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
file, _ := os.Open("file4.txt")
defer file.Close()

_, err := pb.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "documents+": "file4.txt", // Note: In Go SDK, you still use Files map
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

## Deleting Files

### Delete All Files

```go
// Delete all files in a field (set to empty array)
_, err := pb.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "documents": []string{},
    },
})
```

### Delete Specific Files (Using - Modifier)

```go
// Delete individual files by filename
_, err := pb.Collection("example").Update("RECORD_ID", &bosbase.CrudMutateOptions{
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
record, err := pb.Collection("example").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

// Single file field (returns string)
filename := record["documents"].(string)
url := pb.Files.GetURL(record, filename, nil)

// Multiple file field (returns array)
if files, ok := record["documents"].([]interface{}); ok {
    firstFile := files[0].(string)
    url := pb.Files.GetURL(record, firstFile, nil)
}
```

### Image Thumbnails

If your file field has thumbnail sizes configured, you can request thumbnails:

```go
record, err := pb.Collection("example").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

filename := record["avatar"].(string) // Image file

// Get thumbnail with specific size
thumbUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
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
record, err := pb.Collection("products").GetOne("PRODUCT_ID", nil)
if err != nil {
    log.Fatal(err)
}

image := record["image"].(string)

// Different thumbnail sizes
thumbSmall := pb.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "100x100",
})
thumbMedium := pb.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "300x300f",
})
thumbLarge := pb.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "800x600",
})
```

### Force Download

To force browser download instead of preview:

```go
url := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Download: true, // Force download
})
```

## Protected Files

By default, all files are publicly accessible if you know the full URL. For sensitive files, you can mark the field as "Protected" in the collection settings.

### Setting Up Protected Files

```go
collection, err := pb.Collections.GetOne("example", nil)
if err != nil {
    log.Fatal(err)
}

fields, _ := collection["fields"].([]interface{})
for _, f := range fields {
    if field, ok := f.(map[string]interface{}); ok {
        if name, _ := field["name"].(string); name == "documents" {
            field["protected"] = true
            break
        }
    }
}

_, err = pb.Collections.Update("example", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "fields": fields,
    },
})
```

### Accessing Protected Files

Protected files require authentication and a file token:

```go
// Step 1: Authenticate
_, err := pb.Collection("users").AuthWithPassword("user@example.com", "password123", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Step 2: Get file token (valid for ~2 minutes)
fileToken, err := pb.Files.GetToken(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Step 3: Get protected file URL with token
record, err := pb.Collection("example").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

filename := record["privateDocument"].(string)
url := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Token: fileToken,
})

fmt.Println("File URL:", url)
```

**Important:**
- File tokens are short-lived (~2 minutes)
- Only authenticated users satisfying the collection's `viewRule` can access protected files
- Tokens must be regenerated when they expire

### Complete Protected File Example

```go
func loadProtectedImage(pb *bosbase.BosBase, recordID, filename string) (string, error) {
    // Check if authenticated
    if !pb.AuthStore.IsValid() {
        return "", fmt.Errorf("not authenticated")
    }

    // Get fresh token
    token, err := pb.Files.GetToken(nil, nil, nil)
    if err != nil {
        return "", err
    }

    // Get file URL
    record, err := pb.Collection("example").GetOne(recordID, nil)
    if err != nil {
        return "", err
    }

    url := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
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
    "log"
    "os"
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://localhost:8090")
    _, err := pb.Admins.AuthWithPassword("admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Create collection with image field and thumbnails
    collection, err := pb.Collections.CreateBase("products", map[string]interface{}{
        "fields": []map[string]interface{}{
            {"name": "name", "type": "text", "required": true},
            {
                "name":      "image",
                "type":      "file",
                "maxSelect": 1,
                "mimeTypes": []string{"image/jpeg", "image/png"},
                "thumbs":    []string{"100x100", "300x300", "800x600f"},
            },
        },
    }, nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Upload product with image
    file, _ := os.Open("product.jpg")
    defer file.Close()

    product, err := pb.Collection("products").Create(&bosbase.CrudMutateOptions{
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

    // Display thumbnail
    filename := product["image"].(string)
    thumbnailUrl := pb.Files.GetURL(product, filename, &bosbase.FileURLOptions{
        Thumb: "300x300",
    })
    fmt.Println("Thumbnail URL:", thumbnailUrl)
}
```

### Example 2: Multiple File Upload

```go
func uploadMultipleFiles(pb *bosbase.BosBase, files []string) error {
    fileParams := make(map[string]bosbase.FileParam)
    
    for i, filepath := range files {
        file, err := os.Open(filepath)
        if err != nil {
            return err
        }
        defer file.Close()
        
        filename := filepath[strings.LastIndex(filepath, "/")+1:]
        fileParams[fmt.Sprintf("documents_%d", i)] = bosbase.FileParam{
            Filename:    filename,
            Reader:      file,
            ContentType: "application/octet-stream",
        }
    }
    
    _, err := pb.Collection("example").Create(&bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "title": "Document Set",
        },
        Files: fileParams,
    })
    
    return err
}
```

### Example 3: File Management

```go
type FileManager struct {
    pb          *bosbase.BosBase
    collectionID string
    recordID     string
    record       map[string]interface{}
}

func (fm *FileManager) Load() error {
    record, err := fm.pb.Collection(fm.collectionID).GetOne(fm.recordID, nil)
    if err != nil {
        return err
    }
    fm.record = record
    return nil
}

func (fm *FileManager) DeleteFile(filename string) error {
    _, err := fm.pb.Collection(fm.collectionID).Update(fm.recordID, &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "documents-": []string{filename},
        },
    })
    if err != nil {
        return err
    }
    return fm.Load() // Reload
}

func (fm *FileManager) AddFile(filepath string) error {
    file, err := os.Open(filepath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    filename := filepath[strings.LastIndex(filepath, "/")+1:]
    _, err = fm.pb.Collection(fm.collectionID).Update(fm.recordID, &bosbase.CrudMutateOptions{
        Body: map[string]interface{}{
            "documents+": filename,
        },
        Files: map[string]bosbase.FileParam{
            "documents+": {
                Filename:    filename,
                Reader:      file,
                ContentType: "application/octet-stream",
            },
        },
    })
    if err != nil {
        return err
    }
    return fm.Load() // Reload
}
```

## File Field Modifiers

### Summary

- **No modifier** - Replace all files: `documents: [file1, file2]`
- **`+` suffix** - Append files: `documents+: file3`
- **`+` prefix** - Prepend files: `+documents: file0`
- **`-` suffix** - Delete files: `documents-: ['file1.pdf']`

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
file, err := os.Open("test.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

_, err = pb.Collection("example").Create(&bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "title": "Test",
    },
    Files: map[string]bosbase.FileParam{
        "documents": {
            Filename:    "test.txt",
            Reader:      file,
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
            fmt.Println("Upload failed:", err)
        }
    } else {
        fmt.Println("Upload failed:", err)
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

