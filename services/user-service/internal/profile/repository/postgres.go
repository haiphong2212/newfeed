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
		INSERT INTO user_profiles (user_id, display_name, bio, preferences)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (user_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			bio = EXCLUDED.bio,
			updated_at = now()
	`, profile.UserID, profile.DisplayName, profile.Bio, preferences)
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
