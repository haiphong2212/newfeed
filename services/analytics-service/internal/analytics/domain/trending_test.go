package domain

import "testing"

func TestTrendingScore(t *testing.T) {
	score := (ArticleMetrics{
		ViewCount:     100,
		CommentCount:  50,
		ChatActivity:  20,
		ReactionCount: 10,
	}).TrendingScore()

	if score != 60 {
		t.Fatalf("expected score 60, got %v", score)
	}
}
