package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoRowsAffected = errors.New("no rows affected")
	ErrNoTokensFound  = errors.New("no tokens found")
	ErrRowsAffected   = errors.New("rows affected is not equal to 1")
)

type RefreshTokensService struct {
	refreshTokensRepo *repositories.RefreshTokensRepository
	jwtMaker          *helpers.JWTMaker
	db                *pgxpool.Pool
}

func NewRefreshTokensService(db *pgxpool.Pool, refreshTokensRepo *repositories.RefreshTokensRepository, jwtMaker *helpers.JWTMaker) *RefreshTokensService {
	return &RefreshTokensService{
		refreshTokensRepo: refreshTokensRepo,
		jwtMaker:          jwtMaker,
		db:                db,
	}
}

func (s *RefreshTokensService) RefreshToken(ctx context.Context, refreshToken string) (models.RefreshTokenResponse, error) {
	userId, err := s.refreshTokensRepo.GetUserIdByRefreshToken(ctx, helpers.HashToken(refreshToken))
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}
	if userId == 0 {
		return models.RefreshTokenResponse{}, ErrNoTokensFound
	}

	roleLevel, err := s.refreshTokensRepo.GetRoleLevelByUserId(ctx, userId)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// Генерируем новый access token
	accessToken, expiresAt, err := s.jwtMaker.Issue(userId, roleLevel)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// Генерируем новый refresh token
	newRefreshToken, err := helpers.NewRefreshToken()
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// Начинаем транзакцию
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}

	defer func() {
		if er := tx.Rollback(ctx); er != nil && !errors.Is(er, pgx.ErrTxClosed) {
			// логируем, но не ломаем успешный путь
		}
	}()

	// Обновляем refresh token
	tag, err := s.refreshTokensRepo.UpdateRefreshTokenTx(
		ctx,
		tx,
		userId,
		helpers.HashToken(newRefreshToken),
	)
	if err != nil {
		return models.RefreshTokenResponse{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.RefreshTokenResponse{}, ErrNoRowsAffected
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return models.RefreshTokenResponse{}, err
	}

	// Формируем ответ
	return models.RefreshTokenResponse{
		UserID: userId,
		AuthTokens: models.AuthTokens{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpUnix:      expiresAt,
		},
	}, nil
}
