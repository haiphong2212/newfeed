package domain

type Profile struct {
	UserID         string   `json:"user_id"`
	DisplayName    string   `json:"display_name"`
	Bio            string   `json:"bio"`
	Headline       string   `json:"headline"`
	Education      string   `json:"education"`
	Occupation     string   `json:"occupation"`
	Location       string   `json:"location"`
	WebsiteURL     string   `json:"website_url"`
	AvatarObjectID string   `json:"avatar_object_id"`
	CoverObjectID  string   `json:"cover_object_id"`
	Topics         []string `json:"topics"`
	FollowingIDs   []string `json:"following_ids"`
}
