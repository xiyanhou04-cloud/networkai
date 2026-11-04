package backend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"networkai/config"
	"networkai/constants"

	"github.com/olivere/elastic/v7"
)

var ESBackend *ElasticsearchBackend

type ElasticsearchBackend struct {
	client *elastic.Client
}

const postMapping = `{
  "mappings": {"properties": {
    "id": {"type": "keyword"},
    "user": {"type": "keyword"},
    "message": {"type": "text"},
    "url": {"type": "keyword", "index": false},
    "type": {"type": "keyword", "index": false}
  }}
}`

const userMapping = `{
  "mappings": {"properties": {
    "username": {"type": "keyword"},
    "password": {"type": "keyword"},
    "age": {"type": "long", "index": false},
    "gender": {"type": "keyword", "index": false}
  }}
}`

func InitElasticsearchBackend() error {
	cfg, err := config.LoadElasticsearchConfig()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	backend, err := NewElasticsearchBackend(ctx, cfg)
	if err != nil {
		return err
	}
	ESBackend = backend
	fmt.Println("Indexes are created.")
	return nil
}

func NewElasticsearchBackend(ctx context.Context, cfg config.ElasticsearchConfig) (*ElasticsearchBackend, error) {
	client, err := elastic.DialContext(ctx,
		elastic.SetURL(cfg.URL),
		elastic.SetBasicAuth(cfg.Username, cfg.Password),
		elastic.SetSniff(false),
		elastic.SetHttpClient(&http.Client{Timeout: 10 * time.Second}),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to Elasticsearch: %w", err)
	}
	for _, index := range []struct{ name, mapping string }{
		{constants.POST_INDEX, postMapping},
		{constants.USER_INDEX, userMapping},
	} {
		exists, err := client.IndexExists(index.name).Do(ctx)
		if err != nil {
			client.Stop()
			return nil, fmt.Errorf("check index %s: %w", index.name, err)
		}
		if exists {
			continue
		}
		if _, err := client.CreateIndex(index.name).Body(index.mapping).Do(ctx); err != nil {
			client.Stop()
			return nil, fmt.Errorf("create index %s: %w", index.name, err)
		}
	}
	return &ElasticsearchBackend{client: client}, nil
}

func (backend *ElasticsearchBackend) Close() {
	if backend != nil && backend.client != nil {
		backend.client.Stop()
	}
}

func (backend *ElasticsearchBackend) ReadFromES(ctx context.Context, query elastic.Query, index string) (*elastic.SearchResult, error) {
	if backend == nil || backend.client == nil {
		return nil, errors.New("Elasticsearch backend is not initialized")
	}
	return backend.client.Search().Index(index).Query(query).Do(ctx)
}
