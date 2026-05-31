package usecase

import "github.com/newfeed/community-news/services/chat-service/internal/room/domain"

type Service struct{}

func NewService() Service {
	return Service{}
}

func (Service) RoomForArticle(articleID, articleTitle string) domain.Room {
	return domain.Room{ID: articleID, ArticleID: articleID, Name: articleTitle + " Live Discussion"}
}
