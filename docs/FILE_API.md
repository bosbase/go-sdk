# File API - Go SDK Documentation

## Overview

The File API provides endpoints for downloading and accessing files stored in collection records. It supports thumbnail generation for images, protected file access with tokens, and force download options.

**Key Features:**
- Download files from collection records
- Generate thumbnails for images (crop, fit, resize)
- Protected file access with short-lived tokens
- Force download option for any file type
- Automatic content-type detection
- Support for Range requests and caching

**Backend Endpoints:**
- `GET /api/files/{collection}/{recordId}/{filename}` - Download/fetch file
- `POST /api/files/token` - Generate protected file token

## Download / Fetch File

Downloads a single file resource from a record.

### Basic Usage

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
    
    // Get a record with a file field
    record, err := client.Collection("posts").GetOne("RECORD_ID", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get the file URL
    image, _ := record["image"].(string)
    fileUrl := client.Files.GetURL(record, image, nil)
    
    fmt.Println("File URL:", fileUrl)
}
```

### File URL Structure

The file URL follows this pattern:
```
/api/files/{collectionIdOrName}/{recordId}/{filename}
```

Example:
```
http://127.0.0.1:8090/api/files/posts/abc123/photo_xyz789.jpg
```

## Thumbnails

Generate thumbnails for image files on-the-fly.

### Thumbnail Formats

The following thumbnail formats are supported:

| Format | Example | Description |
|--------|---------|-------------|
| `WxH` | `100x300` | Crop to WxH viewbox (from center) |
| `WxHt` | `100x300t` | Crop to WxH viewbox (from top) |
| `WxHb` | `100x300b` | Crop to WxH viewbox (from bottom) |
| `WxHf` | `100x300f` | Fit inside WxH viewbox (without cropping) |
| `0xH` | `0x300` | Resize to H height preserving aspect ratio |
| `Wx0` | `100x0` | Resize to W width preserving aspect ratio |

### Using Thumbnails

```go
record, err := client.Collection("posts").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

image, _ := record["image"].(string)

// Get thumbnail URL
thumbUrl := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "100x100",
})

// Different thumbnail sizes
smallThumb := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "50x50",
})

mediumThumb := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "200x200",
})

largeThumb := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "500x500",
})

// Fit thumbnail (no cropping)
fitThumb := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "200x200f",
})

// Resize to specific width
widthThumb := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "300x0",
})

// Resize to specific height
heightThumb := client.Files.GetURL(record, image, &bosbase.FileURLOptions{
    Thumb: "0x200",
})
```

### Thumbnail Behavior

- **Image Files Only**: Thumbnails are only generated for image files (PNG, JPG, JPEG, GIF, WEBP)
- **Non-Image Files**: For non-image files, the thumb parameter is ignored and the original file is returned
- **Caching**: Thumbnails are cached and reused if already generated
- **Fallback**: If thumbnail generation fails, the original file is returned
- **Field Configuration**: Thumb sizes must be defined in the file field's `thumbs` option or use default `100x100`

## Protected Files

Protected files require a special token for access, even if you're authenticated.

### Getting a File Token

```go
// Must be authenticated first
_, err := client.Collection("users").AuthWithPassword(
    "user@example.com", "password123", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Get file token
token, err := client.Files.GetToken(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println("File token:", token)
```

### Using Protected File Token

```go
// Get protected file URL with token
record, err := client.Collection("documents").GetOne("RECORD_ID", nil)
if err != nil {
    log.Fatal(err)
}

filename, _ := record["document"].(string)
protectedFileUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Token: token,
})

fmt.Println("Protected file URL:", protectedFileUrl)
```

### Protected File Example

```go
func displayProtectedImage(client *bosbase.BosBase, recordID string) error {
    // Authenticate
    _, err := client.Collection("users").AuthWithPassword(
        "user@example.com", "password123", "", "", nil, nil, nil)
    if err != nil {
        return err
    }
    
    // Get record
    record, err := client.Collection("documents").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    // Get file token
    token, err := client.Files.GetToken(nil, nil, nil)
    if err != nil {
        return err
    }
    
    // Get protected file URL
    filename, _ := record["thumbnail"].(string)
    imageUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
        Thumb: "300x300",
    })
    
    fmt.Println("Protected image URL:", imageUrl)
    return nil
}
```

### Token Lifetime

- File tokens are short-lived (typically expires after a few minutes)
- Tokens are associated with the authenticated user/superuser
- Generate a new token if the previous one expires

## Force Download

Force files to download instead of being displayed in the browser.

```go
// Force download
downloadUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Download: true,
})

