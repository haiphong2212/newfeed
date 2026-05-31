package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, article domain.Article) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO articles (id, author_id, title, slug, content, category, tags, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			slug = EXCLUDED.slug,
			content = EXCLUDED.content,
			category = EXCLUDED.category,
			tags = EXCLUDED.tags,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, article.ID, article.AuthorID, article.Title, article.Slug, article.Content, article.Category, article.Tags, article.Status, article.CreatedAt, article.UpdatedAt)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO discussion_rooms (id, article_id, name)
		VALUES ($1,$2,$3)
		ON CONFLICT (article_id) DO UPDATE SET name = EXCLUDED.name
	`, article.ID, article.ID, article.DiscussionRoomName())
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*domain.Article, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, author_id, title, slug, content, category, tags, status, created_at, updated_at
		FROM articles
		WHERE id = $1
	`, id)
	var article domain.Article
	if err := row.Scan(&article.ID, &article.AuthorID, &article.Title, &article.Slug, &article.Content, &article.Category, &article.Tags, &article.Status, &article.CreatedAt, &article.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidArticle
		}
		return nil, err
	}
	return &article, nil
}
