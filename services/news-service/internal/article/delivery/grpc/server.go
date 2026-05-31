package grpc

import (
	"context"

	newsv1 "github.com/newfeed/community-news/gen/news/v1"
	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
	"github.com/newfeed/community-news/services/news-service/internal/article/usecase"
)

type Server struct {
	newsv1.UnimplementedNewsServiceServer
	articles *usecase.Service
}

func NewServer(articles *usecase.Service) *Server {
	return &Server{articles: articles}
}

func (s *Server) PublishArticle(ctx context.Context, req *newsv1.PublishArticleRequest) (*newsv1.PublishArticleResponse, error) {
	article := domain.Article{
		ID:       req.GetId(),
		AuthorID: req.GetAuthorId(),
		Title:    req.GetTitle(),
		Content:  req.GetContent(),
		Category: req.GetCategory(),
		Tags:     req.GetTags(),
	}
	if err := s.articles.Publish(ctx, article); err != nil {
		return nil, err
	}
	return &newsv1.PublishArticleResponse{Status: string(domain.StatusPublished), DiscussionRoomName: article.DiscussionRoomName()}, nil
}
