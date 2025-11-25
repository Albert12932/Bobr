package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var (
	ErrInvalidPassword = errors.New("неправильный пароль")
	ErrUserNotFound    = errors.New("пользователь не найден")
)

type LoginService struct {
	refreshTokensRepo *repositories.RefreshTokensRepository
	userRepo          *repositories.UserRepository
	tokenProvider     *TokenProvider
	jwtMaker          *helpers.JWTMaker
	db                *pgxpool.Pool
}

func NewLoginService(tokenProvider *TokenProvider, userRepo *repositories.UserRepository, db *pgxpool.Pool) *LoginService {
	return &LoginService{
		userRepo:      userRepo,
		tokenProvider: tokenProvider,
		db:            db,
	}
}

func (s *LoginService) Login(ctx context.Context, email string, password string) (models.LoginResponse, error) {

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.LoginResponse{}, ErrUserNotFound
		}
		return models.LoginResponse{}, err
	}

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return models.LoginResponse{}, ErrInvalidPassword
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.LoginResponse{}, err
	}

	defer func() {
		if er := tx.Rollback(ctx); er != nil && !errors.Is(er, pgx.ErrTxClosed) {
			log.Println("rollback failed:", er)
		}
	}()

	tokens, err := s.tokenProvider.IssuePairTx(ctx, tx, user.Id, user.RoleLevel)
	if err != nil {
		return models.LoginResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.LoginResponse{}, err
	}

	return models.LoginResponse{
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
	}, nil
}
