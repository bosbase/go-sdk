package bosbase

import (
	"fmt"
	"strconv"
)

// VectorDocument represents a single vector embedding entry.
type VectorDocument struct {
	ID       string
	Vector   []float64
	Metadata map[string]interface{}
	Content  string
}

func (d VectorDocument) ToMap() map[string]interface{} {
	payload := map[string]interface{}{"vector": d.Vector}
	if d.ID != "" {
		payload["id"] = d.ID
	}
	if d.Metadata != nil {
		payload["metadata"] = d.Metadata
	}
	if d.Content != "" {
		payload["content"] = d.Content
	}
	return payload
}

func VectorDocumentFromMap(data map[string]interface{}) VectorDocument {
	vec := []float64{}
	if raw, ok := data["vector"].([]interface{}); ok {
		for _, v := range raw {
			switch n := v.(type) {
			case float64:
				vec = append(vec, n)
			case int:
				vec = append(vec, float64(n))
			}
		}
	}
	meta, _ := data["metadata"].(map[string]interface{})
	content, _ := data["content"].(string)
	id, _ := data["id"].(string)
	return VectorDocument{ID: id, Vector: vec, Metadata: meta, Content: content}
}

type VectorSearchOptions struct {
	QueryVector     []float64
	Limit           *int
	Filter          map[string]interface{}
	MinScore        *float64
	MaxDistance     *float64
	IncludeDistance *bool
	IncludeContent  *bool
}

func (o VectorSearchOptions) ToMap() map[string]interface{} {
	payload := map[string]interface{}{"queryVector": o.QueryVector}
	if o.Limit != nil {
		payload["limit"] = *o.Limit
	}
	if o.Filter != nil {
		payload["filter"] = o.Filter
	}
	if o.MinScore != nil {
		payload["minScore"] = *o.MinScore
	}
	if o.MaxDistance != nil {
		payload["maxDistance"] = *o.MaxDistance
	}
	if o.IncludeDistance != nil {
		payload["includeDistance"] = *o.IncludeDistance
	}
	if o.IncludeContent != nil {
		payload["includeContent"] = *o.IncludeContent
	}
	return payload
}

type VectorSearchResult struct {
	Document VectorDocument
	Score    float64
	Distance *float64
}

type VectorSearchResponse struct {
	Results      []VectorSearchResult
	TotalMatches *int
	QueryTime    *int
}

func VectorSearchResponseFromMap(data map[string]interface{}) VectorSearchResponse {
	results := []VectorSearchResult{}
	if arr, ok := data["results"].([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				doc := VectorDocumentFromMap(asMap(m["document"]))
				score := asFloat(m["score"])
				var distance *float64
				if m["distance"] != nil {
					val := asFloat(m["distance"])
					distance = &val
				}
				results = append(results, VectorSearchResult{Document: doc, Score: score, Distance: distance})
			}
		}
	}
	var totalMatches *int
	if v, ok := asIntPointer(data["totalMatches"]); ok {
		totalMatches = v
	}
	var queryTime *int
	if v, ok := asIntPointer(data["queryTime"]); ok {
		queryTime = v
	}
	return VectorSearchResponse{Results: results, TotalMatches: totalMatches, QueryTime: queryTime}
}

type VectorBatchInsertOptions struct {
	Documents      []VectorDocument
	SkipDuplicates *bool
}

func (o VectorBatchInsertOptions) ToMap() map[string]interface{} {
	docs := make([]map[string]interface{}, 0, len(o.Documents))
	for _, d := range o.Documents {
		docs = append(docs, d.ToMap())
	}
	payload := map[string]interface{}{"documents": docs}
	if o.SkipDuplicates != nil {
		payload["skipDuplicates"] = *o.SkipDuplicates
	}
	return payload
}

type VectorInsertResponse struct {
	ID      string
	Success bool
}

func VectorInsertResponseFromMap(data map[string]interface{}) VectorInsertResponse {
	id, _ := data["id"].(string)
	success, _ := data["success"].(bool)
	return VectorInsertResponse{ID: id, Success: success}
}

type VectorBatchInsertResponse struct {
	InsertedCount int
	FailedCount   int
	IDs           []string
	Errors        []string
}

func VectorBatchInsertResponseFromMap(data map[string]interface{}) VectorBatchInsertResponse {
	ids := []string{}
	if arr, ok := data["ids"].([]interface{}); ok {
		for _, v := range arr {
			ids = append(ids, fmt.Sprint(v))
		}
	}
	var errorsList []string
	if arr, ok := data["errors"].([]interface{}); ok {
		for _, v := range arr {
			errorsList = append(errorsList, fmt.Sprint(v))
		}
	}
	return VectorBatchInsertResponse{
		InsertedCount: int(asFloat(data["insertedCount"])),
		FailedCount:   int(asFloat(data["failedCount"])),
		IDs:           ids,
		Errors:        errorsList,
	}
}

