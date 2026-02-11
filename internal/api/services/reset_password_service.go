package services

import (
	"bobri/internal/api/repositories"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidResetToken = errors.New("невалидный или истекший токен")
	ErrWeakPassword      = errors.New("слишком слабый пароль")
)

type ResetPasswordService struct {
	resetRepo     *repositories.ResetPasswordRepository
	userRepo      *repositories.UserRepository
	emailProvider *EmailProvider
	uow           *repositories.UoW
}

func NewResetPasswordService(
	resetRepo *repositories.ResetPasswordRepository,
	userRepo *repositories.UserRepository,
	emailProvider *EmailProvider,
	uow *repositories.UoW,
) *ResetPasswordService {
	return &ResetPasswordService{
		resetRepo:     resetRepo,
		userRepo:      userRepo,
		emailProvider: emailProvider,
		uow:           uow,
	}
}

func (s *ResetPasswordService) ResetPassword(ctx context.Context, email string) error {
	userId, err := s.resetRepo.GetUserIdByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	rawToken, err := helpers.GenerateTokenRaw(32)
	if err != nil {
		return errors.New("ошибка генерации токена")
	}

	tokenHash := helpers.HashToken(rawToken)
	expiresAt := time.Now().Add(15 * time.Minute)

	// выполняем в транзакции
	err = s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		return s.resetRepo.WithDB(tx).UpsertResetToken(ctx, userId, email, tokenHash, expiresAt)
	})
	if err != nil {
		return err
	}

	fmt.Println("token:", rawToken)
	return s.emailProvider.SendResetPassword(email, rawToken)

}
func (s *ResetPasswordService) SetNewPassword(ctx context.Context, token string, newPassword string) error {
	tokenHash := helpers.HashToken(token)

	userId, err := s.resetRepo.GetUserIdByResetToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return err
	}

	if userId == 0 {
		return ErrInvalidResetToken
	}

	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		// обновляем пароль
		err := s.userRepo.WithDB(tx).UpdatePassword(ctx, userId, hashedPassword)
		if err != nil {
			return err
		}

		// удаляем токен
		err = s.resetRepo.WithDB(tx).DeleteResetToken(ctx, userId)
		if err != nil {
			return err
		}

		return nil
	})
}
