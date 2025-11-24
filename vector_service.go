package bosbase

import "net/http"

type VectorService struct {
    BaseService
    basePath string
}

func NewVectorService(client *BosBase) *VectorService {
    return &VectorService{BaseService{client: client}, "/api/vectors"}
}

func (s *VectorService) collectionPath(collection string) string {
    if collection != "" {
        return s.basePath + "/" + encodePathSegment(collection)
    }
    return s.basePath
}

func (s *VectorService) Insert(doc VectorDocument, collection string, query map[string]interface{}, headers map[string]string) (VectorInsertResponse, error) {
    data, err := s.client.Send(s.collectionPath(collection), &RequestOptions{Method: http.MethodPost, Body: doc.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return VectorInsertResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return VectorInsertResponseFromMap(m), nil
    }
    return VectorInsertResponse{}, nil
}

func (s *VectorService) BatchInsert(opts VectorBatchInsertOptions, collection string, query map[string]interface{}, headers map[string]string) (VectorBatchInsertResponse, error) {
    data, err := s.client.Send(s.collectionPath(collection)+"/documents/batch", &RequestOptions{Method: http.MethodPost, Body: opts.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return VectorBatchInsertResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return VectorBatchInsertResponseFromMap(m), nil
    }
    return VectorBatchInsertResponse{}, nil
}

func (s *VectorService) Update(documentID string, doc VectorDocument, collection string, query map[string]interface{}, headers map[string]string) (VectorInsertResponse, error) {
    path := s.collectionPath(collection) + "/" + encodePathSegment(documentID)
    data, err := s.client.Send(path, &RequestOptions{Method: http.MethodPatch, Body: doc.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return VectorInsertResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return VectorInsertResponseFromMap(m), nil
    }
    return VectorInsertResponse{}, nil
}

func (s *VectorService) Delete(documentID string, collection string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    path := s.collectionPath(collection) + "/" + encodePathSegment(documentID)
    _, err := s.client.Send(path, &RequestOptions{Method: http.MethodDelete, Body: body, Query: query, Headers: headers})
    return err
}

func (s *VectorService) Search(options VectorSearchOptions, collection string, query map[string]interface{}, headers map[string]string) (VectorSearchResponse, error) {
    data, err := s.client.Send(s.collectionPath(collection)+"/documents/search", &RequestOptions{Method: http.MethodPost, Body: options.ToMap(), Query: query, Headers: headers})
    if err != nil {
        return VectorSearchResponse{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return VectorSearchResponseFromMap(m), nil
    }
    return VectorSearchResponse{}, nil
}

func (s *VectorService) Get(documentID string, collection string, query map[string]interface{}, headers map[string]string) (VectorDocument, error) {
    path := s.collectionPath(collection) + "/" + encodePathSegment(documentID)
    data, err := s.client.Send(path, &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return VectorDocument{}, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return VectorDocumentFromMap(m), nil
    }
    return VectorDocument{}, nil
}

func (s *VectorService) List(collection string, page *int, perPage *int, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
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

func (s *VectorService) CreateCollection(name string, config VectorCollectionConfig, query map[string]interface{}, headers map[string]string) error {
    path := s.basePath + "/collections/" + encodePathSegment(name)
    _, err := s.client.Send(path, &RequestOptions{Method: http.MethodPost, Body: config.ToMap(), Query: query, Headers: headers})
    return err
}

func (s *VectorService) UpdateCollection(name string, config VectorCollectionConfig, query map[string]interface{}, headers map[string]string) error {
    path := s.basePath + "/collections/" + encodePathSegment(name)
    _, err := s.client.Send(path, &RequestOptions{Method: http.MethodPatch, Body: config.ToMap(), Query: query, Headers: headers})
    return err
}

func (s *VectorService) DeleteCollection(name string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    path := s.basePath + "/collections/" + encodePathSegment(name)
    _, err := s.client.Send(path, &RequestOptions{Method: http.MethodDelete, Body: body, Query: query, Headers: headers})
    return err
}

func (s *VectorService) ListCollections(query map[string]interface{}, headers map[string]string) ([]VectorCollectionInfo, error) {
    data, err := s.client.Send(s.basePath+"/collections", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    var result []VectorCollectionInfo
    if arr, ok := data.([]interface{}); ok {
        for _, item := range arr {
            if m, ok := item.(map[string]interface{}); ok {
                result = append(result, VectorCollectionInfoFromMap(m))
            }
        }
    }
    return result, nil
}
