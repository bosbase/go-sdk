package bosbase

import "net/http"

// FileURLOptions configures file URL generation.
type FileURLOptions struct {
    Thumb    string
    Token    string
    Download bool
    Query    map[string]interface{}
}

type FileService struct {
    BaseService
}

func NewFileService(client *BosBase) *FileService {
    return &FileService{BaseService{client: client}}
}

// GetURL builds the download URL for a specific file.
func (s *FileService) GetURL(record map[string]interface{}, filename string, opts *FileURLOptions) string {
    recordID, _ := record["id"].(string)
    if recordID == "" || filename == "" {
        return ""
    }
    collection := ""
    if v, ok := record["collectionId"].(string); ok {
        collection = v
    }
    if collection == "" {
        if v, ok := record["collectionName"].(string); ok {
            collection = v
        }
    }
    params := map[string]interface{}{}
    if opts != nil {
        for k, v := range opts.Query {
            params[k] = v
        }
        if opts.Thumb != "" {
            params["thumb"] = opts.Thumb
        }
        if opts.Token != "" {
            params["token"] = opts.Token
        }
        if opts.Download {
            params["download"] = ""
        }
    }
    return s.client.BuildURL(
        "/api/files/"+encodePathSegment(collection)+"/"+encodePathSegment(recordID)+"/"+encodePathSegment(filename),
        params,
    )
}

// GetToken requests a temporary file token.
func (s *FileService) GetToken(body map[string]interface{}, query map[string]interface{}, headers map[string]string) (string, error) {
    data, err := s.client.Send("/api/files/token", &RequestOptions{Method: http.MethodPost, Body: body, Query: query, Headers: headers})
    if err != nil {
        return "", err
    }
    if m, ok := data.(map[string]interface{}); ok {
        if tok, ok := m["token"].(string); ok {
            return tok, nil
        }
    }
    return "", nil
}
