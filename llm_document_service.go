package bosbase

import "net/http"

type LLMDocumentService struct {
    BaseService
    basePath string
}

func NewLLMDocumentService(client *BosBase) *LLMDocumentService {
    return &LLMDocumentService{BaseService{client: client}, "/api/llm-documents"}
}

func (s *LLMDocumentService) collectionPath(collection string) string {
    return s.basePath + "/" + encodePathSegment(collection)
}

func (s *LLMDocumentService) ListCollections(query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    data, err := s.client.Send(s.basePath+"/collections", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    var result []map[string]interface{}
    if arr, ok := data.([]interface{}); ok {
        for _, item := range arr {
            if m, ok := item.(map[string]interface{}); ok {
                result = append(result, m)
            }
        }
    }
    return result, nil
}

func (s *LLMDocumentService) CreateCollection(name string, metadata map[string]string, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send(s.basePath+"/collections/"+encodePathSegment(name), &RequestOptions{Method: http.MethodPost, Body: map[string]interface{}{"metadata": metadata}, Query: query, Headers: headers})
    return err
}

func (s *LLMDocumentService) DeleteCollection(name string, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send(s.basePath+"/collections/"+encodePathSegment(name), &RequestOptions{Method: http.MethodDelete, Query: query, Headers: headers})
    return err
}

func (s *LLMDocumentService) Insert(collection string, doc LLMDocument, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send(s.collectionPath(collection), &RequestOptions{Method: http.MethodPost, Body: doc.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *LLMDocumentService) Get(collection, documentID string, query map[string]interface{}, headers map[string]string) (LLMDocument, error) {
    data, err := s.client.Send(s.collectionPath(collection)+"/"+encodePathSegment(documentID), &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return LLMDocument{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return LLMDocumentFromMap(m), nil
    }
    return LLMDocument{}, nil
}

func (s *LLMDocumentService) Update(collection, documentID string, doc LLMDocumentUpdate, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send(s.collectionPath(collection)+"/"+encodePathSegment(documentID), &RequestOptions{Method: http.MethodPatch, Body: doc.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *LLMDocumentService) Delete(collection, documentID string, query map[string]interface{}, headers map[string]string) error {
    _, err := s.client.Send(s.collectionPath(collection)+"/"+encodePathSegment(documentID), &RequestOptions{Method: http.MethodDelete, Query: query, Headers: headers})
    return err
}

func (s *LLMDocumentService) List(collection string, page *int, perPage *int, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    params := cloneQuery(query)
    if page != nil {
        params["page"] = *page
    }
    if perPage != nil {
        params["perPage"] = *perPage
    }
    data, err := s.client.Send(s.collectionPath(collection), &RequestOptions{Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *LLMDocumentService) Query(collection string, options LLMQueryOptions, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    path := s.collectionPath(collection) + "/documents/query"
    data, err := s.client.Send(path, &RequestOptions{Method: http.MethodPost, Body: options.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}
