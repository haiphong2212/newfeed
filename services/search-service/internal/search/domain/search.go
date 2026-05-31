package domain

type Query struct {
	Text     string
	Title    string
	Tag      string
	Category string
	Limit    int
	Cursor   string
}

type Document struct {
	ArticleID string   `json:"article_id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
}

type Result struct {
	ArticleID string
	Title     string
	Snippet   string
	Score     float64
}
