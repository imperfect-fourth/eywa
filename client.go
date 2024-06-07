package eywa

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
	headers    map[string]string
}

// NewClient accepts a graphql endpoint and returns back a Client.
// It uses the http.DefaultClient as the underlying http client by default.
func NewClient(gqlEndpoint string) *Client {
	return &Client{
		endpoint:   gqlEndpoint,
		httpClient: http.DefaultClient,
	}
}

// WithHeaders adds the given headers to all the requests made by the Client.
func (c *Client) WithHeaders(headers map[string]string) *Client {
	c.headers = headers
	return c
}

// WithHttpClient sets the underlying HTTP client that will be used for making graphql calls.
func (c *Client) WithHttpClient(hc *http.Client) *Client {
	c.httpClient = hc
	return c
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
