package bosbase

import "net/http"

type LangChaingoService struct {
    BaseService
    basePath string
}

func NewLangChaingoService(client *BosBase) *LangChaingoService {
    return &LangChaingoService{BaseService{client: client}, "/api/langchaingo"}
}

func (s *LangChaingoService) Completions(req LangChaingoCompletionRequest, query map[string]string, headers map[string]string) (LangChaingoCompletionResponse, error) {
    data, err := s.client.Send(s.basePath+"/completions", &RequestOptions{Method: http.MethodPost, Body: req.ToMap(), Query: toAnyMap(query), Headers: headers})
    if err != nil {
        return LangChaingoCompletionResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return LangChaingoCompletionResponseFromMap(m), nil
    }
    return LangChaingoCompletionResponse{}, nil
}

func (s *LangChaingoService) RAG(req LangChaingoRAGRequest, query map[string]string, headers map[string]string) (LangChaingoRAGResponse, error) {
    data, err := s.client.Send(s.basePath+"/rag", &RequestOptions{Method: http.MethodPost, Body: req.ToMap(), Query: toAnyMap(query), Headers: headers})
    if err != nil {
        return LangChaingoRAGResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return LangChaingoRAGResponseFromMap(m), nil
    }
    return LangChaingoRAGResponse{}, nil
}

func (s *LangChaingoService) QueryDocuments(req LangChaingoRAGRequest, query map[string]string, headers map[string]string) (LangChaingoRAGResponse, error) {
    data, err := s.client.Send(s.basePath+"/documents/query", &RequestOptions{Method: http.MethodPost, Body: req.ToMap(), Query: toAnyMap(query), Headers: headers})
    if err != nil {
        return LangChaingoRAGResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return LangChaingoRAGResponseFromMap(m), nil
    }
    return LangChaingoRAGResponse{}, nil
}

func (s *LangChaingoService) SQL(req LangChaingoSQLRequest, query map[string]string, headers map[string]string) (LangChaingoSQLResponse, error) {
    data, err := s.client.Send(s.basePath+"/sql", &RequestOptions{Method: http.MethodPost, Body: req.ToMap(), Query: toAnyMap(query), Headers: headers})
    if err != nil {
        return LangChaingoSQLResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return LangChaingoSQLResponseFromMap(m), nil
    }
    return LangChaingoSQLResponse{}, nil
}

func toAnyMap(input map[string]string) map[string]interface{} {
    if input == nil {
        return nil
    }
    out := make(map[string]interface{}, len(input))
    for k, v := range input {
        out[k] = v
    }
    return out
}
