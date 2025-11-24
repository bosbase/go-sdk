package bosbase

import "net/http"

type BackupService struct {
    BaseService
}

func NewBackupService(client *BosBase) *BackupService {
    return &BackupService{BaseService{client: client}}
}

func (s *BackupService) GetFullList(query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    data, err := s.client.Send("/api/backups", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    var result []map[string]interface{}
    if arr, ok := data.([]interface{}); ok {
        result = make([]map[string]interface{}, 0, len(arr))
        for _, item := range arr {
            if m, ok := item.(map[string]interface{}); ok {
                result = append(result, m)
            }
        }
    }
    return result, nil
}

func (s *BackupService) Create(name string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["name"] = name
    _, err := s.client.Send("/api/backups", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *BackupService) Upload(files map[string]FileParam, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send("/api/backups/upload", &RequestOptions{Method: http.MethodPost, Body: body, Query: query, Headers: headers, Files: files})
    return err
}

func (s *BackupService) Delete(key string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send("/api/backups/"+encodePathSegment(key), &RequestOptions{Method: http.MethodDelete, Body: body, Query: query, Headers: headers})
    return err
}

func (s *BackupService) Restore(key string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send("/api/backups/"+encodePathSegment(key)+"/restore", &RequestOptions{Method: http.MethodPost, Body: body, Query: query, Headers: headers})
    return err
}

func (s *BackupService) GetDownloadURL(token, key string, query map[string]interface{}) string {
    params := cloneQuery(query)
    params["token"] = token
    return s.client.BuildURL("/api/backups/"+encodePathSegment(key), params)
}
