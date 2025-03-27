package eywa_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imperfect-fourth/eywa"
)

type MockQuerable struct {
	QueryStr string                 `json:"query"`
	Vars     map[string]interface{} `json:"variables"`
}

func (m *MockQuerable) Query() string {
	return m.QueryStr
}

func (m *MockQuerable) Variables() map[string]interface{} {
	return m.Vars
}

func TestDoClient(t *testing.T) {
	tt := []struct {
		name             string
		server           *httptest.Server
		expectedErr      error
		expectedResponse []byte
	}{
		{
			name: "Valid GQL response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"test": "test"}}`))
			})),
			expectedResponse: []byte(`{"data": {"test": "test"}}`),
		},
		{
			name: "A 401 response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "unauthorized"}}`))
			})),
			expectedErr: eywa.ErrHTTPRequestFailed,
		},
		{
			name: "A 401 status with no response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			})),
			expectedErr: eywa.ErrHTTPRequestFailed,
		},
		{
			name: "A 403 response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "forbidden"}}`))
			})),
			expectedErr: eywa.ErrHTTPRequestFailed,
		},
		{
			name: "A 503 non-json response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`Service unavailable`))
			})),
			expectedErr: eywa.ErrHTTPRequestFailed,
		},
		{
			name: "A 301 Redirect",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusMovedPermanently)
				w.Write([]byte(`Moved Permanently`))
			})),
			expectedErr: eywa.ErrHTTPRequestRedirect,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gqlClient := eywa.NewClient(tc.server.URL, nil)
			resp, err := gqlClient.Do(context.TODO(), &MockQuerable{})
			if err != nil && tc.expectedErr == nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}
			if err == nil && tc.expectedErr != nil {
				t.Errorf("Expected error %v, got nil", tc.expectedErr)
				return
			}
			if err != nil && tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
					return
				}
				return
			}
			if !bytes.Equal(tc.expectedResponse, resp.Bytes()) {
				t.Errorf("Expected response %s, got %s", string(tc.expectedResponse), resp.String())
				return
			}
		})
	}
}

func TestRawClient(t *testing.T) {
	tt := []struct {
		name             string
		server           *httptest.Server
		expectedErr      error
		expectedResponse []byte
	}{
		{
			name: "Valid GQL response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"test": "test"}}`))
			})),
			expectedResponse: []byte(`{"data": {"test": "test"}}`),
		},
		{
			name: "A 401 response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "unauthorized"}}`))
			})),
			expectedResponse: []byte(`{"error": {"message": "unauthorized"}}`),
		},
		{
			name: "A 403 response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": {"message": "forbidden"}}`))
			})),
			expectedResponse: []byte(`{"error": {"message": "forbidden"}}`),
		},
		{
			name: "A 503 non-json response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`Service unavailable`))
			})),
			expectedResponse: []byte(`Service unavailable`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gqlClient := eywa.NewClient(tc.server.URL, nil)
			resp, err := gqlClient.Raw(context.TODO(), &MockQuerable{})
			if err != nil && tc.expectedErr == nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}
			if err == nil && tc.expectedErr != nil {
				t.Errorf("Expected error %v, got nil", tc.expectedErr)
				return
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Error reading response body: %v", err)
				return
			}

			if !bytes.Equal(tc.expectedResponse, respBody) {
				t.Errorf("Expected response %s, got %s", string(tc.expectedResponse), string(respBody))
				return
			}
		})
	}
}
