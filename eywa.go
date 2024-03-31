package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
}

func NewClient(endpoint string, headers map[string]string) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: http.DefaultClient,
		headers:    headers,
	}
}

func (c *Client) do(q string) (*bytes.Buffer, error) {
	reqObj := graphqlRequest{
		Query: q,
	}

	var reqBytes bytes.Buffer
	err := json.NewEncoder(&reqBytes).Encode(&reqObj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.endpoint, &reqBytes)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Add(key, value)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBytes bytes.Buffer
	_, err = io.Copy(&respBytes, resp.Body)
	return &respBytes, err
}

type Model interface {
	ModelName() string
}

type query[T Model] struct {
	operationName string
	selectFields  []string

	queryModifier map[string]interface{}

	model T
}

func Query[T Model](model T) *query[T] {
	q := &query[T]{
		operationName: "Get",
		model:         model,
		queryModifier: map[string]interface{}{},
	}
	return q
}

func (q *query[T]) Select(fields ...string) *query[T] {
	q.selectFields = fields
	return q
}

func (q *query[T]) Limit(n int) *query[T] {
	q.queryModifier["limit"] = n
	return q
}

func (q *query[T]) DistinctOn(field string) *query[T] {
	q.queryModifier["distinct_on"] = field
	return q
}

type OrderByEnum string

const (
	OrderAsc            OrderByEnum = "asc"
	OrderAscNullsFirst  OrderByEnum = "asc_nulls_first"
	OrderAscNullsLast   OrderByEnum = "asc_nulls_last"
	OrderDesc           OrderByEnum = "desc"
	OrderDescNullsFirst OrderByEnum = "desc_nulls_first"
	OrderDescNullsLast  OrderByEnum = "desc_nulls_last"
)

func (q *query[T]) OrderBy(orderBys map[string]OrderByEnum) *query[T] {
	orderByClause := ""
	for k, v := range orderBys {
		if orderByClause == "" {
			orderByClause = fmt.Sprintf("{%s: %s", k, v)
		} else {
			orderByClause = fmt.Sprintf("%s, %s: %s", orderByClause, k, v)
		}
	}
	orderByClause += "}"
	q.queryModifier["order_by"] = orderByClause
	return q
}

func (q *query[T]) Where(where string) *query[T] {
	q.queryModifier["where"] = where
	return q
}

func (q *query[T]) build() string {
	baseQueryFormat := "query %s {%s%s {%s}}"

	modifierClause := ""
	for k, v := range q.queryModifier {
		if modifierClause == "" {
			modifierClause = fmt.Sprintf("(%s: %v", k, v)
		} else {
			modifierClause = fmt.Sprintf("%s, %s: %v", modifierClause, k, v)
		}
	}
	if modifierClause != "" {
		modifierClause = modifierClause + ")"
	}

	gql := fmt.Sprintf(
		baseQueryFormat,
		q.operationName,
		q.model.ModelName(),
		modifierClause,
		strings.Join(q.selectFields, "\n"),
	)
	return gql
}

func (q *query[T]) Exec(client *Client) ([]T, error) {
	respBytes, err := client.do(q.build())

	type graphqlResponse struct {
		Data   map[string][]T `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(respBytes).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[q.model.ModelName()], nil
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
}
