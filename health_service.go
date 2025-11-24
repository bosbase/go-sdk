package bosbase

// HealthService exposes health checks.
type HealthService struct {
    BaseService
}

func NewHealthService(client *BosBase) *HealthService {
    return &HealthService{BaseService{client: client}}
}

func (s *HealthService) Check(query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send("/api/health", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}
