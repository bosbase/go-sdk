package bosbase

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "net/textproto"
    "strings"
    "sync"
    "time"
)

const userAgent = "bosbase-go-sdk/0.1.0"

// FileParam represents a file part in multipart requests.
type FileParam struct {
    Filename    string
    Reader      io.Reader
    ContentType string
}

// RequestOptions describes a generic HTTP call.
type RequestOptions struct {
    Method  string
    Headers map[string]string
    Query   map[string]interface{}
    Body    interface{}
    Files   map[string]FileParam
    Timeout time.Duration
}

// HookOptions passed to BeforeSend allowing mutation.
type HookOptions struct {
    Method  string
    Headers map[string]string
    Body    interface{}
    Query   map[string]interface{}
    Files   map[string]FileParam
    Timeout time.Duration
}

// HookOverride allows overriding the request after BeforeSend.
type HookOverride struct {
    URL     string
    Options *HookOptions
}

// BosBase HTTP client.
type BosBase struct {
    BaseURL   string
    Lang      string
    Timeout   time.Duration
    AuthStore *AuthStore

    BeforeSend func(url string, options *HookOptions) (*HookOverride, error)
    AfterSend  func(resp *http.Response, data interface{}) (interface{}, error)

    httpClient *http.Client
    mu         sync.Mutex
    records    map[string]*RecordService

    Collections *CollectionService
    Files       *FileService
    Logs        *LogService
    Realtime    *RealtimeService
    Settings    *SettingsService
    Health      *HealthService
    Backups     *BackupService
    Crons       *CronService
    Vectors     *VectorService
    LangChaingo *LangChaingoService
    LLMDocuments *LLMDocumentService
    Caches      *CacheService
    GraphQL     *GraphQLService
    PubSub      *PubSubService
}

// ClientOption configures BosBase during construction.
type ClientOption func(*BosBase)

// WithLanguage overrides the default Accept-Language header.
func WithLanguage(lang string) ClientOption {
    return func(c *BosBase) { c.Lang = lang }
}

// WithAuthStore sets a custom auth store implementation.
func WithAuthStore(store *AuthStore) ClientOption {
    return func(c *BosBase) {
        if store != nil {
            c.AuthStore = store
        }
    }
}

// WithTimeout sets the default request timeout.
func WithTimeout(d time.Duration) ClientOption {
    return func(c *BosBase) { c.Timeout = d }
}

// WithHTTPClient allows supplying a preconfigured HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
    return func(c *BosBase) {
        if client != nil {
            c.httpClient = client
        }
    }
}

// New creates a new BosBase client instance.
func New(baseURL string, opts ...ClientOption) *BosBase {
    c := &BosBase{
        BaseURL: strings.TrimRight(baseURL, "/"),
        Lang:    "en-US",
        Timeout: 30 * time.Second,
        AuthStore: NewAuthStore(),
        records: make(map[string]*RecordService),
    }
    for _, opt := range opts {
        opt(c)
    }
    if c.BaseURL == "" {
        c.BaseURL = "/"
    }
    if c.httpClient == nil {
        c.httpClient = &http.Client{Timeout: c.Timeout}
    }
    c.Collections = NewCollectionService(c)
    c.Files = NewFileService(c)
    c.Logs = NewLogService(c)
    c.Realtime = NewRealtimeService(c)
    c.Settings = NewSettingsService(c)
    c.Health = NewHealthService(c)
    c.Backups = NewBackupService(c)
    c.Crons = NewCronService(c)
    c.Vectors = NewVectorService(c)
    c.LangChaingo = NewLangChaingoService(c)
    c.LLMDocuments = NewLLMDocumentService(c)
    c.Caches = NewCacheService(c)
    c.GraphQL = NewGraphQLService(c)
    c.PubSub = NewPubSubService(c)
    return c
}

// Close cleans up open realtime/pubsub connections.
func (c *BosBase) Close() {
    if c.Realtime != nil {
        c.Realtime.Disconnect()
    }
    if c.PubSub != nil {
        c.PubSub.Disconnect()
    }
}

// Collection returns a RecordService scoped to a collection.
func (c *BosBase) Collection(collectionIDOrName string) *RecordService {
    c.mu.Lock()
    defer c.mu.Unlock()
    if svc, ok := c.records[collectionIDOrName]; ok {
        return svc
    }
    svc := NewRecordService(c, collectionIDOrName)
    c.records[collectionIDOrName] = svc
    return svc
}

