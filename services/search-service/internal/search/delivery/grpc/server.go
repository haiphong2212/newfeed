package grpc

import (
	"context"

	searchv1 "github.com/newfeed/community-news/gen/search/v1"
	"github.com/newfeed/community-news/services/search-service/internal/search/domain"
	"github.com/newfeed/community-news/services/search-service/internal/search/usecase"
)

type Server struct {
	searchv1.UnimplementedSearchServiceServer
	search usecase.Service
}

func NewServer(search usecase.Service) *Server {
	return &Server{search: search}
}

func (s *Server) IndexArticle(ctx context.Context, req *searchv1.IndexArticleRequest) (*searchv1.IndexArticleResponse, error) {
	return &searchv1.IndexArticleResponse{}, s.search.IndexArticle(ctx, domain.Document{
		ArticleID: req.GetArticleId(),
		Title:     req.GetTitle(),
		Content:   req.GetContent(),
		Category:  req.GetCategory(),
		Tags:      req.GetTags(),
	})
}

func (s *Server) SearchArticles(ctx context.Context, req *searchv1.SearchArticlesRequest) (*searchv1.SearchArticlesResponse, error) {
	results, err := s.search.Search(ctx, domain.Query{
		Text:     req.GetQuery(),
		Title:    req.GetTitle(),
		Tag:      req.GetTag(),
		Category: req.GetCategory(),
		Limit:    int(req.GetLimit()),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*searchv1.SearchArticleResult, 0, len(results))
	for _, result := range results {
		out = append(out, &searchv1.SearchArticleResult{
			ArticleId: result.ArticleID,
			Title:     result.Title,
			Snippet:   result.Snippet,
			Score:     result.Score,
		})
	}
	return &searchv1.SearchArticlesResponse{Results: out}, nil
}
