# Backups API - Go SDK Documentation

## Overview

The Backups API provides endpoints for managing application data backups. You can create backups, upload existing backup files, download backups, delete backups, and restore the application from a backup.

**Key Features:**
- List all available backup files
- Create new backups with custom names or auto-generated names
- Upload existing backup ZIP files
- Download backup files (requires file token)
- Delete backup files
- Restore the application from a backup (restarts the app)

**Backend Endpoints:**
- `GET /api/backups` - List backups
- `POST /api/backups` - Create backup
- `POST /api/backups/upload` - Upload backup
- `GET /api/backups/{key}` - Download backup
- `DELETE /api/backups/{key}` - Delete backup
- `POST /api/backups/{key}/restore` - Restore backup

**Note**: All Backups API operations require superuser authentication (except download which requires a superuser file token).

## Authentication

All Backups API operations require superuser authentication:

```go
package main

import (
    "log"
    
    bosbase "github.com/bosbase/go-sdk"
)

func main() {
    client := bosbase.New("http://127.0.0.1:8090")
    defer client.Close()
    
    // Authenticate as superuser
    _, err := client.Collection("_superusers").AuthWithPassword(
        "admin@example.com", "password", "", "", nil, nil, nil)
    if err != nil {
        log.Fatal(err)
    }
}
```

**Downloading backups** requires a superuser file token (obtained via `client.Files.GetToken()`), but does not require the Authorization header.

## List Backups

Returns a list of all available backup files with their metadata.

### Basic Usage

```go
// Get all backups
backups, err := client.Backups.GetFullList(nil, nil)
if err != nil {
    log.Fatal(err)
}

for _, backup := range backups {
    key, _ := backup["key"].(string)
    size, _ := backup["size"].(float64)
    modified, _ := backup["modified"].(string)
    fmt.Printf("Backup: %s, Size: %.2f MB, Modified: %s\n", 
        key, size/1024/1024, modified)
}
```

## Create Backup

Creates a new backup of the application data. The backup process is asynchronous and may take some time depending on the size of your data.

### Basic Usage

```go
// Create backup with custom name
err := client.Backups.Create("my_backup_2024.zip", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Create backup with auto-generated name (pass empty string)
err = client.Backups.Create("", nil, nil, nil)
```

### Backup Name Format

Backup names must follow the format: `[a-z0-9_-].zip`
- Only lowercase letters, numbers, underscores, and hyphens
- Must end with `.zip`
- Maximum length: 150 characters
- Must be unique (no existing backup with the same name)

## Upload Backup

Uploads an existing backup ZIP file to the server.

### Basic Usage

```go
file, err := os.Open("backup.zip")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = client.Backups.Upload(map[string]bosbase.FileParam{
    "file": {
        Filename:    "backup.zip",
        Reader:      file,
        ContentType: "application/zip",
    },
}, nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Download Backup

Downloads a backup file. Requires a superuser file token for authentication.

### Basic Usage

```go
// Get file token
token, err := client.Files.GetToken(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

// Build download URL
url := client.Backups.GetDownloadURL(token, "pb_backup_20230519162514.zip", nil)
fmt.Println("Download URL:", url)
```

## Delete Backup

Deletes a backup file from the server.

### Basic Usage

```go
err := client.Backups.Delete("pb_backup_20230519162514.zip", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

## Restore Backup

Restores the application from a backup file. **This operation will restart the application**.

### Basic Usage

```go
err := client.Backups.Restore("pb_backup_20230519162514.zip", nil, nil, nil)
if err != nil {
    log.Fatal(err)
}
```

### Important Warnings

⚠️ **CRITICAL**: Restoring a backup will:
1. Replace all current application data with data from the backup
2. **Restart the application process**
3. Any unsaved changes will be lost
4. The application will be unavailable during the restore process

## Complete Examples

### Example 1: Backup Manager

```go
type BackupManager struct {
    client *bosbase.BosBase
}

func (bm *BackupManager) List() ([]map[string]interface{}, error) {
    return bm.client.Backups.GetFullList(nil, nil)
}

func (bm *BackupManager) Create(name string) error {
    if name == "" {
        name = fmt.Sprintf("backup_%s.zip", time.Now().Format("20060102_150405"))
    }
    return bm.client.Backups.Create(name, nil, nil, nil)
}

func (bm *BackupManager) Download(key string) (string, error) {
    token, err := bm.client.Files.GetToken(nil, nil, nil)
    if err != nil {
        return "", err
    }
    return bm.client.Backups.GetDownloadURL(token, key, nil), nil
}

func (bm *BackupManager) Delete(key string) error {
    return bm.client.Backups.Delete(key, nil, nil, nil)
}

// Usage
manager := &BackupManager{client: client}
backups, _ := manager.List()
err := manager.Create("weekly_backup.zip")
```

## Best Practices

1. **Regular Backups**: Create backups regularly (daily, weekly, or based on your needs)
2. **Naming Convention**: Use clear, consistent naming (e.g., `backup_YYYY-MM-DD.zip`)
3. **Backup Rotation**: Implement cleanup to remove old backups and prevent storage issues
4. **Test Restores**: Periodically test restoring backups to ensure they work
5. **Off-site Storage**: Download and store backups in a separate location

## Related Documentation

- [File API](./FILE_API.md) - File handling and tokens
- [Health API](./HEALTH_API.md) - Check backup readiness

