package bosbase

import (
    "encoding/json"
    "errors"
    "fmt"
    "net/url"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

type PubSubMessage struct {
    ID      string
    Topic   string
    Created string
    Data    interface{}
}

type PublishAck struct {
    ID      string
    Topic   string
    Created string
}

type pubsubPending struct {
    ch    chan map[string]interface{}
    timer *time.Timer
}

type pubsubListener struct {
    id string
    fn func(PubSubMessage)
}

type PubSubService struct {
    BaseService
    conn     *websocket.Conn
    mu       sync.RWMutex
    subs     map[string][]pubsubListener
    pending  map[string]*pubsubPending
    isReady  bool
    clientID string
    counter  int64
}

func NewPubSubService(client *BosBase) *PubSubService {
    return &PubSubService{
        BaseService: BaseService{client: client},
        subs:        map[string][]pubsubListener{},
        pending:     map[string]*pubsubPending{},
    }
}

func (p *PubSubService) Publish(topic string, data interface{}) (PublishAck, error) {
    if topic == "" {
        return PublishAck{}, errors.New("topic must be set")
    }
    if err := p.ensureSocket(); err != nil {
        return PublishAck{}, err
    }
    reqID := p.nextRequestID()
    ackCh := p.waitForAck(reqID)
    envelope := map[string]interface{}{
        "type":      "publish",
        "topic":     topic,
        "data":      data,
        "requestId": reqID,
    }
    if err := p.sendEnvelope(envelope); err != nil {
        return PublishAck{}, err
    }
    payload := <-ackCh
    if payload == nil {
        return PublishAck{}, errors.New("missing publish ack")
    }
    return PublishAck{ID: fmt.Sprint(payload["id"]), Topic: topic, Created: fmt.Sprint(payload["created"])}, nil
}

func (p *PubSubService) Subscribe(topic string, callback func(PubSubMessage)) (func(), error) {
    if topic == "" {
        return nil, errors.New("topic must be set")
    }
    if callback == nil {
        return nil, errors.New("callback must be set")
    }
    p.mu.Lock()
    p.counter++
    listenerID := fmt.Sprintf("l-%d", p.counter)
    listeners := p.subs[topic]
    listeners = append(listeners, pubsubListener{id: listenerID, fn: callback})
    p.subs[topic] = listeners
    shouldSend := len(listeners) == 1
    p.mu.Unlock()

    if err := p.ensureSocket(); err != nil {
        return nil, err
    }
    if shouldSend {
        reqID := p.nextRequestID()
        ack := p.waitForAck(reqID)
        _ = p.sendEnvelope(map[string]interface{}{"type": "subscribe", "topic": topic, "requestId": reqID})
        <-ack
    }

    return func() {
        p.mu.Lock()
        listeners := p.subs[topic]
        filtered := []pubsubListener{}
        for _, entry := range listeners {
            if entry.id == listenerID {
                continue
            }
            filtered = append(filtered, entry)
        }
        if len(filtered) == 0 {
            delete(p.subs, topic)
            reqID := p.nextRequestID()
            ack := p.waitForAck(reqID)
            _ = p.sendEnvelope(map[string]interface{}{"type": "unsubscribe", "topic": topic, "requestId": reqID})
            <-ack
        } else {
            p.subs[topic] = filtered
        }
        p.mu.Unlock()
        if !p.hasSubscriptions() {
            p.Disconnect()
        }
    }, nil
}

func (p *PubSubService) Unsubscribe(topic string) {
    if topic == "" {
        p.mu.Lock()
        p.subs = map[string][]pubsubListener{}
        p.mu.Unlock()
        _ = p.sendEnvelope(map[string]interface{}{"type": "unsubscribe"})
        p.Disconnect()
        return
    }
    p.mu.Lock()
    if _, ok := p.subs[topic]; ok {
        delete(p.subs, topic)
        reqID := p.nextRequestID()
        ack := p.waitForAck(reqID)
        _ = p.sendEnvelope(map[string]interface{}{"type": "unsubscribe", "topic": topic, "requestId": reqID})
        <-ack
    }
    p.mu.Unlock()
    if !p.hasSubscriptions() {
        p.Disconnect()
    }
}

func (p *PubSubService) Disconnect() {
    p.mu.Lock()
    if p.conn != nil {
        _ = p.conn.Close()
        p.conn = nil
    }
    p.isReady = false
    p.pending = map[string]*pubsubPending{}
    p.mu.Unlock()
}

func (p *PubSubService) ensureSocket() error {
    p.mu.RLock()
    if p.conn != nil && p.isReady {
        p.mu.RUnlock()
        return nil
    }
    p.mu.RUnlock()

    wsURL, err := p.buildWSURL()
    if err != nil {
        return err
    }
    conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
    if err != nil {
        return err
    }
    p.mu.Lock()
    p.conn = conn
    p.isReady = false
    p.mu.Unlock()
    go p.listen()
    return nil
}

func (p *PubSubService) buildWSURL() (string, error) {
    query := map[string]interface{}{}
    if p.client.AuthStore != nil && p.client.AuthStore.IsValid() {
        query["token"] = p.client.AuthStore.Token()
    }
    base := p.client.BuildURL("/api/pubsub", query)
    u, err := url.Parse(base)
    if err != nil {
        return "", err
    }
    if u.Scheme == "https" {
        u.Scheme = "wss"
    } else {
        u.Scheme = "ws"
    }
    return u.String(), nil
}

func (p *PubSubService) listen() {
    for {
        p.mu.RLock()
        conn := p.conn
        p.mu.RUnlock()
        if conn == nil {
            return
        }
        _, msg, err := conn.ReadMessage()
        if err != nil {
            p.Disconnect()
            return
        }
        var data map[string]interface{}
        if err := json.Unmarshal(msg, &data); err != nil {
            continue
        }
        p.handleMessage(data)
    }
}

func (p *PubSubService) handleMessage(data map[string]interface{}) {
    msgType := fmt.Sprint(data["type"])
    switch msgType {
    case "ready":
        p.mu.Lock()
        p.clientID = fmt.Sprint(data["clientId"])
        p.isReady = true
        topics := p.getTopicsLocked()
        p.mu.Unlock()
        for _, topic := range topics {
            reqID := p.nextRequestID()
            ack := p.waitForAck(reqID)
            _ = p.sendEnvelope(map[string]interface{}{"type": "subscribe", "topic": topic, "requestId": reqID})
            <-ack
        }
    case "message":
        topic := fmt.Sprint(data["topic"])
        message := PubSubMessage{ID: fmt.Sprint(data["id"]), Topic: topic, Created: fmt.Sprint(data["created"]), Data: data["data"]}
        p.mu.RLock()
        listeners := append([]pubsubListener{}, p.subs[topic]...)
        p.mu.RUnlock()
        for _, entry := range listeners {
            func(cb func(PubSubMessage)) {
                defer func() { recover() }()
                cb(message)
            }(entry.fn)
        }
    case "published", "subscribed", "unsubscribed", "pong":
        if reqID, ok := data["requestId"].(string); ok {
            p.resolvePending(reqID, data)
        }
    case "error":
        if reqID, ok := data["requestId"].(string); ok {
            p.rejectPending(reqID, &ClientResponseError{Response: map[string]interface{}{"message": fmt.Sprint(data["message"])}})
        }
    }
}

func (p *PubSubService) sendEnvelope(data map[string]interface{}) error {
    if err := p.ensureSocket(); err != nil {
        return err
    }
    p.mu.RLock()
    conn := p.conn
    p.mu.RUnlock()
    if conn == nil {
        return errors.New("pubsub connection not initialized")
    }
    payload, _ := json.Marshal(data)
    return conn.WriteMessage(websocket.TextMessage, payload)
}

func (p *PubSubService) waitForAck(requestID string) <-chan map[string]interface{} {
    ch := make(chan map[string]interface{}, 1)
    timer := time.AfterFunc(10*time.Second, func() { ch <- nil })
    p.mu.Lock()
    p.pending[requestID] = &pubsubPending{ch: ch, timer: timer}
    p.mu.Unlock()
    return ch
}

func (p *PubSubService) resolvePending(requestID string, payload map[string]interface{}) {
    p.mu.Lock()
    if pending := p.pending[requestID]; pending != nil {
        pending.timer.Stop()
        pending.ch <- payload
        delete(p.pending, requestID)
    }
    p.mu.Unlock()
}

func (p *PubSubService) rejectPending(requestID string, err error) {
    p.mu.Lock()
    if pending := p.pending[requestID]; pending != nil {
        pending.timer.Stop()
        pending.ch <- map[string]interface{}{"error": err}
        delete(p.pending, requestID)
    }
    p.mu.Unlock()
}

func (p *PubSubService) hasSubscriptions() bool {
    p.mu.RLock()
    defer p.mu.RUnlock()
    return len(p.subs) > 0
}

func (p *PubSubService) getTopicsLocked() []string {
    topics := make([]string, 0, len(p.subs))
    for topic := range p.subs {
        topics = append(topics, topic)
    }
    return topics
}

func (p *PubSubService) nextRequestID() string {
    p.counter++
    return fmt.Sprintf("%d", time.Now().UnixNano()+p.counter)
}
