package bosbase

import (
    "errors"
    "fmt"
    "net/http"
    "strings"
)

// BaseService provides access to the shared client.
type BaseService struct {
    client *BosBase
}

// CrudListOptions configures list retrieval.
type CrudListOptions struct {
    Page      int
    PerPage   int
    SkipTotal bool
    Expand    string
    Filter    string
    Sort      string
    Fields    string
    Query     map[string]interface{}
    Headers   map[string]string
}

// CrudViewOptions configures single item retrieval.
type CrudViewOptions struct {
    Expand  string
    Fields  string
    Query   map[string]interface{}
    Headers map[string]string
}

// CrudMutateOptions configures create/update operations.
type CrudMutateOptions struct {
    Expand  string
    Fields  string
    Query   map[string]interface{}
    Headers map[string]string
    Files   map[string]FileParam
    Body    interface{}
}

// CrudDeleteOptions configures delete operations.
type CrudDeleteOptions struct {
    Body    interface{}
    Query   map[string]interface{}
    Headers map[string]string
}

// BaseCrudService implements generic CRUD helpers using a base path.
type BaseCrudService struct {
    BaseService
    path func() string
}

// NewBaseCrudService constructs a CRUD helper with a dynamic path resolver.
func NewBaseCrudService(client *BosBase, path func() string) BaseCrudService {
    return BaseCrudService{BaseService{client: client}, path}
}

func (s *BaseCrudService) basePath() string {
    return s.path()
}

// GetFullList retrieves all records in batches.
func (s *BaseCrudService) GetFullList(batch int, opts *CrudListOptions) ([]interface{}, error) {
    if batch <= 0 {
        return nil, errors.New("batch must be > 0")
    }
    options := opts
    if options == nil {
        options = &CrudListOptions{}
    }
    page := 1
    var result []interface{}
    for {
        options.Page = page
        options.PerPage = batch
        options.SkipTotal = true
        data, err := s.GetList(options)
        if err != nil {
            return nil, err
        }
        items, _ := data["items"].([]interface{})
        result = append(result, items...)
        perPage, _ := data["perPage"].(float64)
        if perPage == 0 {
            perPage = float64(batch)
        }
        if len(items) < int(perPage) {
            break
        }
        page++
    }
    return result, nil
}

// GetList retrieves a paginated list.
func (s *BaseCrudService) GetList(opts *CrudListOptions) (map[string]interface{}, error) {
    options := opts
    if options == nil {
        options = &CrudListOptions{}
    }
    params := cloneQuery(options.Query)
    if options.Page <= 0 {
        options.Page = 1
    }
    if options.PerPage <= 0 {
        options.PerPage = 30
    }
    params["page"] = options.Page
    params["perPage"] = options.PerPage
    if options.SkipTotal {
        params["skipTotal"] = true
    }
    if options.Filter != "" {
        params["filter"] = options.Filter
    }
    if options.Sort != "" {
        params["sort"] = options.Sort
    }
    if options.Expand != "" {
        params["expand"] = options.Expand
    }
    if options.Fields != "" {
        params["fields"] = options.Fields
    }

    data, err := s.client.Send(s.basePath(), &RequestOptions{
        Method:  http.MethodGet,
        Query:   params,
        Headers: options.Headers,
    })
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

// GetOne fetches a single record by id.
func (s *BaseCrudService) GetOne(recordID string, opts *CrudViewOptions) (map[string]interface{}, error) {
    if strings.TrimSpace(recordID) == "" {
        return nil, &ClientResponseError{
            URL:    s.client.BuildURL(fmt.Sprintf("%s/", s.basePath()), nil),
            Status: 404,
            Response: map[string]interface{}{
                "code":    404,
                "message": "Missing required record id.",
                "data":    map[string]interface{}{},
            },
        }
    }
    options := opts
    if options == nil {
        options = &CrudViewOptions{}
    }
    params := cloneQuery(options.Query)
    if options.Expand != "" {
        params["expand"] = options.Expand
    }
    if options.Fields != "" {
        params["fields"] = options.Fields
    }
    encoded := encodePathSegment(recordID)
    data, err := s.client.Send(fmt.Sprintf("%s/%s", s.basePath(), encoded), &RequestOptions{
        Method:  http.MethodGet,
        Query:   params,
        Headers: options.Headers,
    })
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

// GetFirstListItem returns the first record matching the filter.
func (s *BaseCrudService) GetFirstListItem(filter string, opts *CrudViewOptions) (map[string]interface{}, error) {
    options := opts
    if options == nil {
        options = &CrudViewOptions{}
    }
    listOpts := &CrudListOptions{
        Page:    1,
        PerPage: 1,
        SkipTotal: true,
        Filter: filter,
        Expand: options.Expand,
        Fields: options.Fields,
        Query:  options.Query,
        Headers: options.Headers,
    }
    data, err := s.GetList(listOpts)
    if err != nil {
        return nil, err
    }
    items, _ := data["items"].([]interface{})
    if len(items) == 0 {
        return nil, &ClientResponseError{
            Status: 404,
            Response: map[string]interface{}{
                "code":    404,
                "message": "The requested resource wasn't found.",
                "data":    map[string]interface{}{},
            },
        }
    }
    if obj, ok := items[0].(map[string]interface{}); ok {
        return obj, nil
    }
    return map[string]interface{}{}, nil
}

// Create inserts a new record.
func (s *BaseCrudService) Create(opts *CrudMutateOptions) (map[string]interface{}, error) {
    options := opts
    if options == nil {
        options = &CrudMutateOptions{}
    }
    params := cloneQuery(options.Query)
    if options.Expand != "" {
        params["expand"] = options.Expand
    }
    if options.Fields != "" {
        params["fields"] = options.Fields
    }

    data, err := s.client.Send(s.basePath(), &RequestOptions{
        Method:  http.MethodPost,
        Body:    options.Body,
        Query:   params,
        Files:   options.Files,
        Headers: options.Headers,
    })
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

// Update modifies a record.
func (s *BaseCrudService) Update(recordID string, opts *CrudMutateOptions) (map[string]interface{}, error) {
    options := opts
    if options == nil {
        options = &CrudMutateOptions{}
    }
    params := cloneQuery(options.Query)
    if options.Expand != "" {
        params["expand"] = options.Expand
    }
    if options.Fields != "" {
        params["fields"] = options.Fields
    }
    encoded := encodePathSegment(recordID)
    data, err := s.client.Send(fmt.Sprintf("%s/%s", s.basePath(), encoded), &RequestOptions{
        Method:  http.MethodPatch,
        Body:    options.Body,
        Query:   params,
        Files:   options.Files,
        Headers: options.Headers,
    })
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

// Delete removes a record.
func (s *BaseCrudService) Delete(recordID string, opts *CrudDeleteOptions) error {
    options := opts
    if options == nil {
        options = &CrudDeleteOptions{}
    }
    encoded := encodePathSegment(recordID)
    _, err := s.client.Send(fmt.Sprintf("%s/%s", s.basePath(), encoded), &RequestOptions{
        Method:  http.MethodDelete,
        Body:    options.Body,
        Query:   options.Query,
        Headers: options.Headers,
    })
    return err
}
