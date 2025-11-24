package bosbase

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"
)

type RealtimeService struct {
    BaseService
    ClientID     string
    OnDisconnect func([]string)

    mu            sync.RWMutex
    subscriptions map[string][]realtimeListener
    stopCh        chan struct{}
    readyCh       chan struct{}
    running       bool
    counter       int64
}

type realtimeListener struct {
    id string
    fn func(map[string]interface{})
}

func NewRealtimeService(client *BosBase) *RealtimeService {
    return &RealtimeService{
        BaseService:   BaseService{client: client},
        subscriptions: map[string][]realtimeListener{},
    }
}

func (r *RealtimeService) Subscribe(topic string, callback func(map[string]interface{}), query map[string]interface{}, headers map[string]string) (func(), error) {
    if topic == "" {
        return nil, errors.New("topic must be set")
    }
    if callback == nil {
        return nil, errors.New("callback must be set")
    }
    key := r.buildSubscriptionKey(topic, query, headers)
    r.mu.Lock()
    r.counter++
    listenerID := fmt.Sprintf("l-%d", r.counter)
    listeners := r.subscriptions[key]
    listeners = append(listeners, realtimeListener{id: listenerID, fn: callback})
        r.subscriptions[key] = listeners
    r.mu.Unlock()

    r.ensureThread()
    if err := r.EnsureConnected(10 * time.Second); err != nil {
        return nil, err
    }
    r.submitSubscriptions()

    return func() { r.UnsubscribeByTopicAndID(topic, listenerID) }, nil
}

func (r *RealtimeService) Unsubscribe(topic string) {
    r.mu.Lock()
    if topic == "" {
        r.subscriptions = map[string][]realtimeListener{}
    } else {
        for key := range r.subscriptions {
            if key == topic || strings.HasPrefix(key, topic+"?") {
                delete(r.subscriptions, key)
            }
        }
    }
    has := len(r.subscriptions) > 0
    r.mu.Unlock()

    if has {
        r.submitSubscriptions()
    } else {
        r.Disconnect()
    }
}

func (r *RealtimeService) UnsubscribeByPrefix(prefix string) {
    r.Unsubscribe(prefix)
}

func (r *RealtimeService) UnsubscribeByTopicAndID(topic, id string) {
    r.mu.Lock()
    for key, listeners := range r.subscriptions {
        if key != topic && !strings.HasPrefix(key, topic+"?") {
            continue
        }
        filtered := []realtimeListener{}
        for _, entry := range listeners {
            if entry.id == id {
                continue
            }
            filtered = append(filtered, entry)
        }
        if len(filtered) == 0 {
            delete(r.subscriptions, key)
        } else {
            r.subscriptions[key] = filtered
        }
    }
    has := len(r.subscriptions) > 0
    r.mu.Unlock()

    if has {
        r.submitSubscriptions()
    } else {
        r.Disconnect()
    }
}

func (r *RealtimeService) Disconnect() {
    r.mu.Lock()
    if r.stopCh != nil {
        close(r.stopCh)
    }
    r.running = false
    r.readyCh = nil
    r.ClientID = ""
    r.mu.Unlock()
}

func (r *RealtimeService) EnsureConnected(timeout time.Duration) error {
    r.ensureThread()
    r.mu.RLock()
    readyCh := r.readyCh
    r.mu.RUnlock()
    if readyCh == nil {
        return errors.New("realtime connection not initialized")
    }
    select {
    case <-readyCh:
        return nil
    case <-time.After(timeout):
        return &ClientResponseError{Response: map[string]interface{}{"message": "Realtime connection not established"}}
    }
}

func (r *RealtimeService) ensureThread() {
    r.mu.Lock()
    if r.running {
        r.mu.Unlock()
        return
    }
    r.stopCh = make(chan struct{})
    r.readyCh = make(chan struct{})
    r.running = true
    r.mu.Unlock()
    go r.run()
}

func (r *RealtimeService) run() {
    backoff := []time.Duration{200 * time.Millisecond, 500 * time.Millisecond, time.Second, 2 * time.Second, 5 * time.Second}
    attempt := 0
    baseURL := r.client.BuildURL("/api/realtime", nil)

    for {
        r.mu.RLock()
        stopCh := r.stopCh
        r.mu.RUnlock()
        select {
        case <-stopCh:
            return
        default:
        }

        req, _ := http.NewRequest(http.MethodGet, baseURL, nil)
        req.Header.Set("Accept", "text/event-stream")
        req.Header.Set("Cache-Control", "no-store")
        req.Header.Set("Accept-Language", r.client.Lang)
        req.Header.Set("User-Agent", userAgent)
        if r.client.AuthStore != nil && r.client.AuthStore.IsValid() {
            req.Header.Set("Authorization", r.client.AuthStore.Token())
        }

        resp, err := r.client.httpClient.Do(req)
        if err != nil || resp.StatusCode >= 400 {
            r.handleDisconnect()
            delay := backoff[min(attempt, len(backoff)-1)]
            attempt++
            time.Sleep(delay)
            continue
        }
        attempt = 0
        r.listen(resp)
        resp.Body.Close()
        r.handleDisconnect()
        if !r.hasSubscriptions() {
            return
        }
    }
}

