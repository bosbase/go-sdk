package bosbase

import "net/http"

type CacheService struct {
    BaseService
}

func NewCacheService(client *BosBase) *CacheService {
    return &CacheService{BaseService{client: client}}
}

func (s *CacheService) List(query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    data, err := s.client.Send("/api/cache", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    // API may return {items: [...]} or an array
    var items []interface{}
    if m, ok := data.(map[string]interface{}); ok {
        if raw, ok := m["items"].([]interface{}); ok {
            items = raw
        }
    }
    if arr, ok := data.([]interface{}); ok {
        items = arr
    }
    result := make([]map[string]interface{}, 0, len(items))
    for _, item := range items {
        if m, ok := item.(map[string]interface{}); ok {
            result = append(result, m)
        }
    }
    return result, nil
}

func (s *CacheService) Create(name string, sizeBytes, defaultTTLSeconds, readTimeoutMs *int, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["name"] = name
    if sizeBytes != nil {
        payload["sizeBytes"] = *sizeBytes
    }
    if defaultTTLSeconds != nil {
        payload["defaultTTLSeconds"] = *defaultTTLSeconds
    }
    if readTimeoutMs != nil {
        payload["readTimeoutMs"] = *readTimeoutMs
    }
    data, err := s.client.Send("/api/cache", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CacheService) Update(name string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send("/api/cache/"+encodePathSegment(name), &RequestOptions{Method: http.MethodPatch, Body: body, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CacheService) Delete(name string, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send("/api/cache/"+encodePathSegment(name), &RequestOptions{Method: http.MethodDelete, Query: query, Headers: headers})
    return err
}

func (s *CacheService) SetEntry(cache, key string, value interface{}, ttlSeconds *int, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["value"] = value
    if ttlSeconds != nil {
        payload["ttlSeconds"] = *ttlSeconds
    }
    path := "/api/cache/" + encodePathSegment(cache) + "/entries/" + encodePathSegment(key)
    data, err := s.client.Send(path, &RequestOptions{Method: http.MethodPut, Body: payload, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CacheService) GetEntry(cache, key string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    path := "/api/cache/" + encodePathSegment(cache) + "/entries/" + encodePathSegment(key)
    data, err := s.client.Send(path, &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CacheService) RenewEntry(cache, key string, ttlSeconds *int, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    if ttlSeconds != nil {
        payload["ttlSeconds"] = *ttlSeconds
    }
    path := "/api/cache/" + encodePathSegment(cache) + "/entries/" + encodePathSegment(key)
    data, err := s.client.Send(path, &RequestOptions{Method: http.MethodPatch, Body: payload, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CacheService) DeleteEntry(cache, key string, query map[string]interface{}, headers map[string]string) error {
    path := "/api/cache/" + encodePathSegment(cache) + "/entries/" + encodePathSegment(key)
    _, err := s.client.Send(path, &RequestOptions{Method: http.MethodDelete, Query: query, Headers: headers})
    return err
}
