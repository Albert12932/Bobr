package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"github.com/jackc/pgx/v5"
	"time"
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

func (p *TokenProvider) IssuePairTx(ctx context.Context, tx pgx.Tx, userId, roleLevel int64) (models.GetTokensResponse, error) {
	accessToken, exp, err := p.jwtMaker.Issue(userId, roleLevel)
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	_, err = p.refreshTokensRepo.DeleteRefreshTokensTx(ctx, tx, userId)
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	refreshToken, err := helpers.NewRefreshToken()
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	tag, err := p.refreshTokensRepo.CreateRefreshTokenTx(
		ctx,
		tx,
		userId,
		helpers.HashToken(refreshToken),
		time.Now().Add(30*24*time.Hour),
	)
	if err != nil {
		return models.GetTokensResponse{}, err
	}

	if tag.RowsAffected() != 1 {
		return models.GetTokensResponse{}, ErrRowsAffected
	}

	return models.GetTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpUnix:      exp,
	}, nil
}
