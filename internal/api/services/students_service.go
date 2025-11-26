package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

var (
	ErrStudentByBookIdNotFound = errors.New("не удалось найти пользователя")
	ErrUserAlreadyExists       = errors.New("пользователь с таким номером зачетки уже существует")
	ErrTokenGenerationError    = errors.New("ошибка при генерации токена")
	ErrUpsetLinkToken          = errors.New("failed to upsert link token")
)

type StudentsService struct {
	studentsRepo *repositories.StudentsRepository
	db           *pgxpool.Pool
}

func NewStudentsService(repository *repositories.StudentsRepository, db *pgxpool.Pool) *StudentsService {
	return &StudentsService{
		studentsRepo: repository,
		db:           db,
	}
}

func (s *StudentsService) CheckStudent(ctx context.Context, bookId int64) (models.AuthStatus, error) {
	student, err := s.studentsRepo.GetStudentByBookId(ctx, bookId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.AuthStatus{}, ErrStudentByBookIdNotFound
		}
	}

	exists, err := s.studentsRepo.UserExistsByBookId(ctx, bookId)
	if err != nil {
		return models.AuthStatus{}, err
	}
	if exists {
		return models.AuthStatus{}, ErrUserAlreadyExists
	}

	rawToken, err := helpers.GenerateTokenRaw(32)
	if err != nil {
		return models.AuthStatus{}, ErrTokenGenerationError
	}
	tokenHash := helpers.HashToken(rawToken)
	expiresAt := time.Now().Add(helpers.LinkTokenTTL)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.AuthStatus{}, err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := s.studentsRepo.UpsertLinkTokenTx(ctx, tx, bookId, tokenHash, expiresAt)
	if err != nil {
		return models.AuthStatus{}, err
	}
	if tag.RowsAffected() != 1 {
		return models.AuthStatus{}, ErrUpsetLinkToken
	}

	if err := tx.Commit(ctx); err != nil {
		return models.AuthStatus{}, err
	}

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
