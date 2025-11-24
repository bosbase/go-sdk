package bosbase

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "strings"
    "sync"
    "time"
)

type AuthListener func(token string, record map[string]interface{})

// AuthStore keeps token and auth record in memory.
type AuthStore struct {
    mu        sync.RWMutex
    token     string
    record    map[string]interface{}
    listeners map[string]AuthListener
    nextID    int64
}

func NewAuthStore() *AuthStore {
    return &AuthStore{listeners: make(map[string]AuthListener)}
}

func (s *AuthStore) Token() string {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.token
}

func (s *AuthStore) Record() map[string]interface{} {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if s.record == nil {
        return nil
    }
    clone := make(map[string]interface{}, len(s.record))
    for k, v := range s.record {
        clone[k] = v
    }
    return clone
}

// IsValid returns true when a non-expired JWT token is stored.
func (s *AuthStore) IsValid() bool {
    s.mu.RLock()
    token := s.token
    s.mu.RUnlock()

    if token == "" {
        return false
    }
    parts := splitToken(token)
    if len(parts) != 3 {
        return false
    }

    payloadPart := parts[1]
    padding := len(payloadPart) % 4
    if padding > 0 {
        payloadPart += strings.Repeat("=", 4-padding)
    }
    decoded, err := base64.RawURLEncoding.DecodeString(payloadPart)
    if err != nil {
        return false
    }

    var payload map[string]interface{}
    if err := json.Unmarshal(decoded, &payload); err != nil {
        return false
    }

    expVal, ok := payload["exp"]
    if !ok {
        return false
    }
    expFloat, ok := expVal.(float64)
    if !ok {
        return false
    }
    return int64(expFloat) > time.Now().Unix()
}

func (s *AuthStore) AddListener(fn AuthListener) string {
    s.mu.Lock()
    s.nextID++
    id := s.nextID
    key := fmt.Sprintf("listener-%d", id)
    s.listeners[key] = fn
    s.mu.Unlock()
    return key
}

func (s *AuthStore) RemoveListener(id string) {
    s.mu.Lock()
    delete(s.listeners, id)
    s.mu.Unlock()
}

func (s *AuthStore) Save(token string, record map[string]interface{}) {
    s.mu.Lock()
    s.token = token
    s.record = record
    listeners := make([]AuthListener, 0, len(s.listeners))
    for _, fn := range s.listeners {
        listeners = append(listeners, fn)
    }
    s.mu.Unlock()

    for _, fn := range listeners {
        func(cb AuthListener) {
            defer func() { recover() }()
            cb(token, record)
        }(fn)
    }
}

func (s *AuthStore) Clear() {
    s.Save("", nil)
}

func splitToken(token string) []string {
    return strings.Split(token, ".")
}
