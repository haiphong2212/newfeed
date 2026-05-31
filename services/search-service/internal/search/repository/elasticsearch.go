package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/newfeed/community-news/services/search-service/internal/search/domain"
)

const ArticleIndex = "articles"

type ElasticsearchRepository struct {
	es *elasticsearch.Client
}

func NewElasticsearchRepository(es *elasticsearch.Client) *ElasticsearchRepository {
	return &ElasticsearchRepository{es: es}
}

func (r *ElasticsearchRepository) IndexArticle(ctx context.Context, doc domain.Document) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	res, err := r.es.Index(ArticleIndex, bytes.NewReader(body), r.es.Index.WithContext(ctx), r.es.Index.WithDocumentID(doc.ArticleID), r.es.Index.WithRefresh("true"))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		data, _ := io.ReadAll(res.Body)
		return &Error{Message: string(data)}
	}
	return nil
}

func (r *ElasticsearchRepository) Search(ctx context.Context, query domain.Query) ([]domain.Result, error) {
	body := map[string]any{
		"size": query.Limit,
		"query": map[string]any{
			"bool": map[string]any{
				"must":   mustQueries(query),
				"filter": filterQueries(query),
			},
		},
	}
	data, _ := json.Marshal(body)
	res, err := r.es.Search(r.es.Search.WithContext(ctx), r.es.Search.WithIndex(ArticleIndex), r.es.Search.WithBody(bytes.NewReader(data)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		data, _ := io.ReadAll(res.Body)
		return nil, &Error{Message: string(data)}
	}
	var payload struct {
		Hits struct {
			Hits []struct {
				Score  float64         `json:"_score"`
				Source domain.Document `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	results := make([]domain.Result, 0, len(payload.Hits.Hits))
	for _, hit := range payload.Hits.Hits {
		results = append(results, domain.Result{
			ArticleID: hit.Source.ArticleID,
			Title:     hit.Source.Title,
			Snippet:   hit.Source.Content,
			Score:     hit.Score,
		})
	}
	return results, nil
}

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func mustQueries(query domain.Query) []map[string]any {
	if query.Text == "" && query.Title == "" {
		return []map[string]any{{"match_all": map[string]any{}}}
	}
	must := make([]map[string]any, 0, 2)
	if query.Text != "" {
		must = append(must, map[string]any{"multi_match": map[string]any{"query": query.Text, "fields": []string{"title^3", "content", "tags"}}})
	}
	if query.Title != "" {
		must = append(must, map[string]any{"match": map[string]any{"title": query.Title}})
	}
	return must
}

func filterQueries(query domain.Query) []map[string]any {
	filter := make([]map[string]any, 0, 2)
	if query.Tag != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"tags.keyword": query.Tag}})
	}
	if query.Category != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"category.keyword": query.Category}})
	}
	return filter
}
