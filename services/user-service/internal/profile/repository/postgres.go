package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/user-service/internal/profile/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) UpsertProfile(ctx context.Context, profile domain.Profile) error {
	preferences, _ := json.Marshal(map[string]any{})
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, display_name, bio, headline, education, occupation, location, website_url, avatar_object_id, cover_object_id, preferences)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9, '')::uuid,NULLIF($10, '')::uuid,$11)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			bio = EXCLUDED.bio,
			headline = EXCLUDED.headline,
			education = EXCLUDED.education,
			occupation = EXCLUDED.occupation,
			location = EXCLUDED.location,
			website_url = EXCLUDED.website_url,
			avatar_object_id = EXCLUDED.avatar_object_id,
			cover_object_id = EXCLUDED.cover_object_id,
			updated_at = now()
	`, profile.UserID, profile.DisplayName, profile.Bio, profile.Headline, profile.Education, profile.Occupation, profile.Location, profile.WebsiteURL, profile.AvatarObjectID, profile.CoverObjectID, preferences)
	return err
}

func (r *PostgresRepository) GetProfile(ctx context.Context, userID string) (domain.Profile, error) {
	var profile domain.Profile
	err := r.db.QueryRow(ctx, `
		SELECT user_id, display_name, bio, headline, education, occupation, location, website_url,
		       COALESCE(avatar_object_id::text, ''), COALESCE(cover_object_id::text, '')
		FROM user_profiles
		WHERE user_id = $1
	`, userID).Scan(&profile.UserID, &profile.DisplayName, &profile.Bio, &profile.Headline, &profile.Education, &profile.Occupation, &profile.Location, &profile.WebsiteURL, &profile.AvatarObjectID, &profile.CoverObjectID)
	return profile, err
}

func (r *PostgresRepository) UpdateAvatar(ctx context.Context, userID, objectID string) error {
	_, err := r.db.Exec(ctx, `UPDATE user_profiles SET avatar_object_id = NULLIF($2, '')::uuid, updated_at = now() WHERE user_id = $1`, userID, objectID)
	return err
}

func (r *PostgresRepository) UpdateCover(ctx context.Context, userID, objectID string) error {
	_, err := r.db.Exec(ctx, `UPDATE user_profiles SET cover_object_id = NULLIF($2, '')::uuid, updated_at = now() WHERE user_id = $1`, userID, objectID)
	return err
}

func (r *PostgresRepository) FollowUser(ctx context.Context, followerID, followedID string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO user_follows (follower_id, followed_id)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING
	`, followerID, followedID)
	return err
}

func (r *PostgresRepository) FollowTopic(ctx context.Context, userID, topic string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO followed_topics (user_id, topic)
		VALUES ($1,$2)
		ON CONFLICT DO NOTHING
	`, userID, topic)
	return err
}
