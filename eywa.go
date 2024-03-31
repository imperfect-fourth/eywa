package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
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
	whereClause   string
	limitClause   string
	orderByClause string

	model T
}

func Query[T Model](model T, opts ...queryOption[T]) *query[T] {
	q := &query[T]{
		operationName: "Get",
		model:         model,
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

func (q *query[T]) Select(fields []string) *query[T] {
	q.selectFields = fields
	return q
}

func (q *query[T]) Exec(client *Client) ([]T, error) {
	queryFormat := "query %s {%s {%s}}"
	respBytes, err := client.do(fmt.Sprintf(queryFormat, q.operationName, q.model.ModelName(), strings.Join(q.selectFields, "\n")))

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

type queryOption[T Model] func(*query[T])

func WithOperationName[T Model](oprnName string) queryOption[T] {
	return func(q *query[T]) {
		q.operationName = oprnName
	}
}

func SelectFields[T Model](fields []string) queryOption[T] {
	return func(q *query[T]) {
		q.selectFields = fields
	}
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions"`
}

const selectQueryFormat string = "query Get {%s {%s}}"

func Select[T Model](client *Client, model T) ([]T, error) {
	modelFields := reflect.VisibleFields(reflect.TypeOf(model).Elem())

	var queryFields []string
	for _, field := range modelFields {
		queryFields = append(queryFields, field.Tag.Get("graphql"))
	}
	query := fmt.Sprintf(selectQueryFormat, model.ModelName(), strings.Join(queryFields, "\n"))

	reqObj := graphqlRequest{
		Query: query,
	}

	var reqBytes bytes.Buffer
	err := json.NewEncoder(&reqBytes).Encode(&reqObj)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, client.endpoint, &reqBytes)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	for key, value := range client.headers {
		req.Header.Add(key, value)
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type graphqlResponse struct {
		Data   map[string][]T `json:"data"`
		Errors []graphqlError `json:"errors"`
	}

	respObj := graphqlResponse{}

	err = json.NewDecoder(resp.Body).Decode(&respObj)
	if err != nil {
		return nil, err
	}
	return respObj.Data[model.ModelName()], nil

}
