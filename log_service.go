package bosbase

import (
    "fmt"
    "strings"
)

type LogService struct {
    BaseService
}

func NewLogService(client *BosBase) *LogService {
    return &LogService{BaseService{client: client}}
}

func (s *LogService) GetList(page, perPage int, filter string, sort string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    if page <= 0 {
        page = 1
    }
    if perPage <= 0 {
        perPage = 30
    }
    params := cloneQuery(query)
    params["page"] = page
    params["perPage"] = perPage
    if filter != "" {
        params["filter"] = filter
    }
    if sort != "" {
        params["sort"] = sort
    }
    data, err := s.client.Send("/api/logs", &RequestOptions{Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *LogService) GetOne(logID string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    if strings.TrimSpace(logID) == "" {
        return nil, &ClientResponseError{
            URL:    s.client.BuildURL("/api/logs/", nil),
            Status: 404,
            Response: map[string]interface{}{
                "code":    404,
                "message": "Missing required log id.",
                "data":    map[string]interface{}{},
            },
        }
    }
    data, err := s.client.Send(fmt.Sprintf("/api/logs/%s", logID), &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *LogService) GetStats(query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    data, err := s.client.Send("/api/logs/stats", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if list, ok := data.([]interface{}); ok {
        result := make([]map[string]interface{}, 0, len(list))
        for _, item := range list {
            if m, ok := item.(map[string]interface{}); ok {
                result = append(result, m)
            }
        }
        return result, nil
    }
    return []map[string]interface{}{}, nil
}
