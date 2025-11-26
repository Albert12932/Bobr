package services

import (
	"bobri/internal/api/repositories"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var (
	ErrInvalidResetToken = errors.New("невалидный или истёкший токен")
	ErrWeakPassword      = errors.New("слишком слабый пароль")
)

type ResetPasswordService struct {
	repo          *repositories.ResetPasswordRepository
	emailProvider *EmailProvider
	db            *pgxpool.Pool
}

func NewResetPasswordService(
	repo *repositories.ResetPasswordRepository,
	emailProvider *EmailProvider,
	db *pgxpool.Pool,
) *ResetPasswordService {
	return &ResetPasswordService{
		repo:          repo,
		emailProvider: emailProvider,
		db:            db,
	}
}

func (s *ResetPasswordService) ResetPassword(ctx context.Context, email string) error {
	userId, err := s.repo.GetUserIdByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	rawToken, err := helpers.GenerateTokenRaw(32)
	if err != nil {
		return errors.New("ошибка генерации токена")
	}

	tokenHash := helpers.HashToken(rawToken)
	expiresAt := time.Now().Add(15 * time.Minute)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if er := tx.Rollback(ctx); er != nil && !errors.Is(er, pgx.ErrTxClosed) {
			log.Println("rollback failed:", er)
		}
	}()

	err = s.repo.UpsertResetToken(ctx, tx, userId, email, tokenHash, expiresAt)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return s.emailProvider.SendResetPassword(email, rawToken)
}

func (s *ResetPasswordService) SetNewPassword(ctx context.Context, token string, newPassword string) error {

	tokenHash := helpers.HashToken(token)

	userId, err := s.repo.GetUserIdByResetToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return err
	}

	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if er := tx.Rollback(ctx); er != nil && !errors.Is(er, pgx.ErrTxClosed) {
			log.Println("rollback failed:", er)
		}
	}()

	_, err = tx.Exec(ctx, `
		UPDATE users SET password = $1 WHERE id = $2
	`, hashedPassword, userId)
	if err != nil {
		return err
	}

	err = s.repo.DeleteResetTokenTx(ctx, tx, userId)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