type VectorCollectionConfig struct {
	Dimension int
	Distance  string
	Options   map[string]interface{}
}

func (c VectorCollectionConfig) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if c.Dimension != 0 {
		payload["dimension"] = c.Dimension
	}
	if c.Distance != "" {
		payload["distance"] = c.Distance
	}
	if c.Options != nil {
		payload["options"] = c.Options
	}
	return payload
}

type VectorCollectionInfo struct {
	Name      string
	Count     *int
	Dimension *int
}

func VectorCollectionInfoFromMap(data map[string]interface{}) VectorCollectionInfo {
	name, _ := data["name"].(string)
	var count *int
	if v, ok := asIntPointer(data["count"]); ok {
		count = v
	}
	var dim *int
	if v, ok := asIntPointer(data["dimension"]); ok {
		dim = v
	}
	return VectorCollectionInfo{Name: name, Count: count, Dimension: dim}
}

// LangChaingo types

type LangChaingoModelConfig struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

func (c LangChaingoModelConfig) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if c.Provider != "" {
		payload["provider"] = c.Provider
	}
	if c.Model != "" {
		payload["model"] = c.Model
	}
	if c.APIKey != "" {
		payload["apiKey"] = c.APIKey
	}
	if c.BaseURL != "" {
		payload["baseUrl"] = c.BaseURL
	}
	return payload
}

type LangChaingoCompletionMessage struct {
	Content string
	Role    string
}

func (m LangChaingoCompletionMessage) ToMap() map[string]interface{} {
	payload := map[string]interface{}{"content": m.Content}
	if m.Role != "" {
		payload["role"] = m.Role
	}
	return payload
}

type LangChaingoCompletionRequest struct {
	Model          *LangChaingoModelConfig
	Prompt         string
	Messages       []LangChaingoCompletionMessage
	Temperature    *float64
	MaxTokens      *int
	TopP           *float64
	CandidateCount *int
	Stop           []string
	JSONResponse   *bool
}

func (r LangChaingoCompletionRequest) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if r.Model != nil {
		payload["model"] = r.Model.ToMap()
	}
	if r.Prompt != "" {
		payload["prompt"] = r.Prompt
	}
	if len(r.Messages) > 0 {
		msgs := make([]map[string]interface{}, 0, len(r.Messages))
		for _, msg := range r.Messages {
			msgs = append(msgs, msg.ToMap())
		}
		payload["messages"] = msgs
	}
	if r.Temperature != nil {
		payload["temperature"] = *r.Temperature
	}
	if r.MaxTokens != nil {
		payload["maxTokens"] = *r.MaxTokens
	}
	if r.TopP != nil {
		payload["topP"] = *r.TopP
	}
	if r.CandidateCount != nil {
		payload["candidateCount"] = *r.CandidateCount
	}
	if len(r.Stop) > 0 {
		payload["stop"] = r.Stop
	}
	if r.JSONResponse != nil {
		payload["json"] = *r.JSONResponse
	}
	return payload
}

type LangChaingoFunctionCall struct {
	Name      string
	Arguments string
}

type LangChaingoToolCall struct {
	ID           string
	Type         string
	FunctionCall *LangChaingoFunctionCall
}

type LangChaingoCompletionResponse struct {
	Content        string
	StopReason     string
	GenerationInfo map[string]interface{}
	FunctionCall   *LangChaingoFunctionCall
	ToolCalls      []LangChaingoToolCall
}

func LangChaingoCompletionResponseFromMap(data map[string]interface{}) LangChaingoCompletionResponse {
	content, _ := data["content"].(string)
	stopReason, _ := data["stopReason"].(string)
	genInfo, _ := data["generationInfo"].(map[string]interface{})
	var functionCall *LangChaingoFunctionCall
	fc := asMap(data["functionCall"])
	if len(fc) > 0 {
		name, _ := fc["name"].(string)
		args, _ := fc["arguments"].(string)
		functionCall = &LangChaingoFunctionCall{Name: name, Arguments: args}
	}
	var toolCalls []LangChaingoToolCall
	if arr, ok := data["toolCalls"].([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				tc := LangChaingoToolCall{ID: fmt.Sprint(m["id"]), Type: fmt.Sprint(m["type"])}
				fc := asMap(m["functionCall"])
				if len(fc) > 0 {
					name, _ := fc["name"].(string)
					args, _ := fc["arguments"].(string)
					tc.FunctionCall = &LangChaingoFunctionCall{Name: name, Arguments: args}
				}
				toolCalls = append(toolCalls, tc)
			}
		}
	}
	return LangChaingoCompletionResponse{
		Content:        content,
		StopReason:     stopReason,
		GenerationInfo: genInfo,
		FunctionCall:   functionCall,
		ToolCalls:      toolCalls,
	}
}

