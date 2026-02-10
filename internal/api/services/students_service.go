package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUserAlreadyExists    = errors.New("пользователь с таким номером зачетки уже существует")
	ErrTokenGenerationError = errors.New("ошибка при генерации токена")
	ErrUpsetLinkToken       = errors.New("не удалось сохранить link token")
)

type StudentsService struct {
	studentsRepo *repositories.StudentsRepository
	uow          *repositories.UoW
}

func NewStudentsService(repo *repositories.StudentsRepository, uow *repositories.UoW) *StudentsService {
	return &StudentsService{
		studentsRepo: repo,
		uow:          uow,
	}
}

func (s *StudentsService) CheckStudent(ctx context.Context, bookId int64) (models.AuthStatus, error) {
	// проверяем, есть ли студент
	student, err := s.studentsRepo.GetStudentByBookId(ctx, bookId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AuthStatus{}, ErrStudentByBookIdNotFound
		}
		return models.AuthStatus{}, err
	}

	// проверяем, есть ли уже пользователь с таким book id
	exists, err := s.studentsRepo.UserExistsByBookId(ctx, bookId)
	if err != nil {
		return models.AuthStatus{}, err
	}
	if exists {
		return models.AuthStatus{}, ErrUserAlreadyExists
	}

	// генерируем token
	rawToken, err := helpers.GenerateTokenRaw(32)
	if err != nil {
		return models.AuthStatus{}, ErrTokenGenerationError
	}

	tokenHash := helpers.HashToken(rawToken)
	expiresAt := time.Now().Add(helpers.LinkTokenTTL)

	// транзакция через UoW
	err = s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		tag, err := s.studentsRepo.WithDB(tx).UpsertLinkToken(ctx, bookId, tokenHash, expiresAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrUpsetLinkToken
		}

		return nil
	})
	if err != nil {
		return models.AuthStatus{}, err
	}

	// успешный ответ
	return models.AuthStatus{
		Status:          "free",
		DisplayName:     student.Name,
		StudentGroup:    student.StudentGroup,
		LinkToken:       rawToken,
		LinkTokenTtlSec: int64(helpers.LinkTokenTTL.Seconds()),
	}, nil
}

func (s *StudentsService) GetStudents(ctx context.Context) ([]models.Student, error) {
	return s.studentsRepo.GetAllStudents(ctx)
}
