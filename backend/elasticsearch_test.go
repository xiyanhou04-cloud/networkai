package backend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"networkai/config"
)

func TestInitializationPreservesExistingIndexes(t *testing.T) {
	var mu sync.Mutex
	exists := map[string]bool{"/post": true}
	created := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "course" || password != "test-password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodHead:
			if exists[r.URL.Path] {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case http.MethodPut:
			created[r.URL.Path]++
			exists[r.URL.Path] = true
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"acknowledged":true,"shards_acknowledged":true}`)
		default:
			t.Errorf("unexpected operation: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	cfg := config.ElasticsearchConfig{URL: server.URL, Username: "course", Password: "test-password"}
	for range 2 {
		b, err := NewElasticsearchBackend(context.Background(), cfg)
		if err != nil {
			t.Fatal(err)
		}
		b.Close()
	}
	mu.Lock()
	defer mu.Unlock()
	if created["/post"] != 0 || created["/user"] != 1 {
		t.Fatalf("existing indexes should survive initialization; creations: %v", created)
	}
}
