package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"time"
)

var (
	ErrRowsAffected   = errors.New("rows affected != 1")
	ErrNoRowsAffected = errors.New("zero rows affected")
)

type TokenProvider struct {
	jwtMaker          *helpers.JWTMaker
	refreshTokensRepo *repositories.RefreshTokensRepository
}

func NewTokenProvider(jwt *helpers.JWTMaker, repo *repositories.RefreshTokensRepository) *TokenProvider {
	return &TokenProvider{
		jwtMaker:          jwt,
		refreshTokensRepo: repo,
	}
}

// IssuePair генерирует access token и refresh token и сохраняет обновленный refresh token.
// Работает с DBTX, что позволяет использовать и транзакцию, и пул.
func (p *TokenProvider) IssuePair(ctx context.Context, db repositories.DBTX, userId, roleLevel int64) (models.GetTokensResponse, error) {
	// генерируем access token
	accessToken, expUnix, err := p.jwtMaker.Issue(userId, roleLevel)
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	// удаляем предыдущие refresh токены
	_, err = p.refreshTokensRepo.WithDB(db).DeleteRefreshTokens(ctx, userId)
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	// создаем новый refresh token
	rawRefreshToken, err := helpers.NewRefreshToken()
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	hash := helpers.HashToken(rawRefreshToken)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	tag, err := p.refreshTokensRepo.WithDB(db).CreateRefreshToken(ctx, userId, hash, expiresAt)
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	if tag.RowsAffected() != 1 {
		return models.GetTokensResponse{}, ErrRowsAffected
	}

	return models.GetTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpUnix:      expUnix,
	}, nil
}
