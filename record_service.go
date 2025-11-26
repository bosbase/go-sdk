package bosbase

import (
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
)

type RecordService struct {
    BaseCrudService
    collection string
}

func NewRecordService(client *BosBase, collection string) *RecordService {
    svc := &RecordService{collection: collection}
    svc.BaseCrudService = NewBaseCrudService(client, func() string {
        return fmt.Sprintf("/api/collections/%s/records", encodePathSegment(collection))
    })
    return svc
}

func (s *RecordService) baseCollectionPath() string {
    return "/api/collections/" + encodePathSegment(s.collection)
}

func (s *RecordService) Subscribe(topic string, callback func(map[string]interface{}), query map[string]interface{}, headers map[string]string) (func(), error) {
    if topic == "" {
        return nil, errors.New("topic must be set")
    }
    if callback == nil {
        return nil, errors.New("callback must be set")
    }
    fullTopic := s.collection + "/" + topic
    return s.client.Realtime.Subscribe(fullTopic, callback, query, headers)
}

func (s *RecordService) Unsubscribe(topic string) {
    if topic != "" {
        s.client.Realtime.Unsubscribe(s.collection + "/" + topic)
    } else {
        s.client.Realtime.UnsubscribeByPrefix(s.collection)
    }
}

func (s *RecordService) Update(recordID string, opts *CrudMutateOptions) (map[string]interface{}, error) {
    item, err := s.BaseCrudService.Update(recordID, opts)
    if err != nil {
        return nil, err
    }
    s.maybeUpdateAuthRecord(item)
    return item, nil
}

func (s *RecordService) Delete(recordID string, opts *CrudDeleteOptions) error {
    if err := s.BaseCrudService.Delete(recordID, opts); err != nil {
        return err
    }
    if s.isAuthRecord(recordID) {
        s.client.AuthStore.Clear()
    }
    return nil
}

func (s *RecordService) GetCount(filter, expand, fields string, query map[string]interface{}, headers map[string]string) (int, error) {
    params := cloneQuery(query)
    if filter != "" {
        params["filter"] = filter
    }
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    data, err := s.client.Send(s.basePath()+"/count", &RequestOptions{Query: params, Headers: headers})
    if err != nil {
        return 0, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        if count, ok := m["count"].(float64); ok {
            return int(count), nil
        }
    }
    return 0, nil
}

