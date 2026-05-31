package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/media-service/internal/media/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, object domain.Object) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO media_objects (id, owner_id, bucket, object_key, content_type, size_bytes, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, object.ID, object.OwnerID, object.Bucket, object.Key, object.ContentType, object.Size, object.CreatedAt)
	return err
}
