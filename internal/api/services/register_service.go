package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserWithEmailAlreadyExists = errors.New("пользователь с такой почтой уже существует")
	ErrTokenNotFound              = errors.New("токен не найден или истек")
	ErrStudentByBookIdNotFound    = errors.New("студент с таким book id не найден")
)

type RegisterService struct {
	userRepo      *repositories.UserRepository
	authRepo      *repositories.StudentsRepository
	tokenProvider *TokenProvider
	uow           *repositories.UoW
}

func NewRegisterService(
	userRepo *repositories.UserRepository,
	authRepo *repositories.StudentsRepository,
	tokenProvider *TokenProvider,
	uow *repositories.UoW,
) *RegisterService {
	return &RegisterService{
		userRepo:      userRepo,
		authRepo:      authRepo,
		tokenProvider: tokenProvider,
		uow:           uow,
	}
}

func (s *RegisterService) RegisterUser(ctx context.Context, email, password, token string) (models.RegisterResponse, error) {
	// проверяем, есть ли пользователь с такой почтой
	exists, err := s.userRepo.CheckUserWithEmailExists(ctx, email)
	if err != nil {
		return models.RegisterResponse{}, err
	}
	if exists {
		return models.RegisterResponse{}, ErrUserWithEmailAlreadyExists
	}

	var result models.RegisterResponse

	// выполняем атомарную логику через UoW
	err = s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		// удаляем токен привязки и получаем bookId
		bookId, err := s.authRepo.WithDB(tx).DeleteLinkToken(ctx, helpers.HashToken(token))
		if err != nil {
			return err
		}
		if bookId == 0 {
			return ErrTokenNotFound
		}

		// получаем студента
		student, err := s.userRepo.GetStudentByBookId(ctx, bookId)
		if err != nil {
			return ErrStudentByBookIdNotFound
		}

		// хешируем пароль
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user := models.User{
			BookId:       student.BookId,
			Name:         student.Name,
			Surname:      student.Surname,
			MiddleName:   student.MiddleName,
			BirthDate:    student.BirthDate,
			StudentGroup: student.StudentGroup,
			Password:     hashed,
			Email:        email,
			RoleLevel:    10,
		}

		// создаем пользователя
		userId, err := s.userRepo.WithDB(tx).CreateUser(ctx, user)
		if err != nil {
			return err
		}

		// генерируем пару токенов
		tokens, err := s.tokenProvider.IssuePair(ctx, tx, userId, user.RoleLevel)
		if err != nil {
			return err
		}

		result = models.RegisterResponse{
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

		return nil
	})

	return result, err
}
