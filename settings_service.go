package bosbase

import "net/http"

type SettingsService struct {
    BaseService
}

func NewSettingsService(client *BosBase) *SettingsService {
    return &SettingsService{BaseService{client: client}}
}

func (s *SettingsService) GetAll(query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send("/api/settings", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *SettingsService) Update(body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send("/api/settings", &RequestOptions{Method: http.MethodPatch, Body: body, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *SettingsService) TestS3(filesystem string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    if _, ok := payload["filesystem"]; !ok {
        payload["filesystem"] = filesystem
    }
    _, err := s.client.Send("/api/settings/test/s3", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *SettingsService) TestEmail(toEmail, template string, collection string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["email"] = toEmail
    payload["template"] = template
    if collection != "" {
        payload["collection"] = collection
    }
    _, err := s.client.Send("/api/settings/test/email", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    return err
}

func (s *SettingsService) GenerateAppleClientSecret(clientID, teamID, keyID, privateKey string, duration int, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["clientId"] = clientID
    payload["teamId"] = teamID
    payload["keyId"] = keyID
    payload["privateKey"] = privateKey
    payload["duration"] = duration
    data, err := s.client.Send("/api/settings/apple/generate-client-secret", &RequestOptions{Method: http.MethodPost, Body: payload, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *SettingsService) GetCategory(category string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    settings, err := s.GetAll(query, headers)
    if err != nil {
        return nil, err
    }
    if category == "" {
        return settings, nil
    }
    if val, ok := settings[category].(map[string]interface{}); ok {
        return val, nil
    }
    return nil, nil
}

func (s *SettingsService) UpdateMeta(appName, appURL, senderName, senderAddress string, hideControls *bool, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    meta := map[string]interface{}{}
    if appName != "" {
        meta["appName"] = appName
    }
    if appURL != "" {
        meta["appURL"] = appURL
    }
    if senderName != "" {
        meta["senderName"] = senderName
    }
    if senderAddress != "" {
        meta["senderAddress"] = senderAddress
    }
    if hideControls != nil {
        meta["hideControls"] = *hideControls
    }
    return s.Update(map[string]interface{}{"meta": meta}, query, headers)
}

func (s *SettingsService) GetApplicationSettings(query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    settings, err := s.GetAll(query, headers)
    if err != nil {
        return nil, err
    }
    return map[string]interface{}{
        "meta":         settings["meta"],
        "trustedProxy": settings["trustedProxy"],
        "rateLimits":   settings["rateLimits"],
        "batch":        settings["batch"],
    }, nil
}

func (s *SettingsService) UpdateApplicationSettings(meta, trustedProxy, rateLimits, batch map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    payload := map[string]interface{}{}
    if meta != nil {
        payload["meta"] = meta
    }
    if trustedProxy != nil {
        payload["trustedProxy"] = trustedProxy
    }
    if rateLimits != nil {
        payload["rateLimits"] = rateLimits
    }
    if batch != nil {
        payload["batch"] = batch
    }
    return s.Update(payload, query, headers)
}
