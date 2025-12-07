package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
)

var (
	ErrNoTokensFound = errors.New("no tokens found")
)

type RefreshTokensService struct {
	refreshRepo   *repositories.RefreshTokensRepository
	tokenProvider *TokenProvider
	uow           *repositories.UoW
}

func NewRefreshTokensService(
	refreshRepo *repositories.RefreshTokensRepository,
	tokenProvider *TokenProvider,
	uow *repositories.UoW,
) *RefreshTokensService {
	return &RefreshTokensService{
		refreshRepo:   refreshRepo,
		tokenProvider: tokenProvider,
		uow:           uow,
	}
}

func (s *RefreshTokensService) RefreshToken(ctx context.Context, refreshToken string) (models.RefreshTokenResponse, error) {
	// проверяем токен
	userId, err := s.refreshRepo.GetUserIdByRefreshToken(ctx, helpers.HashToken(refreshToken))
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}
	if userId == 0 {
		return models.RefreshTokenResponse{}, ErrNoTokensFound
	}

	// роль пользователя
	roleLevel, err := s.refreshRepo.GetRoleLevelByUserId(ctx, userId)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	var tokens models.GetTokensResponse

	// транзакция через UoW
	err = s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		// TokenProvider сам:
		// - удаляет старые токены
		// - генерирует новые
		// - сохраняет обновленные значения
		tokens, err = s.tokenProvider.IssuePair(ctx, tx, userId, roleLevel)
		return err
	})
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	return models.RefreshTokenResponse{
		UserID: userId,
		AuthTokens: models.AuthTokens{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpUnix:      tokens.ExpUnix,
		},
	}, nil
}
