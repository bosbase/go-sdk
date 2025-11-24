package bosbase

import "net/http"

type CronService struct {
    BaseService
}

func NewCronService(client *BosBase) *CronService {
    return &CronService{BaseService{client: client}}
}

func (s *CronService) GetFullList(query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    data, err := s.client.Send("/api/crons", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    result := []map[string]interface{}{}
    if arr, ok := data.([]interface{}); ok {
        for _, item := range arr {
            if m, ok := item.(map[string]interface{}); ok {
                result = append(result, m)
            }
        }
    }
    return result, nil
}

func (s *CronService) Run(jobID string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    path := "/api/crons/" + encodePathSegment(jobID)
    _, err := s.client.Send(path, &RequestOptions{Method: http.MethodPost, Body: body, Query: query, Headers: headers})
    return err
}
