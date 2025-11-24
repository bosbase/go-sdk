package bosbase

import (
    "fmt"
    "net/url"
    "strings"
)

// encodePathSegment encodes a single URL path segment.
func encodePathSegment(value interface{}) string {
    return url.PathEscape(fmt.Sprint(value))
}

// normalizeQueryParams converts a generic map into url.Values, skipping nil entries.
func normalizeQueryParams(params map[string]interface{}) url.Values {
    values := url.Values{}
    if params == nil {
        return values
    }
    for key, raw := range params {
        if raw == nil {
            continue
        }
        switch v := raw.(type) {
        case []string:
            for _, item := range v {
                values.Add(key, item)
            }
        case []interface{}:
            for _, item := range v {
                values.Add(key, fmt.Sprint(item))
            }
        case []int:
            for _, item := range v {
                values.Add(key, fmt.Sprint(item))
            }
        case []float64:
            for _, item := range v {
                values.Add(key, fmt.Sprint(item))
            }
        default:
            values.Add(key, fmt.Sprint(raw))
        }
    }
    return values
}

// buildRelativeURL builds a URL path with query parameters (without base host).
func buildRelativeURL(path string, query map[string]interface{}) string {
    rel := "/" + strings.TrimLeft(path, "/")
    if len(query) == 0 {
        return rel
    }
    values := normalizeQueryParams(query)
    if len(values) == 0 {
        return rel
    }
    return rel + "?" + values.Encode()
}

// toSerializable recursively removes nil values so JSON encoding stays compact.
func toSerializable(value interface{}) interface{} {
    switch val := value.(type) {
    case nil:
        return nil
    case map[string]interface{}:
        result := make(map[string]interface{}, len(val))
        for k, v := range val {
            if v == nil {
                continue
            }
            result[k] = toSerializable(v)
        }
        return result
    case []interface{}:
        res := make([]interface{}, 0, len(val))
        for _, item := range val {
            res = append(res, toSerializable(item))
        }
        return res
    case map[string]string:
        result := make(map[string]interface{}, len(val))
        for k, v := range val {
            result[k] = v
        }
        return result
    default:
        return value
    }
}
