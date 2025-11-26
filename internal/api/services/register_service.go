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
	ErrUserWithEmailAlreadyExists = errors.New("пользователь с такой почтой уже существует")
	ErrTokenNotFound              = errors.New("токен не найден или истёк")
)

type RegisterService struct {
	userRepo      *repositories.UserRepository
	authRepo      *repositories.StudentsRepository
	tokenProvider *TokenProvider
	db            *pgxpool.Pool
}

func NewRegisterService(userRepo *repositories.UserRepository, tokenProvider *TokenProvider, authRepo *repositories.StudentsRepository, db *pgxpool.Pool) *RegisterService {
	return &RegisterService{
		userRepo:      userRepo,
		tokenProvider: tokenProvider,
		authRepo:      authRepo,
		db:            db,
	}
}

func (s *RegisterService) RegisterUser(ctx context.Context, email, password, token string) (models.RegisterResponse, error) {

	exists, err := s.userRepo.CheckUserWithEmailExists(ctx, email)
	if err != nil {
		return models.RegisterResponse{}, err
	}
	if exists {
		return models.RegisterResponse{}, ErrUserWithEmailAlreadyExists
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.RegisterResponse{}, err
	}

	defer func() {
		if er := tx.Rollback(ctx); er != nil && !errors.Is(er, pgx.ErrTxClosed) {
			log.Println("rollback failed:", er)
		}
	}()

	bookId, err := s.authRepo.DeleteLinkTokenTx(ctx, tx, helpers.HashToken(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RegisterResponse{}, ErrTokenNotFound
		}
		return models.RegisterResponse{}, err
	}

	var student models.Student
	student, err = s.userRepo.GetStudentByBookId(ctx, bookId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.RegisterResponse{}, ErrStudentByBookIdNotFound
		}
		return models.RegisterResponse{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // cost=10 по умолчанию
	if err != nil {
		return models.RegisterResponse{}, err
	}
	user := models.User{
		BookId:       student.BookId,
		Name:         student.Name,
		Surname:      student.Surname,
		MiddleName:   student.MiddleName,
		BirthDate:    student.BirthDate,
		StudentGroup: student.StudentGroup,
		Password:     hash,
		Email:        email,
		RoleLevel:    10,
	}

	userId, err := s.userRepo.CreateUserTx(ctx, tx, user)
	if err != nil {
		return models.RegisterResponse{}, err
	}

	tokens, err := s.tokenProvider.IssuePairTx(ctx, tx, userId, user.RoleLevel)
	if err != nil {
		return models.RegisterResponse{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return models.RegisterResponse{}, err
	}

	resp := models.RegisterResponse{
		AuthTokens: models.AuthTokens{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpUnix:      tokens.ExpUnix,
		},
		UserSubstructure: models.UserSubstructure{
			ID:           userId,
			Email:        user.Email,
			BookId:       user.BookId,
			FirstName:    user.Name,
			RoleLevel:    user.RoleLevel,
			StudentGroup: user.StudentGroup,
		},
	}
	return resp, nil

}
