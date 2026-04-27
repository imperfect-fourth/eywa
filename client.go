package eywa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client[QR QueryRoot, MR MutationRoot] struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
}

type ClientOpts struct {
	HTTPClient *http.Client
	Headers    map[string]string
}

// NewClient accepts a graphql endpoint and returns back a Client.
// It uses the http.DefaultClient as the underlying http client by default.
func NewClient[QR QueryRoot, MR MutationRoot](gqlEndpoint string, opts *ClientOpts) *Client[QR, MR] {
	c := &Client[QR, MR]{
		endpoint:   gqlEndpoint,
		httpClient: http.DefaultClient,
	}

	if opts == nil {
		return c
	}

	if opts.HTTPClient != nil {
		c.httpClient = opts.HTTPClient
	}
	if len(opts.Headers) > 0 {
		c.headers = opts.Headers
	}
	return c
}

func (c *Client[QR, MR]) Query(ctx context.Context, operationName string, fields ...*Field[QR]) (*Response[QR], error) {
	resp, err := c.Do(ctx, Request{
		OperationName: operationName,
		Query:         fmt.Sprintf("query %s %s", operationName, marshalGQLFields(fields)),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http response failed with status %d: %s", resp.StatusCode, body)
	}
	var gqlresp Response[QR]
	err = json.NewDecoder(resp.Body).Decode(&gqlresp)
	return &gqlresp, err
}

func (c *Client[QR, MR]) Mutation(ctx context.Context, operationName string, fields ...*Field[QR]) (*Response[MR], error) {
	resp, err := c.Do(ctx, Request{
		OperationName: operationName,
		Query:         fmt.Sprintf("mutation %s %s", operationName, marshalGQLFields(fields)),
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http response failed with status %d: %s", resp.StatusCode, body)
	}
	var gqlresp Response[MR]
	err = json.NewDecoder(resp.Body).Decode(&gqlresp)
	return &gqlresp, err
}

func (c *Client[QR, MR]) Do(ctx context.Context, gqlreq Request) (*http.Response, error) {
	var reqBytes bytes.Buffer
	if err := json.NewEncoder(&reqBytes).Encode(gqlreq); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, &reqBytes)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	for key, value := range c.headers {
		req.Header.Add(key, value)
	}

	return c.httpClient.Do(req)
}

type Request struct {
	OperationName string `json:"operationName"`
	Query         string `json:"query"`
}

type Response[R root] struct {
	Data   R       `json:"data"`
	Errors []Error `json:"errors"`
}

type Error struct {
	Message    string         `json:"message"`
	Extensions map[string]any `json:"extensions"`
}