func (r *RealtimeService) listen(resp *http.Response) {
    reader := bufio.NewReader(resp.Body)
    event := map[string]string{"event": "message", "data": "", "id": ""}
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            return
        }
        select {
        case <-r.stopCh:
            return
        default:
        }
        line = strings.TrimRight(line, "\r\n")
        if line == "" {
            r.dispatchEvent(event)
            event = map[string]string{"event": "message", "data": "", "id": ""}
            continue
        }
        if strings.HasPrefix(line, ":") {
            continue
        }
        if parts := strings.SplitN(line, ":", 2); len(parts) == 2 {
            field := parts[0]
            value := strings.TrimLeft(parts[1], " ")
            switch field {
            case "event":
                event["event"] = value
            case "data":
                event["data"] += value + "\n"
            case "id":
                event["id"] = value
            }
        }
    }
}

func (r *RealtimeService) dispatchEvent(evt map[string]string) {
    name := evt["event"]
    if name == "" {
        name = "message"
    }
    dataStr := strings.TrimSuffix(evt["data"], "\n")
    var payload map[string]interface{}
    if dataStr != "" {
        _ = json.Unmarshal([]byte(dataStr), &payload)
    }
    if payload == nil {
        payload = map[string]interface{}{}
    }

    if name == "PB_CONNECT" {
        r.mu.Lock()
        clientID := fmt.Sprint(payload["clientId"])
        if clientID == "" {
            clientID = evt["id"]
        }
        r.ClientID = clientID
        ready := r.readyCh
        if ready != nil {
            select {
            case <-ready:
            default:
                close(ready)
            }
        }
        r.mu.Unlock()
        r.submitSubscriptions()
        if r.OnDisconnect != nil {
            r.OnDisconnect([]string{})
        }
        return
    }

    r.mu.RLock()
    entries := append([]realtimeListener{}, r.subscriptions[name]...)
    r.mu.RUnlock()
    for _, entry := range entries {
        func(cb func(map[string]interface{})) {
            defer func() { recover() }()
            cb(payload)
        }(entry.fn)
    }
}

func (r *RealtimeService) submitSubscriptions() {
    r.mu.RLock()
    clientID := r.ClientID
    subs := r.getActiveSubscriptionsLocked()
    r.mu.RUnlock()
    if clientID == "" || len(subs) == 0 {
        return
    }
    payload := map[string]interface{}{
        "clientId":     clientID,
        "subscriptions": subs,
    }
    _, _ = r.client.Send("/api/realtime", &RequestOptions{Method: http.MethodPost, Body: payload})
}

func (r *RealtimeService) GetActiveSubscriptions() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.getActiveSubscriptionsLocked()
}

func (r *RealtimeService) getActiveSubscriptionsLocked() []string {
    subs := make([]string, 0, len(r.subscriptions))
    for key := range r.subscriptions {
        subs = append(subs, key)
    }
    return subs
}

func (r *RealtimeService) hasSubscriptions() bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return len(r.subscriptions) > 0
}

func (r *RealtimeService) handleDisconnect() {
    r.mu.Lock()
    if r.readyCh != nil {
        r.readyCh = make(chan struct{})
    }
    r.ClientID = ""
    subs := r.getActiveSubscriptionsLocked()
    r.mu.Unlock()
    if r.OnDisconnect != nil {
        r.OnDisconnect(subs)
    }
}

func (r *RealtimeService) buildSubscriptionKey(topic string, query map[string]interface{}, headers map[string]string) string {
    key := topic
    options := map[string]interface{}{}
    if len(query) > 0 {
        options["query"] = query
    }
    if len(headers) > 0 {
        options["headers"] = headers
    }
    if len(options) > 0 {
        serialized, _ := json.Marshal(options)
        suffix := "options=" + url.QueryEscape(string(serialized))
        if strings.Contains(key, "?") {
            key += "&" + suffix
        } else {
            key += "?" + suffix
        }
    }
    return key
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
