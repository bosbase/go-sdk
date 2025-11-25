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
    "github.com/your-org/bosbase-go-sdk"
)

func main() {
    pb := bosbase.New("http://127.0.0.1:8090")
    
    // Get a record with a file field
    record, err := pb.Collection("posts").GetOne("RECORD_ID", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get the file URL
    filename := record["image"].(string)
    fileUrl := pb.Files.GetURL(record, filename, nil)
    
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
// Get thumbnail URL
thumbUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "100x100",
})

// Different thumbnail sizes
smallThumb := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "50x50",
})

mediumThumb := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "200x200",
})

largeThumb := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "500x500",
})

// Fit thumbnail (no cropping)
fitThumb := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "200x200f",
})

// Resize to specific width
widthThumb := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Thumb: "300x0",
})

// Resize to specific height
heightThumb := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
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
_, err := pb.Collection("users").AuthWithPassword("user@example.com", "password", "", "", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Get file token
token, err := pb.Files.GetToken(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Token:", token)
```

### Using Protected File Token

```go
// Get protected file URL with token
protectedFileUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Token: token,
})

// Access the file using HTTP client
resp, err := http.Get(protectedFileUrl)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// Read file content
body, err := io.ReadAll(resp.Body)
if err != nil {
    log.Fatal(err)
}
```

### Protected File Example

```go
func displayProtectedImage(pb *bosbase.BosBase, recordID string) error {
    // Authenticate
    _, err := pb.Collection("users").AuthWithPassword("user@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        return err
    }
    
    // Get record
    record, err := pb.Collection("documents").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    // Get file token
    token, err := pb.Files.GetToken(nil, nil, nil)
    if err != nil {
        return err
    }
    
    // Get protected file URL
    filename := record["thumbnail"].(string)
    imageUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
        Thumb: "300x300",
    })
    
    fmt.Println("Image URL:", imageUrl)
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
downloadUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
    Download: true,
})

// Use in HTTP client to download
resp, err := http.Get(downloadUrl)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// Save to file
out, err := os.Create("downloaded_file.pdf")
if err != nil {
    log.Fatal(err)
}
defer out.Close()

io.Copy(out, resp.Body)
```

## Complete Examples

### Example 1: Image Gallery

```go
func displayImageGallery(pb *bosbase.BosBase, recordID string) error {
    record, err := pb.Collection("posts").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    var images []string
    if img, ok := record["images"].([]interface{}); ok {
        for _, i := range img {
            images = append(images, i.(string))
        }
    } else if img, ok := record["image"].(string); ok {
        images = []string{img}
    }
    
    for _, filename := range images {
        // Thumbnail for gallery
        thumbUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
            Thumb: "200x200",
        })
        
        // Full image URL
        fullUrl := pb.Files.GetURL(record, filename, nil)
        
        fmt.Printf("Thumbnail: %s\n", thumbUrl)
        fmt.Printf("Full: %s\n", fullUrl)
    }
    
    return nil
}
```

### Example 2: File Download Handler

```go
func downloadFile(pb *bosbase.BosBase, recordID, filename string) error {
    record, err := pb.Collection("documents").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    // Get download URL
    downloadUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Download: true,
    })
    
    // Download file
    resp, err := http.Get(downloadUrl)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // Save to local file
    out, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer out.Close()
    
    _, err = io.Copy(out, resp.Body)
    return err
}
```

### Example 3: Protected File Viewer

```go
func viewProtectedFile(pb *bosbase.BosBase, recordID string) error {
    // Authenticate
    if !pb.AuthStore.IsValid() {
        _, err := pb.Collection("users").AuthWithPassword("user@example.com", "password", "", "", nil, nil, nil)
        if err != nil {
            return err
        }
    }
    
    // Get record
    record, err := pb.Collection("private_docs").GetOne(recordID, nil)
    if err != nil {
        return err
    }
    
    // Get token
    token, err := pb.Files.GetToken(nil, nil, nil)
    if err != nil {
        return fmt.Errorf("failed to get file token: %v", err)
    }
    
    // Get file URL
    filename := record["file"].(string)
    fileUrl := pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
    })
    
    fmt.Println("File URL:", fileUrl)
    return nil
}
```

## Error Handling

```go
fileUrl := pb.Files.GetURL(record, filename, nil)

// Verify URL is valid
if fileUrl == "" {
    log.Fatal("Invalid file URL")
}

// Load file using HTTP client
resp, err := http.Get(fileUrl)
if err != nil {
    log.Fatal("Failed to load file:", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
    log.Fatal("File access error:", resp.StatusCode)
}
```

### Protected File Token Error Handling

```go
func getProtectedFileUrl(pb *bosbase.BosBase, record map[string]interface{}, filename string) (string, error) {
    // Get token
    token, err := pb.Files.GetToken(nil, nil, nil)
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
    return pb.Files.GetURL(record, filename, &bosbase.FileURLOptions{
        Token: token,
    }), nil
}
```

## Best Practices

1. **Use Thumbnails for Lists**: Use thumbnails when displaying images in lists/grids to reduce bandwidth
2. **Lazy Loading**: Implement lazy loading for images below the fold
3. **Cache Tokens**: Store file tokens and reuse them until they expire
4. **Error Handling**: Always handle file loading errors gracefully
5. **Content-Type**: Let the server handle content-type detection automatically
6. **Range Requests**: The API supports Range requests for efficient video/audio streaming
7. **Caching**: Files are cached with a 30-day cache-control header
8. **Security**: Always use tokens for protected files, never expose them in client-side code

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

