package service

import (
	"context"
	"encoding/json"
	"fmt"

	"networkai/backend"
	"networkai/constants"
	"networkai/model"

	"github.com/olivere/elastic/v7"
)

func SearchPostsByUser(ctx context.Context, user string) ([]model.Post, error) {
	return searchPosts(ctx, elastic.NewTermQuery("user", user))
}

func SearchPostsByKeywords(ctx context.Context, keywords string) ([]model.Post, error) {
	query := elastic.NewMatchQuery("message", keywords).Operator("AND")
	if keywords == "" {
		query.ZeroTermsQuery("all")
	}
	return searchPosts(ctx, query)
}

func searchPosts(ctx context.Context, query elastic.Query) ([]model.Post, error) {
	result, err := backend.ESBackend.ReadFromES(ctx, query, constants.POST_INDEX)
	if err != nil {
		return nil, err
	}
	posts := make([]model.Post, 0)
	if result.Hits == nil {
		return posts, nil
	}
	for _, hit := range result.Hits.Hits {
		var post model.Post
		if err := json.Unmarshal(hit.Source, &post); err != nil {
			return nil, fmt.Errorf("decode post %s: %w", hit.Id, err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}