// Filter interpolates placeholders in filter expressions safely.
func (c *BosBase) Filter(expr string, params map[string]interface{}) string {
    if len(params) == 0 {
        return expr
    }
    result := expr
    for key, val := range params {
        placeholder := "{:" + key + "}"
        switch v := val.(type) {
        case string:
            safe := strings.ReplaceAll(v, "'", "\\'")
            result = strings.ReplaceAll(result, placeholder, "'"+safe+"'")
        case nil:
            result = strings.ReplaceAll(result, placeholder, "null")
        case bool:
            if v {
                result = strings.ReplaceAll(result, placeholder, "true")
            } else {
                result = strings.ReplaceAll(result, placeholder, "false")
            }
        case time.Time:
            ts := v.Format("2006-01-02 15:04:05")
            result = strings.ReplaceAll(result, placeholder, "'"+ts+"'")
        default:
            b, err := json.Marshal(v)
            if err != nil {
                b, _ = json.Marshal(fmt.Sprint(v))
            }
            serialized := string(b)
            serialized = strings.ReplaceAll(serialized, "'", "\\'")
            result = strings.ReplaceAll(result, placeholder, "'"+serialized+"'")
        }
    }
    return result
}

// BuildURL resolves a path against the base host and attaches query parameters.
func (c *BosBase) BuildURL(path string, query map[string]interface{}) string {
    base := c.BaseURL
    if !strings.HasSuffix(base, "/") {
        base += "/"
    }
    rel := strings.TrimLeft(path, "/")
    full := base + rel
    if query != nil {
        values := normalizeQueryParams(query)
        if len(values) > 0 {
            if strings.Contains(full, "?") {
                full += "&" + values.Encode()
            } else {
                full += "?" + values.Encode()
            }
        }
    }
    return full
}

// CreateBatch returns a new batch builder bound to this client.
func (c *BosBase) CreateBatch() *BatchService {
    return NewBatchService(c)
}

// GetFileURL builds the download URL for a record file.
func (c *BosBase) GetFileURL(record map[string]interface{}, filename string, opts *FileURLOptions) string {
    return c.Files.GetURL(record, filename, opts)
}

// Send executes an HTTP request to the BosBase API.
func (c *BosBase) Send(path string, options *RequestOptions) (interface{}, error) {
    if options == nil {
        options = &RequestOptions{}
    }
    method := strings.ToUpper(strings.TrimSpace(options.Method))
    if method == "" {
        method = http.MethodGet
    }

    currentQuery := cloneQuery(options.Query)
    urlStr := c.BuildURL(path, currentQuery)

    headers := map[string]string{
        "Accept-Language": c.Lang,
        "User-Agent":      userAgent,
    }
    for k, v := range options.Headers {
        headers[k] = v
    }
    if _, ok := headers["Authorization"]; !ok && c.AuthStore != nil && c.AuthStore.IsValid() {
        headers["Authorization"] = c.AuthStore.Token()
    }

    files := cloneFiles(options.Files)
    payload := toSerializable(options.Body)
    hookOpts := &HookOptions{
        Method:  method,
        Headers: cloneHeaders(headers),
        Body:    payload,
        Query:   cloneQuery(currentQuery),
        Files:   cloneFiles(files),
        Timeout: options.Timeout,
    }

    if c.BeforeSend != nil {
        override, err := c.BeforeSend(urlStr, hookOpts)
        if err != nil {
            return nil, err
        }
        method = strings.ToUpper(hookOpts.Method)
        headers = cloneHeaders(hookOpts.Headers)
        currentQuery = cloneQuery(hookOpts.Query)
        files = cloneFiles(hookOpts.Files)
        payload = hookOpts.Body
        urlStr = c.BuildURL(path, currentQuery)
        if override != nil {
            if override.URL != "" {
                urlStr = override.URL
            }
            if override.Options != nil {
                if override.Options.Method != "" {
                    method = strings.ToUpper(override.Options.Method)
                }
                if override.Options.Headers != nil {
                    headers = cloneHeaders(override.Options.Headers)
                }
                if override.Options.Query != nil {
                    currentQuery = cloneQuery(override.Options.Query)
                    urlStr = c.BuildURL(path, currentQuery)
                }
                if override.Options.Body != nil {
                    payload = override.Options.Body
                }
                if override.Options.Files != nil {
                    files = cloneFiles(override.Options.Files)
                }
                if override.Options.Timeout > 0 {
                    options.Timeout = override.Options.Timeout
                }
            }
        }
    }

    payload = toSerializable(payload)

    var bodyReader io.Reader
    reqHeaders := make(http.Header)
    for k, v := range headers {
        reqHeaders.Set(k, v)
    }

    if len(files) > 0 {
        buf := &bytes.Buffer{}
        writer := multipart.NewWriter(buf)
        jsonPayload := payload
        if jsonPayload == nil {
            jsonPayload = map[string]interface{}{}
        }
        raw, _ := json.Marshal(jsonPayload)
        _ = writer.WriteField("@jsonPayload", string(raw))
        for field, file := range files {
            partHeaders := textprotoMIMEHeader(field, file)
            part, err := writer.CreatePart(partHeaders)
            if err != nil {
                return nil, err
            }
            if file.Reader != nil {
                if _, err := io.Copy(part, file.Reader); err != nil {
                    return nil, err
                }
            }
        }
        writer.Close()
        bodyReader = buf
        reqHeaders.Set("Content-Type", writer.FormDataContentType())
    } else if payload != nil {
        raw, err := json.Marshal(payload)
        if err != nil {
            return nil, err
        }
        bodyReader = bytes.NewReader(raw)
        reqHeaders.Set("Content-Type", "application/json")
    }

    req, err := http.NewRequest(method, urlStr, bodyReader)
    if err != nil {
        return nil, &ClientResponseError{URL: urlStr, OriginalErr: err}
    }
    req.Header = reqHeaders
    timeout := options.Timeout
    if timeout <= 0 {
        timeout = c.Timeout
    }
    client := c.httpClient
    if client == nil {
        client = &http.Client{Timeout: timeout}
    }
    if timeout > 0 && client.Timeout != timeout {
        clone := *client
        clone.Timeout = timeout
        client = &clone
    }

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    req = req.WithContext(ctx)

    resp, err := client.Do(req)
    if err != nil {
        isAbort := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
        return nil, &ClientResponseError{URL: urlStr, OriginalErr: err, IsAbort: isAbort}
    }
    defer resp.Body.Close()

    var data interface{}
    if resp.StatusCode != http.StatusNoContent {
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, &ClientResponseError{URL: urlStr, OriginalErr: err}
        }
        contentType := resp.Header.Get("Content-Type")
        if strings.Contains(strings.ToLower(contentType), "application/json") {
            if len(body) > 0 {
                var parsed interface{}
                if err := json.Unmarshal(body, &parsed); err == nil {
                    data = parsed
                } else {
                    data = map[string]interface{}{}
                }
            } else {
                data = map[string]interface{}{}
            }
        } else {
            data = body
        }
    }

    if resp.StatusCode >= 400 {
        respMap, _ := data.(map[string]interface{})
        return nil, &ClientResponseError{URL: urlStr, Status: resp.StatusCode, Response: respMap}
    }

    if c.AfterSend != nil {
        var err error
        data, err = c.AfterSend(resp, data)
        if err != nil {
            return nil, err
        }
    }

    return data, nil
}

