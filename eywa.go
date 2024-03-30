package eywa

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type Model interface {
	ModelName() string
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
