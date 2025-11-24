package bosbase

import (
    "errors"
    "fmt"
    "strings"
)

type CollectionService struct {
    BaseCrudService
}

func NewCollectionService(client *BosBase) *CollectionService {
    svc := &CollectionService{}
    svc.BaseCrudService = NewBaseCrudService(client, func() string { return "/api/collections" })
    return svc
}

func (s *CollectionService) DeleteCollection(idOrName string, opts *CrudDeleteOptions) error {
    return s.Delete(idOrName, opts)
}

func (s *CollectionService) Truncate(idOrName string, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    path := fmt.Sprintf("%s/%s/truncate", s.basePath(), encodePathSegment(idOrName))
    _, err := s.client.Send(path, &RequestOptions{Method: "DELETE", Body: body, Query: query, Headers: headers})
    return err
}

func (s *CollectionService) ImportCollections(collections interface{}, deleteMissing bool, body map[string]interface{}, query map[string]interface{}, headers map[string]string) error {
    payload := cloneQuery(body)
    if payload == nil {
        payload = map[string]interface{}{}
    }
    payload["collections"] = collections
    payload["deleteMissing"] = deleteMissing
    _, err := s.client.Send(s.basePath()+"/import", &RequestOptions{Method: "PUT", Body: payload, Query: query, Headers: headers})
    return err
}

func (s *CollectionService) GetScaffolds(body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send(s.basePath()+"/meta/scaffolds", &RequestOptions{Body: body, Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CollectionService) createFromScaffold(scaffoldType, name string, overrides map[string]interface{}, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    scaffolds, err := s.GetScaffolds(nil, query, headers)
    if err != nil {
        return nil, err
    }
    scaffoldRaw, ok := scaffolds[scaffoldType].(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("scaffold for type '%s' not found", scaffoldType)
    }
    data := cloneQuery(scaffoldRaw)
    data["name"] = name
    for k, v := range overrides {
        data[k] = v
    }
    for k, v := range body {
        data[k] = v
    }
    return s.Create(&CrudMutateOptions{Body: data, Query: query, Headers: headers})
}

func (s *CollectionService) CreateBase(name string, overrides map[string]interface{}, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    return s.createFromScaffold("base", name, overrides, body, query, headers)
}

func (s *CollectionService) CreateAuth(name string, overrides map[string]interface{}, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    return s.createFromScaffold("auth", name, overrides, body, query, headers)
}

func (s *CollectionService) CreateView(name string, viewQuery string, overrides map[string]interface{}, body map[string]interface{}, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    scaffoldOverrides := cloneQuery(overrides)
    if viewQuery != "" {
        scaffoldOverrides["viewQuery"] = viewQuery
    }
    return s.createFromScaffold("view", name, scaffoldOverrides, body, query, headers)
}

func (s *CollectionService) AddIndex(collection string, columns []string, unique bool, indexName string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    if len(columns) == 0 {
        return nil, errors.New("at least one column must be specified")
    }
    current, err := s.GetOne(collection, &CrudViewOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    fields, _ := current["fields"].([]interface{})
    var fieldNames []string
    for _, f := range fields {
        if m, ok := f.(map[string]interface{}); ok {
            if name, ok := m["name"].(string); ok && name != "" {
                fieldNames = append(fieldNames, name)
            }
        }
    }
    for _, col := range columns {
        if col != "id" && !contains(fieldNames, col) {
            return nil, fmt.Errorf("field \"%s\" does not exist in the collection", col)
        }
    }
    collectionName := current["name"]
    cname := collection
    if str, ok := collectionName.(string); ok && str != "" {
        cname = str
    }
    idxName := indexName
    if idxName == "" {
        idxName = fmt.Sprintf("idx_%s_%s", cname, strings.Join(columns, "_"))
    }
    columnsSQL := "`" + strings.Join(columns, "`, `") + "`"
    indexSQL := fmt.Sprintf("CREATE %sINDEX `%s` ON `%s` (%s)", func() string {
        if unique {
            return "UNIQUE "
        }
        return ""
    }(), idxName, cname, columnsSQL)

    indexesRaw, _ := current["indexes"].([]interface{})
    for _, idx := range indexesRaw {
        if str, ok := idx.(string); ok && str == indexSQL {
            return nil, errors.New("index already exists")
        }
    }
    indexes := make([]interface{}, 0, len(indexesRaw)+1)
    indexes = append(indexes, indexesRaw...)
    indexes = append(indexes, indexSQL)
    current["indexes"] = indexes
    return s.Update(collection, &CrudMutateOptions{Body: current, Query: query, Headers: headers})
}

func (s *CollectionService) RemoveIndex(collection string, columns []string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    if len(columns) == 0 {
        return nil, errors.New("at least one column must be specified")
    }
    current, err := s.GetOne(collection, &CrudViewOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    indexesRaw, _ := current["indexes"].([]interface{})
    initial := len(indexesRaw)
    var filtered []interface{}
    for _, idx := range indexesRaw {
        strIdx, ok := idx.(string)
        if !ok {
            continue
        }
        match := true
        for _, col := range columns {
            backticked := "`" + col + "`"
            if strings.Contains(strIdx, backticked) || strings.Contains(strIdx, "("+col+")") || strings.Contains(strIdx, "("+col+",") || strings.Contains(strIdx, ", "+col+")") {
                continue
            }
            match = false
            break
        }
        if !match {
            filtered = append(filtered, strIdx)
        }
    }
    if len(filtered) == initial {
        return nil, errors.New("index not found")
    }
    current["indexes"] = filtered
    return s.Update(collection, &CrudMutateOptions{Body: current, Query: query, Headers: headers})
}

func (s *CollectionService) GetIndexes(collection string, query map[string]interface{}, headers map[string]string) ([]string, error) {
    current, err := s.GetOne(collection, &CrudViewOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    indexesRaw, _ := current["indexes"].([]interface{})
    indexes := []string{}
    for _, idx := range indexesRaw {
        if str, ok := idx.(string); ok {
            indexes = append(indexes, str)
        }
    }
    return indexes, nil
}

func (s *CollectionService) GetSchema(collection string, query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    path := fmt.Sprintf("%s/%s/schema", s.basePath(), encodePathSegment(collection))
    data, err := s.client.Send(path, &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func (s *CollectionService) GetAllSchemas(query map[string]interface{}, headers map[string]string) (map[string]interface{}, error) {
    data, err := s.client.Send(s.basePath()+"/schemas", &RequestOptions{Query: query, Headers: headers})
    if err != nil {
        return nil, err
    }
    if m, ok := data.(map[string]interface{}); ok {
        return m, nil
    }
    return map[string]interface{}{}, nil
}

func contains(list []string, target string) bool {
    for _, v := range list {
        if v == target {
            return true
        }
    }
    return false
}
