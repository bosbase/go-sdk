package bosbase

import (
    "net/http"
    "time"
)

type GraphQLService struct {
    BaseService
}

func NewGraphQLService(client *BosBase) *GraphQLService {
    return &GraphQLService{BaseService{client: client}}
}

func (s *GraphQLService) Query(query string, variables map[string]interface{}, operationName string, queryParams map[string]interface{}, headers map[string]string, timeout time.Duration) (map[string]interface{}, error) {
    payload := map[string]interface{}{
        "query":     query,
        "variables": map[string]interface{}{},
    }
    for k, v := range variables {
        payload["variables"].(map[string]interface{})[k] = v
    }
    if operationName != "" {
        payload["operationName"] = operationName
    }
    data, err := s.client.Send("/api/graphql", &RequestOptions{Method: http.MethodPost, Body: payload, Query: queryParams, Headers: headers, Timeout: timeout})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}
