package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ResetPasswordRepository отвечает за работу с токенами сброса пароля.
// Работает через DBTX, что позволяет использовать как пул, так и транзакцию.
type ResetPasswordRepository struct {
	db DBTX
}

// NewResetPasswordRepository создает новый репозиторий.
func NewResetPasswordRepository(db DBTX) *ResetPasswordRepository {
	return &ResetPasswordRepository{db: db}
}

// WithDB возвращает копию репозитория, привязанную к новому DBTX (tx или pool).
func (r *ResetPasswordRepository) WithDB(db DBTX) *ResetPasswordRepository {
	return &ResetPasswordRepository{db: db}
}

// GetUserIdByEmail возвращает user_id по email.
func (r *ResetPasswordRepository) GetUserIdByEmail(ctx context.Context, email string) (int64, error) {
	var userId int64

	err := r.db.QueryRow(ctx,
		`SELECT id FROM users WHERE email = $1`,
		email,
	).Scan(&userId)
	if err != nil {
		return 0, fmt.Errorf("could not get user id: %w", err)
	}

	return userId, err
}

// UpsertResetToken создает или обновляет токен сброса пароля.
func (r *ResetPasswordRepository) UpsertResetToken(ctx context.Context, userId int64, email string, tokenHash []byte, expiresAt time.Time) error {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO reset_password_tokens (user_id, email, token_hash, expires_at, created_at)
         VALUES ($1, $2, $3, $4, NOW())
         ON CONFLICT (user_id)
         DO UPDATE SET token_hash = EXCLUDED.token_hash,
                       expires_at = EXCLUDED.expires_at,
                       created_at = NOW()`,
		userId,
		email,
		tokenHash,
		expiresAt,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return fmt.Errorf("could not create token: %w", err)
	}

	return nil
}

// GetUserIdByResetToken возвращает user_id по хешу токена сброса, если токен еще не истек.
func (r *ResetPasswordRepository) GetUserIdByResetToken(ctx context.Context, tokenHash []byte) (int64, error) {
	var userId int64

	err := r.db.QueryRow(ctx,
		`SELECT user_id
         FROM reset_password_tokens
         WHERE token_hash = $1 AND expires_at > NOW()`,
		tokenHash,
	).Scan(&userId)

	if err != nil || errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("could not get user id by reset token: %w", err)
	}

	return userId, nil
}

// DeleteResetToken удаляет токен сброса для пользователя.
func (r *ResetPasswordRepository) DeleteResetToken(ctx context.Context, userId int64) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM reset_password_tokens WHERE user_id = $1`,
		userId,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return fmt.Errorf("could not delete reset token: %w", err)
	}

	return nil
}
