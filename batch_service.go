package bosbase

import (
    "fmt"
    "net/http"
)

type batchRequest struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    interface{}
    Files   map[string]FileParam
}

type BatchService struct {
    BaseService
    requests    []batchRequest
    collections map[string]*SubBatchService
}

func NewBatchService(client *BosBase) *BatchService {
    return &BatchService{BaseService{client: client}, []batchRequest{}, map[string]*SubBatchService{}}
}

func (b *BatchService) Collection(collection string) *SubBatchService {
    if svc, ok := b.collections[collection]; ok {
        return svc
    }
    svc := &SubBatchService{batch: b, collection: collection}
    b.collections[collection] = svc
    return svc
}

func (b *BatchService) queueRequest(method, url string, headers map[string]string, body interface{}, files map[string]FileParam) {
    b.requests = append(b.requests, batchRequest{
        Method:  method,
        URL:     url,
        Headers: cloneHeaders(headers),
        Body:    toSerializable(body),
        Files:   cloneFiles(files),
    })
}

func (b *BatchService) Send(body map[string]interface{}, query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    requestsPayload := make([]map[string]interface{}, 0, len(b.requests))
    attachments := map[string]FileParam{}

    for idx, req := range b.requests {
        requestsPayload = append(requestsPayload, map[string]interface{}{
            "method": req.Method,
            "url":    req.URL,
            "headers": req.Headers,
            "body":   req.Body,
        })
        for field, file := range req.Files {
            attachments[fmt.Sprintf("requests.%d.%s", idx, field)] = file
        }
    }

    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["requests"] = requestsPayload

    data, err := b.client.Send("/api/batch", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers, Files: attachments})
    b.requests = nil
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

type SubBatchService struct {
    batch      *BatchService
    collection string
}

func (s *SubBatchService) collectionURL() string {
    return "/api/collections/" + encodePathSegment(s.collection) + "/records"
}

func (s *SubBatchService) Create(body interface{}, query map[string]interface{}, files map[string]FileParam, headers map[string]string, expand string, fields string) {
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    url := buildRelativeURL(s.collectionURL(), params)
    s.batch.queueRequest(http.MethodPost, url, headers, body, files)
}

func (s *SubBatchService) Upsert(body interface{}, query map[string]interface{}, files map[string]FileParam, headers map[string]string, expand string, fields string) {
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    url := buildRelativeURL(s.collectionURL(), params)
    s.batch.queueRequest(http.MethodPut, url, headers, body, files)
}

func (s *SubBatchService) Update(recordID string, body interface{}, query map[string]interface{}, files map[string]FileParam, headers map[string]string, expand string, fields string) {
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    url := buildRelativeURL(s.collectionURL()+"/"+encodePathSegment(recordID), params)
    s.batch.queueRequest(http.MethodPatch, url, headers, body, files)
}

func (s *SubBatchService) Delete(recordID string, body interface{}, query map[string]interface{}, headers map[string]string) {
    url := buildRelativeURL(s.collectionURL()+"/"+encodePathSegment(recordID), query)
    s.batch.queueRequest(http.MethodDelete, url, headers, body, nil)
}