fmt.Println("Download URL:", downloadUrl)
```

## Complete Examples

### Example 1: Image Gallery

```go
func displayImageGallery(client *bosbase.BosBase, recordID string) error {
    record, err := client.Collection("posts").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    var images []string
    if imgList, ok := record["images"].([]interface{}); ok {
        for _, img := range imgList {
            if imgStr, ok := img.(string); ok {
                images = append(images, imgStr)
            }
        }
    } else if img, ok := record["image"].(string); ok {
        images = []string{img}
    }
    
    for _, filename := range images {
        // Thumbnail for gallery
        thumbUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
            Thumb: "200x200",
        })
        
        // Full image URL
        fullUrl := client.Files.GetURL(record, filename, nil)
        
        fmt.Printf("Thumbnail: %s\n", thumbUrl)
        fmt.Printf("Full image: %s\n", fullUrl)
    }
    
    return nil
}
```

### Example 2: File Download Handler

```go
func downloadFile(client *bosbase.BosBase, recordID, filename string) (string, error) {
    record, err := client.Collection("documents").GetOne(recordID, nil)
    if err != nil {
        return "", err
    }
    
    // Get download URL
    downloadUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Download: true,
    })
    
    return downloadUrl, nil
}
```

### Example 3: Protected File Viewer

```go
func viewProtectedFile(client *bosbase.BosBase, recordID string) error {
    // Authenticate
    if !client.AuthStore.IsValid() {
        _, err := client.Collection("users").AuthWithPassword(
            "user@example.com", "password123", "", "", nil, nil, nil)
        if err != nil {
            return err
        }
    }
    
    // Get record
    record, err := client.Collection("private_docs").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    // Get token
    token, err := client.Files.GetToken(nil, nil, nil)
    if err != nil {
        return err
    }
    
    // Get file URL
    filename, _ := record["file"].(string)
    fileUrl := client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
    })
    
    fmt.Println("Protected file URL:", fileUrl)
    return nil
}
```

## Error Handling

```go
func safeGetFileURL(client *bosbase.BosBase, record map[string]interface{}, filename string) (string, error) {
    fileUrl := client.Files.GetURL(record, filename, nil)
    
    if fileUrl == "" {
        return "", fmt.Errorf("invalid file URL")
    }
    
    return fileUrl, nil
}

// Protected file token error handling
func getProtectedFileUrl(client *bosbase.BosBase, record map[string]interface{}, filename string) (string, error) {
    // Get token
    token, err := client.Files.GetToken(nil, nil, nil)
    if err != nil {
        if clientErr, ok := err.(*bosbase.ClientResponseError); ok {
            switch clientErr.Status {
            case 401:
                return "", fmt.Errorf("not authenticated")
            case 403:
                return "", fmt.Errorf("no permission to access file")
            default:
                return "", fmt.Errorf("failed to get file token: %v", err)
            }
        }
        return "", err
    }
    
    // Get file URL
    return client.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
    }), nil
}
```

## Best Practices

1. **Use Thumbnails for Lists**: Use thumbnails when displaying images in lists/grids to reduce bandwidth
2. **Cache Tokens**: Store file tokens and reuse them until they expire
3. **Error Handling**: Always handle file loading errors gracefully
4. **Content-Type**: Let the server handle content-type detection automatically
5. **Range Requests**: The API supports Range requests for efficient video/audio streaming
6. **Caching**: Files are cached with a 30-day cache-control header
7. **Security**: Always use tokens for protected files, never expose them in client-side code

## Thumbnail Size Guidelines

| Use Case | Recommended Size |
|----------|-----------------|
| Profile picture | `100x100` or `150x150` |
| List thumbnails | `200x200` or `300x300` |
| Card images | `400x400` or `500x500` |
| Gallery previews | `300x300f` (fit) or `400x400f` |
| Hero images | Use original or `800x800f` |
| Avatar | `50x50` or `75x75` |

## Limitations

- **Thumbnails**: Only work for image files (PNG, JPG, JPEG, GIF, WEBP)
- **Protected Files**: Require authentication to get tokens
- **Token Expiry**: File tokens expire after a short period (typically minutes)
- **File Size**: Large files may take time to generate thumbnails on first request
- **Thumb Sizes**: Must match sizes defined in field configuration or use default `100x100`

## Related Documentation

- [Files Upload and Handling](./FILES.md) - Uploading and managing files
- [API Records](./API_RECORDS.md) - Working with records
- [Collections](./COLLECTIONS.md) - Collection configuration

