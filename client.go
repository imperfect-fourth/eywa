package eywa

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
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
func NewClient(gqlEndpoint string, opt *ClientOpts) *Client {
	c := &Client{
		endpoint:   gqlEndpoint,
		httpClient: http.DefaultClient,
	}

	if opt != nil {
		if opt.HTTPClient != nil {
			c.httpClient = opt.HTTPClient
		}

		if len(opt.Headers) > 0 {
			c.headers = opt.Headers
		}
	}

	return c
}

var (
	ErrHTTPRedirect     = errors.New("http request redirected")
	ErrHTTPRequestFailed = errors.New("http request failed")
)

// Do performs a gql query and returns early if faced with a non-successful http status code.
func (c *Client) Do(ctx context.Context, q Queryable) (*bytes.Buffer, error) {
	resp, err := c.Raw(ctx, q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBytes bytes.Buffer
	_, err = io.Copy(&respBytes, resp.Body)

	switch {
	case resp.StatusCode > 299 && resp.StatusCode < 399:
		err = fmt.Errorf("%w: %d", ErrHTTPRedirect, resp.StatusCode)
	case resp.StatusCode > 399:
		err = fmt.Errorf("%w: %d", ErrHTTPFailedStatus, resp.StatusCode)
	}

	return &respBytes, err
}

// Raw performs a gql query and returns the raw http response and error from the underlying http client.
// Make sure to close the response body.
func (c *Client) Raw(ctx context.Context, q Queryable) (*http.Response, error) {
	reqObj := graphqlRequest{
		Query:     q.Query(),
		Variables: q.Variables(),
	}

	var reqBytes bytes.Buffer
	err := json.NewEncoder(&reqBytes).Encode(&reqObj)
	if err != nil {
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