func (s *RecordService) ListAuthMethods(fields string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    params := cloneQuery(query)
    if fields == "" {
        fields = "mfa,otp,password,oauth2"
    }
    params["fields"] = fields
    data, err := s.client.Send(s.baseCollectionPath()+"/auth-methods", &RequestOptions{Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *RecordService) AuthWithPassword(identity, password, expand, fields string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["identity"] = identity
    payload["password"] = password
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    data, err := s.client.Send(s.baseCollectionPath()+"/auth-with-password", &RequestOptions{Method: http.MethodPost, Body: payload, Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return s.authResponse(m), nil
    }
    return map[string]interface{}{}, nil
}

func (s *RecordService) AuthWithOAuth2Code(provider, code, codeVerifier, redirectURL string, createData map[string]interface{}, body map[string]interface{}, query map[string]interface{}, headers map[string]string, expand, fields string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["provider"] = provider
    payload["code"] = code
    payload["codeVerifier"] = codeVerifier
    payload["redirectURL"] = redirectURL
    if createData != nil {
        payload["createData"] = createData
    }
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    data, err := s.client.Send(s.baseCollectionPath()+"/auth-with-oauth2", &RequestOptions{Method: http.MethodPost, Body: payload, Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return s.authResponse(m), nil
    }
    return map[string]interface{}{}, nil
}

func (s *RecordService) AuthWithOAuth2(providerName string, urlCallback func(string), scopes []string, createData map[string]interface{}, body map[string]interface{}, query map[string]interface{}, headers map[string]string, expand, fields string, timeout time.Duration) (map[string]interface{}, error) {
    methods, err := s.ListAuthMethods("", nil, nil)
    if err != nil {
        return nil, err
    }
    providers := []interface{}{}
    if oauth, ok := methods["oauth2"].(map[string]interface{}); ok {
        if prov, ok := oauth["providers"].([]interface{}); ok {
            providers = prov
        }
    }
    var provider map[string]interface{}
    for _, item := range providers {
        if m, ok := item.(map[string]interface{}); ok {
            if name, _ := m["name"].(string); name == providerName {
                provider = m
                break
            }
        }
    }
    if provider == nil {
        return nil, &ClientResponseError{Response: map[string]interface{}{"message": fmt.Sprintf("missing provider %s", providerName)}}
    }

    redirectURL := s.client.BuildURL("/api/oauth2-redirect", nil)

    resultChan := make(chan map[string]interface{}, 1)
    errChan := make(chan error, 1)

    handler := func(payload map[string]interface{}) {
        state := fmt.Sprint(payload["state"])
        if state != s.client.Realtime.ClientID {
            return
        }
        if errMsg, ok := payload["error"].(string); ok && errMsg != "" {
            errChan <- &ClientResponseError{Response: map[string]interface{}{"message": errMsg}}
            return
        }
        code := fmt.Sprint(payload["code"])
        if code == "" {
            errChan <- &ClientResponseError{Response: map[string]interface{}{"message": "OAuth2 redirect missing code"}}
            return
        }
        auth, err := s.AuthWithOAuth2Code(providerName, code, fmt.Sprint(provider["codeVerifier"]), redirectURL, createData, body, query, headers, expand, fields)
        if err != nil {
            errChan <- err
            return
        }
        resultChan <- auth
    }

    unsubscribe, subErr := s.client.Realtime.Subscribe("@oauth2", handler, nil, nil)
    if subErr != nil {
        return nil, subErr
    }
    defer unsubscribe()

    if err := s.client.Realtime.EnsureConnected(10 * time.Second); err != nil {
        return nil, err
    }
    state := s.client.Realtime.ClientID

    authURL := fmt.Sprint(provider["authURL"]) + redirectURL
    parsed, _ := url.Parse(authURL)
    q := parsed.Query()
    q.Set("state", state)
    if len(scopes) > 0 {
        q.Set("scope", strings.Join(scopes, " "))
    }
    parsed.RawQuery = q.Encode()
    urlCallback(parsed.String())

    if timeout <= 0 {
        timeout = 180 * time.Second
    }
    select {
    case res := <-resultChan:
        return res, nil
    case err := <-errChan:
        return nil, err
    case <-time.After(timeout):
        return nil, &ClientResponseError{Response: map[string]interface{}{"message": "OAuth2 flow timed out"}}
    }
}

func (s *RecordService) AuthRefresh(expand, fields string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    data, err := s.client.Send(s.baseCollectionPath()+"/auth-refresh", &RequestOptions{Method: http.MethodPost, Body: body, Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return s.authResponse(m), nil
    }
    return map[string]interface{}{}, nil
}

func (s *RecordService) RequestPasswordReset(email string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["email"] = email
    _, err := s.client.Send(s.baseCollectionPath()+"/request-password-reset", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *RecordService) ConfirmPasswordReset(token, password, passwordConfirm string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["token"] = token
    payload["password"] = password
    payload["passwordConfirm"] = passwordConfirm
    _, err := s.client.Send(s.baseCollectionPath()+"/confirm-password-reset", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *RecordService) RequestVerification(email string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["email"] = email
    _, err := s.client.Send(s.baseCollectionPath()+"/request-verification", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *RecordService) ConfirmVerification(token string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["token"] = token
    _, err := s.client.Send(s.baseCollectionPath()+"/confirm-verification", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    if err == nil {
        s.markVerified(token)
    }
    return err
}

func (s *RecordService) RequestEmailChange(newEmail string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["newEmail"] = newEmail
    _, err := s.client.Send(s.baseCollectionPath()+"/request-email-change", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *RecordService) ConfirmEmailChange(token, password string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["token"] = token
    payload["password"] = password
    _, err := s.client.Send(s.baseCollectionPath()+"/confirm-email-change", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    if err == nil {
        s.clearIfSameToken(token)
    }
    return err
}

func (s *RecordService) RequestOTP(email string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["email"] = email
    data, err := s.client.Send(s.baseCollectionPath()+"/request-otp", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *RecordService) AuthWithOTP(otpID, password, expand, fields string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["otpId"] = otpID
    payload["password"] = password
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    data, err := s.client.Send(s.baseCollectionPath()+"/auth-with-otp", &RequestOptions{Method: http.MethodPost, Body: payload, Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return s.authResponse(m), nil
    }
    return map[string]interface{}{}, nil
}

// BindCustomToken binds a custom token to an auth record after verifying the email and password.
func (s *RecordService) BindCustomToken(email, password, token string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["email"] = email
    payload["password"] = password
    payload["token"] = token
    _, err := s.client.Send(s.baseCollectionPath()+"/bind-token", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

// UnbindCustomToken removes a previously bound custom token after verifying the email and password.
func (s *RecordService) UnbindCustomToken(email, password, token string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["email"] = email
    payload["password"] = password
    payload["token"] = token
    _, err := s.client.Send(s.baseCollectionPath()+"/unbind-token", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

// AuthWithToken authenticates an auth collection record using a previously bound custom token.
// On success, this method also automatically updates the client's AuthStore data.
func (s *RecordService) AuthWithToken(token, expand, fields string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["token"] = token
    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }
    data, err := s.client.Send(s.baseCollectionPath()+"/auth-with-token", &RequestOptions{Method: http.MethodPost, Body: payload, Query: params, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return s.authResponse(m), nil
    }
    return map[string]interface{}{}, nil
}

// ListExternalAuths lists all linked external auth providers for the specified auth record.
func (s *RecordService) ListExternalAuths(recordID string, query map[string]interface{}, headers map[string]string) ([]map[string]interface{}, error) {
    filter := s.client.Filter("recordRef = {:id}", map[string]interface{}{"id": recordID})
    params := cloneQuery(query)
    if params == nil {
        params = map[string]interface{}{}
    }
    params["filter"] = filter
    data, err := s.client.Collection("_externalAuths").GetFullList(500, &CrudListOptions{
        Filter:  filter,
        Query:   params,
        Headers: headers,
    })
    if err != nil {
        return nil, err
    }
    result := make([]map[string]interface{}, 0, len(data))
    for _, item := range data {
        if m, ok := item.(map[string]interface{}); ok {
            result = append(result, m)
        }
    }
    return result, nil
}

func (s *RecordService) Impersonate(recordID string, duration int, expand, fields string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (*BosBase, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["duration"] = duration

    params := cloneQuery(query)
    if expand != "" {
        params["expand"] = expand
    }
    if fields != "" {
        params["fields"] = fields
    }

    enrichedHeaders := cloneHeaders(headers)
    if s.client.AuthStore != nil && s.client.AuthStore.IsValid() {
        enrichedHeaders["Authorization"] = s.client.AuthStore.Token()
    }

    newClient := New(s.client.BaseURL, WithLanguage(s.client.Lang))
    data, err := newClient.Send(fmt.Sprintf("%s/impersonate/%s", s.baseCollectionPath(), encodePathSegment(recordID)), &RequestOptions{Method: http.MethodPost, Body: payload, Query: params, Headers: enrichedHeaders})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        token, _ := m["token"].(string)
        record, _ := m["record"].(map[string]interface{})
        if token != "" && record != nil {
            newClient.AuthStore.Save(token, record)
        }
    }
    return newClient, nil
}

func (s *RecordService) authResponse(data map[string]interface{}) map[string]interface{} {
    token, _ := data["token"].(string)
    record, _ := data["record"].(map[string]interface{})
    if token != "" && record != nil {
        s.client.AuthStore.Save(token, record)
    }
    return data
}

func (s *RecordService) maybeUpdateAuthRecord(item map[string]interface{}) {
    current := s.client.AuthStore.Record()
    if current == nil {
        return
    }
    if currentID, _ := current["id"].(string); currentID != item["id"] {
        return
    }
    cid := fmt.Sprint(current["collectionId"])
    cname := fmt.Sprint(current["collectionName"])
    if cid != s.collection && cname != s.collection {
        return
    }
    merged := cloneQuery(current)
    for k, v := range item {
        merged[k] = v
    }
    if expCur, ok := current["expand"].(map[string]interface{}); ok {
        if expNew, ok2 := item["expand"].(map[string]interface{}); ok2 {
            mergedExp := cloneQuery(expCur)
            for k, v := range expNew {
                mergedExp[k] = v
            }
            merged["expand"] = mergedExp
        }
    }
    s.client.AuthStore.Save(s.client.AuthStore.Token(), merged)
}

func (s *RecordService) isAuthRecord(recordID string) bool {
    current := s.client.AuthStore.Record()
    if current == nil {
        return false
    }
    cid := fmt.Sprint(current["collectionId"])
    return fmt.Sprint(current["id"]) == recordID && (cid == s.collection || fmt.Sprint(current["collectionName"]) == s.collection)
}

func (s *RecordService) markVerified(token string) {
    current := s.client.AuthStore.Record()
    if current == nil {
        return
    }
    payload := decodeTokenPayload(token)
    if payload == nil {
        return
    }
    if fmt.Sprint(current["id"]) == fmt.Sprint(payload["id"]) && fmt.Sprint(current["collectionId"]) == fmt.Sprint(payload["collectionId"]) {
        if verified, _ := current["verified"].(bool); !verified {
            current["verified"] = true
            s.client.AuthStore.Save(s.client.AuthStore.Token(), current)
        }
    }
}

func (s *RecordService) clearIfSameToken(token string) {
    current := s.client.AuthStore.Record()
    if current == nil {
        return
    }
    payload := decodeTokenPayload(token)
    if payload == nil {
        return
    }
    if fmt.Sprint(current["id"]) == fmt.Sprint(payload["id"]) && fmt.Sprint(current["collectionId"]) == fmt.Sprint(payload["collectionId"]) {
        s.client.AuthStore.Clear()
    }
}

func decodeTokenPayload(token string) map[string]interface{} {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return nil
    }
    payloadPart := parts[1]
    padding := len(payloadPart) % 4
    if padding > 0 {
        payloadPart += strings.Repeat("=", 4-padding)
    }
    decoded, err := base64.URLEncoding.DecodeString(payloadPart)
    if err != nil {
        return nil
    }
    var payload map[string]interface{}
    if err := json.Unmarshal(decoded, &payload); err != nil {
        return nil
    }
    return payload
}
