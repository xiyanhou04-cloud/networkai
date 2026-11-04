package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"networkai/backend"
	"networkai/config"
)

func TestSearchResponseAndBackendFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   int
	}{
		{"empty results", 200, `{"hits":{"hits":[]}}`, 200},
		{"backend failure", 503, `{"error":{"type":"unavailable","reason":"test failure"},"status":503}`, 500},
		{"invalid post data", 200, `{"hits":{"hits":[{"_id":"bad","_source":{"message":42}}]}}`, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodHead {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			b, err := backend.NewElasticsearchBackend(context.Background(), config.ElasticsearchConfig{URL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			defer b.Close()
			previous := backend.ESBackend
			backend.ESBackend = b
			defer func() { backend.ESBackend = previous }()
			response := httptest.NewRecorder()
			InitRouter().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/posts/search?user=sun", nil))
			if response.Code != tc.want {
				t.Fatalf("HTTP %d, want %d: %s", response.Code, tc.want, response.Body.String())
			}
			if tc.want == 200 && (response.Body.String() != "[]" || response.Header().Get("Content-Type") != "application/json") {
				t.Fatalf("empty results should be a JSON array: %s", response.Body.String())
			}
		})
	}
}
