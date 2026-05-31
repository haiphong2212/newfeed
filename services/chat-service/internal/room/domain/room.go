package domain

type Room struct {
	ID        string
	ArticleID string
	Name      string
}

type Presence struct {
	RoomID string
	UserID string
	Online bool
	Typing bool
}