type LangChaingoRAGFilters struct {
	Where         map[string]string
	WhereDocument map[string]string
}

func (f LangChaingoRAGFilters) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if f.Where != nil {
		payload["where"] = f.Where
	}
	if f.WhereDocument != nil {
		payload["whereDocument"] = f.WhereDocument
	}
	return payload
}

type LangChaingoRAGRequest struct {
	Collection     string
	Question       string
	Model          *LangChaingoModelConfig
	TopK           *int
	ScoreThreshold *float64
	Filters        *LangChaingoRAGFilters
	PromptTemplate string
	ReturnSources  *bool
}

func (r LangChaingoRAGRequest) ToMap() map[string]interface{} {
	payload := map[string]interface{}{
		"collection": r.Collection,
		"question":   r.Question,
	}
	if r.Model != nil {
		payload["model"] = r.Model.ToMap()
	}
	if r.TopK != nil {
		payload["topK"] = *r.TopK
	}
	if r.ScoreThreshold != nil {
		payload["scoreThreshold"] = *r.ScoreThreshold
	}
	if r.Filters != nil {
		payload["filters"] = r.Filters.ToMap()
	}
	if r.PromptTemplate != "" {
		payload["promptTemplate"] = r.PromptTemplate
	}
	if r.ReturnSources != nil {
		payload["returnSources"] = *r.ReturnSources
	}
	return payload
}

type LangChaingoSourceDocument struct {
	Content  string
	Metadata map[string]interface{}
	Score    *float64
}

type LangChaingoRAGResponse struct {
	Answer  string
	Sources []LangChaingoSourceDocument
}

func LangChaingoRAGResponseFromMap(data map[string]interface{}) LangChaingoRAGResponse {
	answer, _ := data["answer"].(string)
	var sources []LangChaingoSourceDocument
	if arr, ok := data["sources"].([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				src := LangChaingoSourceDocument{Content: fmt.Sprint(m["content"])}
				if meta, ok := m["metadata"].(map[string]interface{}); ok {
					src.Metadata = meta
				}
				if m["score"] != nil {
					val := asFloat(m["score"])
					src.Score = &val
				}
				sources = append(sources, src)
			}
		}
	}
	return LangChaingoRAGResponse{Answer: answer, Sources: sources}
}

// DocumentQueryResponse is equivalent to RAG response

type LangChaingoSQLRequest struct {
	Query  string
	Model  *LangChaingoModelConfig
	Tables []string
	TopK   *int
}

func (r LangChaingoSQLRequest) ToMap() map[string]interface{} {
	payload := map[string]interface{}{"query": r.Query}
	if r.Model != nil {
		payload["model"] = r.Model.ToMap()
	}
	if len(r.Tables) > 0 {
		payload["tables"] = r.Tables
	}
	if r.TopK != nil {
		payload["topK"] = *r.TopK
	}
	return payload
}

type LangChaingoSQLResponse struct {
	SQL       string
	Answer    string
	Columns   []string
	Rows      [][]string
	RawResult string
}

func LangChaingoSQLResponseFromMap(data map[string]interface{}) LangChaingoSQLResponse {
	sql, _ := data["sql"].(string)
	answer, _ := data["answer"].(string)
	columns := []string{}
	if arr, ok := data["columns"].([]interface{}); ok {
		for _, v := range arr {
			columns = append(columns, fmt.Sprint(v))
		}
	}
	rows := [][]string{}
	if arr, ok := data["rows"].([]interface{}); ok {
		for _, row := range arr {
			if r, ok := row.([]interface{}); ok {
				line := []string{}
				for _, cell := range r {
					line = append(line, fmt.Sprint(cell))
				}
				rows = append(rows, line)
			}
		}
	}
	rawResult, _ := data["rawResult"].(string)
	return LangChaingoSQLResponse{SQL: sql, Answer: answer, Columns: columns, Rows: rows, RawResult: rawResult}
}

// LLM document types

type LLMDocument struct {
	ID        string
	Content   string
	Metadata  map[string]string
	Embedding []float64
}

func (d LLMDocument) ToMap() map[string]interface{} {
	payload := map[string]interface{}{"content": d.Content}
	if d.ID != "" {
		payload["id"] = d.ID
	}
	if d.Metadata != nil {
		payload["metadata"] = d.Metadata
	}
	if d.Embedding != nil {
		payload["embedding"] = d.Embedding
	}
	return payload
}

