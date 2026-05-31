package domain

type ArticleMetrics struct {
	ArticleID     string
	ViewCount     int64
	CommentCount  int64
	ChatActivity  int64
	ReactionCount int64
}

func (m ArticleMetrics) TrendingScore() float64 {
	return float64(m.ViewCount)*0.4 +
		float64(m.CommentCount)*0.3 +
		float64(m.ChatActivity)*0.2 +
		float64(m.ReactionCount)*0.1
}
