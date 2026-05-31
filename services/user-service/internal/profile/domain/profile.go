package domain

type Profile struct {
	UserID       string
	DisplayName  string
	Bio          string
	Topics       []string
	FollowingIDs []string
}
