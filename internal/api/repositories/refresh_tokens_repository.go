package repositories

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type RefreshTokensRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokensRepository(db *pgxpool.Pool) *RefreshTokensRepository {
	return &RefreshTokensRepository{
		db: db,
	}
}

func (r *RefreshTokensRepository) GetUserIdByRefreshToken(ctx context.Context, refreshTokenHash []byte) (int64, error) {
	var userId int64
	err := r.db.QueryRow(ctx, "SELECT user_id from refresh_tokens where token_hash = $1 and expires_at > now()", refreshTokenHash).Scan(&userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return userId, err
}

func (r *RefreshTokensRepository) GetRoleLevelByUserId(ctx context.Context, userId int64) (int64, error) {
	var roleLevel int64
	err := r.db.QueryRow(ctx, "SELECT role_level from users where id = $1", userId).Scan(&roleLevel)
	return roleLevel, err
}
func (r *RefreshTokensRepository) UpdateRefreshTokenTx(ctx context.Context, tx pgx.Tx, userId int64, newRefreshToken []byte) (pgconn.CommandTag, error) {
	return tx.Exec(ctx, `
			UPDATE refresh_tokens
			SET token_hash = $1, expires_at = $2
			WHERE user_id = $3
		`, newRefreshToken, time.Now().Add(30*24*time.Hour), userId) // 30 дней
}

func (r *RefreshTokensRepository) DeleteRefreshTokensTx(ctx context.Context, tx pgx.Tx, userId int64) (pgconn.CommandTag, error) {
	return tx.Exec(ctx, `
			DELETE FROM refresh_tokens
			WHERE user_id = $1
		`, userId)
}

func (r *RefreshTokensRepository) CreateRefreshTokenTx(ctx context.Context, tx pgx.Tx, userId int64, refreshTokenHash []byte, expiresAt time.Time) (pgconn.CommandTag, error) {
	return tx.Exec(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3)
		`, userId, refreshTokenHash, expiresAt)
}
