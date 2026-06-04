package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func (r *PostgresRepository) ListByAuthor(ctx context.Context, authorID string, limit int, cursor time.Time) ([]domain.Article, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if cursor.IsZero() {
		cursor = time.Now().UTC().Add(100 * 365 * 24 * time.Hour)
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, author_id, title, slug, content, category, tags, status, created_at, updated_at
		FROM articles
		WHERE author_id = $1 AND status = 'published' AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3
	`, authorID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	articles := make([]domain.Article, 0)
	for rows.Next() {
		var article domain.Article
		if err := rows.Scan(&article.ID, &article.AuthorID, &article.Title, &article.Slug, &article.Content, &article.Category, &article.Tags, &article.Status, &article.CreatedAt, &article.UpdatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, rows.Err()
}

func (r *PostgresRepository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	if comment.ID == "" {
		comment.ID = newID()
	}
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, `
		INSERT INTO article_comments (id, article_id, user_id, parent_id, body, created_at)
		VALUES ($1,$2,$3,NULLIF($4, '')::uuid,$5,$6)
		RETURNING id, article_id, user_id, COALESCE(parent_id::text, ''), body, edited_at, deleted_at, created_at
	`, comment.ID, comment.ArticleID, comment.UserID, comment.ParentID, comment.Body, comment.CreatedAt).Scan(
		&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID, &comment.Body, &comment.EditedAt, &comment.DeletedAt, &comment.CreatedAt,
	)
	return comment, err
}

func (r *PostgresRepository) ListComments(ctx context.Context, articleID string, limit int, cursor time.Time) ([]domain.Comment, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if cursor.IsZero() {
		cursor = time.Now().UTC().Add(100 * 365 * 24 * time.Hour)
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, article_id, user_id, COALESCE(parent_id::text, ''), body, edited_at, deleted_at, created_at
		FROM article_comments
		WHERE article_id = $1 AND created_at < $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3
	`, articleID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var comment domain.Comment
		if err := rows.Scan(&comment.ID, &comment.ArticleID, &comment.UserID, &comment.ParentID, &comment.Body, &comment.EditedAt, &comment.DeletedAt, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (r *PostgresRepository) ShareArticle(ctx context.Context, share domain.Share) (domain.Share, error) {
	if share.ID == "" {
		share.ID = newID()
	}
	if share.Visibility == "" {
		share.Visibility = "public"
	}
	if share.CreatedAt.IsZero() {
		share.CreatedAt = time.Now().UTC()
	}
	err := r.db.QueryRow(ctx, `
		INSERT INTO article_shares (id, article_id, user_id, caption, visibility, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, article_id, user_id, caption, visibility, created_at
	`, share.ID, share.ArticleID, share.UserID, share.Caption, share.Visibility, share.CreatedAt).Scan(
		&share.ID, &share.ArticleID, &share.UserID, &share.Caption, &share.Visibility, &share.CreatedAt,
	)
	return share, err
}

func (r *PostgresRepository) ListSharesByUser(ctx context.Context, userID string, limit int, cursor time.Time) ([]domain.Share, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if cursor.IsZero() {
		cursor = time.Now().UTC().Add(100 * 365 * 24 * time.Hour)
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, article_id, user_id, caption, visibility, created_at
		FROM article_shares
		WHERE user_id = $1 AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, cursor, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	shares := make([]domain.Share, 0)
	for rows.Next() {
		var share domain.Share
		if err := rows.Scan(&share.ID, &share.ArticleID, &share.UserID, &share.Caption, &share.Visibility, &share.CreatedAt); err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}
	return shares, rows.Err()
}

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf[0:4]) + "-" + hex.EncodeToString(buf[4:6]) + "-" + hex.EncodeToString(buf[6:8]) + "-" + hex.EncodeToString(buf[8:10]) + "-" + hex.EncodeToString(buf[10:])
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
		INSERT INTO discussion_rooms (id, article_id, name, locked, archived_at)
		VALUES ($1,$2,$3,FALSE,NULL)
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
