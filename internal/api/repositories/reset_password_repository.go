package repositories

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type ResetPasswordRepository struct {
	db *pgxpool.Pool
}

func NewResetPasswordRepository(db *pgxpool.Pool) *ResetPasswordRepository {
	return &ResetPasswordRepository{db: db}
}

func (r *ResetPasswordRepository) GetUserIdByEmail(ctx context.Context, email string) (int64, error) {
	var userId int64
	err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userId)
	return userId, err
}

func (r *ResetPasswordRepository) UpsertResetToken(ctx context.Context, tx pgx.Tx, userId int64, email string, tokenHash []byte, expiresAt time.Time) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO reset_password_tokens (user_id, email, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (user_id)
		DO UPDATE SET token_hash = EXCLUDED.token_hash,
		              expires_at = EXCLUDED.expires_at,
		              created_at = now()
	`, userId, email, tokenHash, expiresAt)
	return err
}

func (r *ResetPasswordRepository) GetUserIdByResetToken(ctx context.Context, tokenHash []byte) (int64, error) {
	var userId int64
	err := r.db.QueryRow(ctx, `
		SELECT user_id 
		FROM reset_password_tokens
		WHERE token_hash = $1 AND expires_at > now()
	`, tokenHash).Scan(&userId)

	return userId, err
}

func (r *ResetPasswordRepository) DeleteResetTokenTx(ctx context.Context, tx pgx.Tx, userId int64) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM reset_password_tokens WHERE user_id = $1
	`, userId)

	return err
}