func cloneHeaders(src map[string]string) map[string]string {
    if src == nil {
        return map[string]string{}
    }
    dst := make(map[string]string, len(src))
    for k, v := range src {
        dst[k] = v
    }
    return dst
}

func cloneQuery(src map[string]interface{}) map[string]interface{} {
    if src == nil {
        return map[string]interface{}{}
    }
    dst := make(map[string]interface{}, len(src))
    for k, v := range src {
        dst[k] = v
    }
    return dst
}

func cloneFiles(src map[string]FileParam) map[string]FileParam {
    if src == nil {
        return map[string]FileParam{}
    }
    dst := make(map[string]FileParam, len(src))
    for k, v := range src {
        dst[k] = v
    }
    return dst
}

func textprotoMIMEHeader(field string, file FileParam) textproto.MIMEHeader {
    header := make(textproto.MIMEHeader)
    disposition := "form-data; name=\"" + field + "\""
    filename := file.Filename
    if filename == "" {
        filename = field
    }
    disposition += "; filename=\"" + filename + "\""
    header.Set("Content-Disposition", disposition)
    ctype := file.ContentType
    if ctype == "" {
        ctype = "application/octet-stream"
    }
    header.Set("Content-Type", ctype)
    return header
}

// ResolveRelative builds a full URL for a path relative to the base.
func (c *BosBase) ResolveRelative(path string) string {
    return c.BuildURL(path, nil)
}

// GetFileToken fetches a temporary file token.
func (c *BosBase) GetFileToken(body map[string]interface{}, query map[string]interface{}) (string, error) {
    data, err := c.Send("/api/files/token", &RequestOptions{Method: http.MethodPost, Body: body, Query: query})
    if err != nil {
        return "", err
    }
    if m, ok := data.(map[string]interface{}); ok {
        if tok, ok := m["token"].(string); ok {
            return tok, nil
        }
    }
    return "", nil
}

// encodeURL builds url with provided path; retained for compatibility.
func (c *BosBase) encodeURL(path string, query map[string]interface{}) string {
    return c.BuildURL(path, query)
}