func LLMDocumentFromMap(data map[string]interface{}) LLMDocument {
	id, _ := data["id"].(string)
	content, _ := data["content"].(string)
	meta := map[string]string{}
	if raw, ok := data["metadata"].(map[string]interface{}); ok {
		for k, v := range raw {
			meta[k] = fmt.Sprint(v)
		}
	}
	var embedding []float64
	if arr, ok := data["embedding"].([]interface{}); ok {
		for _, v := range arr {
			switch n := v.(type) {
			case float64:
				embedding = append(embedding, n)
			case int:
				embedding = append(embedding, float64(n))
			}
		}
	}
	return LLMDocument{ID: id, Content: content, Metadata: meta, Embedding: embedding}
}

type LLMDocumentUpdate struct {
	Content   *string
	Metadata  map[string]string
	Embedding []float64
}

func (u LLMDocumentUpdate) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if u.Content != nil {
		payload["content"] = *u.Content
	}
	if u.Metadata != nil {
		payload["metadata"] = u.Metadata
	}
	if u.Embedding != nil {
		payload["embedding"] = u.Embedding
	}
	return payload
}

type LLMQueryOptions struct {
	QueryText      string
	QueryEmbedding []float64
	Limit          *int
	Where          map[string]string
	Negative       map[string]interface{}
}

func (o LLMQueryOptions) ToMap() map[string]interface{} {
	payload := map[string]interface{}{}
	if o.QueryText != "" {
		payload["queryText"] = o.QueryText
	}
	if o.QueryEmbedding != nil {
		payload["queryEmbedding"] = o.QueryEmbedding
	}
	if o.Limit != nil {
		payload["limit"] = *o.Limit
	}
	if o.Where != nil {
		payload["where"] = o.Where
	}
	if o.Negative != nil {
		payload["negative"] = o.Negative
	}
	return payload
}

type LLMQueryResult struct {
	ID         string
	Content    string
	Metadata   map[string]string
	Similarity float64
}

func LLMQueryResultFromMap(data map[string]interface{}) LLMQueryResult {
	id, _ := data["id"].(string)
	content, _ := data["content"].(string)
	meta := map[string]string{}
	if raw, ok := data["metadata"].(map[string]interface{}); ok {
		for k, v := range raw {
			meta[k] = fmt.Sprint(v)
		}
	}
	return LLMQueryResult{
		ID:         id,
		Content:    content,
		Metadata:   meta,
		Similarity: asFloat(data["similarity"]),
	}
}

// SQLExecuteResponse represents the response from the SQL execute endpoint.
type SQLExecuteResponse struct {
	Columns      []string
	Rows         [][]string
	RowsAffected int
}

// SQLExecuteResponseFromMap converts a raw map response into SQLExecuteResponse.
func SQLExecuteResponseFromMap(data map[string]interface{}) SQLExecuteResponse {
	cols := []string{}
	if arr, ok := data["columns"].([]interface{}); ok {
		for _, c := range arr {
			cols = append(cols, fmt.Sprint(c))
		}
	}
	rows := [][]string{}
	if arr, ok := data["rows"].([]interface{}); ok {
		for _, r := range arr {
			if rowArr, ok := r.([]interface{}); ok {
				row := []string{}
				for _, val := range rowArr {
					row = append(row, fmt.Sprint(val))
				}
				rows = append(rows, row)
			}
		}
	}
	return SQLExecuteResponse{
		Columns:      cols,
		Rows:         rows,
		RowsAffected: int(asFloat(data["rowsAffected"])),
	}
}

// SqlTableDefinition describes an external SQL table to import or register.
type SqlTableDefinition struct {
	Name string
	SQL  string
}

func (d SqlTableDefinition) ToMap() map[string]interface{} {
	payload := map[string]interface{}{"name": d.Name}
	if d.SQL != "" {
		payload["sql"] = d.SQL
	}
	return payload
}

// SqlTableImportResult contains the created collections and skipped table names.
type SqlTableImportResult struct {
	Created []map[string]interface{}
	Skipped []string
}

func SqlTableImportResultFromMap(data map[string]interface{}) SqlTableImportResult {
	result := SqlTableImportResult{}
	if arr, ok := data["created"].([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				result.Created = append(result.Created, m)
			}
		}
	}
	if arr, ok := data["skipped"].([]interface{}); ok {
		for _, item := range arr {
			result.Skipped = append(result.Skipped, fmt.Sprint(item))
		}
	}
	return result
}

// helper utilities for type conversions

func asMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

func asFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float32:
		return float64(n)
	case string:
		if val, err := strconv.ParseFloat(n, 64); err == nil {
			return val
		}
	}
	return 0
}

func asIntPointer(v interface{}) (*int, bool) {
	switch n := v.(type) {
	case int:
		val := n
		return &val, true
	case int64:
		val := int(n)
		return &val, true
	case float64:
		val := int(n)
		return &val, true
	}
	return nil, false
}
