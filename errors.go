package bosbase

import "fmt"

// ClientResponseError represents a normalized HTTP error from BosBase.
type ClientResponseError struct {
    URL          string
    Status       int
    Response     map[string]interface{}
    IsAbort      bool
    OriginalErr  error
}

func (e *ClientResponseError) Error() string {
    return fmt.Sprintf("ClientResponseError(status=%d, url=%s, response=%v, isAbort=%t)", e.Status, e.URL, e.Response, e.IsAbort)
}

// Unwrap allows errors.Is/As to see the original error when present.
func (e *ClientResponseError) Unwrap() error {
    return e.OriginalErr
}
