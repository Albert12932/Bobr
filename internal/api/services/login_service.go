package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("неправильный пароль")
	ErrUserNotFound    = errors.New("пользователь не найден")
)

type LoginService struct {
	userRepo      *repositories.UserRepository
	tokenProvider *TokenProvider
	uow           *repositories.UoW
}

func NewLoginService(
	userRepo *repositories.UserRepository,
	tokenProvider *TokenProvider,
	uow *repositories.UoW,
) *LoginService {
	return &LoginService{
		userRepo:      userRepo,
		tokenProvider: tokenProvider,
		uow:           uow,
	}
}

func (s *LoginService) Login(ctx context.Context, email string, password string) (models.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return models.LoginResponse{}, ErrUserNotFound
	}

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return models.LoginResponse{}, ErrInvalidPassword
	}

	var result models.LoginResponse

	err = s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		// генерируем и сохраняем токены
		tokens, err := s.tokenProvider.IssuePair(ctx, tx, user.Id, user.RoleLevel)
		if err != nil {
			return err
		}

		result = models.LoginResponse{
			AuthTokens: models.AuthTokens{
				AccessToken:  tokens.AccessToken,
				RefreshToken: tokens.RefreshToken,
				ExpUnix:      tokens.ExpUnix,
			},
			UserSubstructure: models.UserSubstructure{
				ID:           user.Id,
				Email:        user.Email,
				BookId:       user.BookId,
				FirstName:    user.Name,
				RoleLevel:    user.RoleLevel,
				StudentGroup: user.StudentGroup,
			},
		}

		return nil
	})

	return result, err
}
