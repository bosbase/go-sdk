package bosbase

import (
	"errors"
	"net/http"
	"strings"
)

// SQLService provides helpers for executing raw SQL statements (superuser-only).
type SQLService struct {
	BaseService
}

// NewSQLService constructs a SQLService bound to the provided client.
func NewSQLService(client *BosBase) *SQLService {
	return &SQLService{BaseService{client: client}}
}

// Execute runs a SQL statement via the management API and returns the result.
// Only superuser tokens are allowed to call this endpoint.
func (s *SQLService) Execute(query string, queryParams map[string]interface{}, headers map[string]string) (SQLExecuteResponse, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return SQLExecuteResponse{}, errors.New("query is required")
	}
	payload := map[string]interface{}{"query": trimmed}
	data, err := s.client.Send("/api/sql/execute", &RequestOptions{
		Method:  http.MethodPost,
		Body:    payload,
		Query:   queryParams,
		Headers: headers,
	})
	if err != nil {
		return SQLExecuteResponse{}, err
	}
	if m, ok := data.(map[string]interface{}); ok {
		return SQLExecuteResponseFromMap(m), nil
	}
	return SQLExecuteResponse{}, nil
}
