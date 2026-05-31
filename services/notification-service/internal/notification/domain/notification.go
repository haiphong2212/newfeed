package domain

import "time"

type EventName string

const (
	EventArticlePublished  EventName = "ArticlePublished"
	EventUserMentioned     EventName = "UserMentioned"
	EventCommentCreated    EventName = "CommentCreated"
	EventFollowTopicCreate EventName = "FollowTopicCreated"
)

type Notification struct {
	ID        string
	UserID    string
	Type      EventName
	Title     string
	Body      string
	ReadAt    *time.Time
	CreatedAt time.Time
}
