package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, verified, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Role, user.Verified, user.CreatedAt, user.UpdatedAt)
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return domain.ErrUserAlreadyExists
	}
	return err
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, `WHERE email = $1`, strings.ToLower(email))
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.findOne(ctx, `WHERE id = $1`, id)
}

func (r *PostgresUserRepository) findOne(ctx context.Context, where string, arg any) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, password_hash, first_name, last_name, role, verified, created_at, updated_at
		FROM users `+where, arg)
	var user domain.User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName, &user.Role, &user.Verified, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

type PostgresRefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRefreshTokenRepository(db *pgxpool.Pool) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, token domain.RefreshToken) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.RevokedAt)
	return err
}

func (r *PostgresRefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, hash)
	var token domain.RefreshToken
	if err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt, &token.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRefreshNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	tag, err := r.db.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL`, id, revokedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrRefreshNotFound
	}
	return nil
}
