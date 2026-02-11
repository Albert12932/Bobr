package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// RefreshTokensRepository - репозиторий для работы с refresh токенами.
// Использует абстракцию DBTX, что позволяет работать как с пулом,
// так и с транзакцией через repo.WithDB(tx).
type RefreshTokensRepository struct {
	db DBTX
}

// NewRefreshTokensRepository - конструктор репозитория.
func NewRefreshTokensRepository(db DBTX) *RefreshTokensRepository {
	return &RefreshTokensRepository{db: db}
}

// WithDB - возвращает копию репозитория с новым DBTX (tx или pool).
func (r *RefreshTokensRepository) WithDB(db DBTX) *RefreshTokensRepository {
	return &RefreshTokensRepository{db: db}
}

// GetUserIdByRefreshToken - возвращает ID пользователя по хешу refresh-токена.
// Если запись не найдена - возвращает userId = 0 и err = nil.
func (r *RefreshTokensRepository) GetUserIdByRefreshToken(ctx context.Context, refreshTokenHash []byte) (int64, error) {
	var userId int64

	err := r.db.QueryRow(ctx,
		`SELECT user_id
         FROM refresh_tokens
         WHERE token_hash = $1 AND expires_at > NOW()`,
		refreshTokenHash,
	).Scan(&userId)

	// Обработка ошибок
	if err != nil {
		return 0, fmt.Errorf("could not get user id: %w", err)
	}

	return userId, nil
}

// GetRoleLevelByUserId - возвращает уровень роли пользователя.
func (r *RefreshTokensRepository) GetRoleLevelByUserId(ctx context.Context, userId int64) (int64, error) {
	var level int64

	err := r.db.QueryRow(ctx,
		`SELECT role_level
         FROM users
         WHERE id = $1`,
		userId,
	).Scan(&level)

	if err != nil {
		return 0, fmt.Errorf("could not get user role_level: %w", err)
	}

	return level, nil
}

// UpdateRefreshToken - обновляет refresh-токен пользователя (обычно внутри транзакции).
func (r *RefreshTokensRepository) UpdateRefreshToken(ctx context.Context, userId int64, newRefreshToken []byte) (pgconn.CommandTag, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens
         SET token_hash = $1, expires_at = $2
         WHERE user_id = $3`,
		newRefreshToken,
		time.Now().Add(30*24*time.Hour),
		userId,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return pgconn.CommandTag{}, fmt.Errorf("could not update refresh token: %w", err)
	}
	return tag, nil
}

// DeleteRefreshTokens - удаляет все refresh-токены пользователя (обычно внутри транзакции).
func (r *RefreshTokensRepository) DeleteRefreshTokens(ctx context.Context, userId int64) (pgconn.CommandTag, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM refresh_tokens
         WHERE user_id = $1`,
		userId,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return pgconn.CommandTag{}, fmt.Errorf("could not delete refresh token: %w", err)
	}
	return tag, nil
}

// CreateRefreshToken - создаёт новый refresh-токен (обычно внутри транзакции).
func (r *RefreshTokensRepository) CreateRefreshToken(ctx context.Context, userId int64, refreshTokenHash []byte, expiresAt time.Time) (pgconn.CommandTag, error) {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
         VALUES ($1, $2, $3)`,
		userId,
		refreshTokenHash,
		expiresAt,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return pgconn.CommandTag{}, fmt.Errorf("could not create refresh token: %w", err)
	}
	return tag, nil
}
