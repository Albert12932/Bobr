package repositories

import (
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

// StudentsRepository отвечает за работу с таблицей students и связанными токенами.
type StudentsRepository struct {
	db DBTX
}

// NewStudentsRepository создает новый экземпляр StudentsRepository.
func NewStudentsRepository(db DBTX) *StudentsRepository {
	return &StudentsRepository{db: db}
}

// WithDB возвращает копию репозитория, использующую указанный DBTX (tx или pool).
func (r *StudentsRepository) WithDB(db DBTX) *StudentsRepository {
	return &StudentsRepository{db: db}
}

// GetStudentByBookId возвращает студента по номеру зачетной книжки.
func (r *StudentsRepository) GetStudentByBookId(ctx context.Context, bookId int64) (models.Student, error) {
	var student models.Student

	err := r.db.QueryRow(ctx,
		`SELECT id, book_id, surname, name, middle_name, birth_date, student_group
         FROM students
         WHERE book_id = $1`,
		bookId,
	).Scan(
		&student.Id,
		&student.BookId,
		&student.Surname,
		&student.Name,
		&student.MiddleName,
		&student.BirthDate,
		&student.StudentGroup,
	)

	return student, err
}

// UserExistsByBookId проверяет, существует ли пользователь с указанным book_id.
func (r *StudentsRepository) UserExistsByBookId(ctx context.Context, bookId int64) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE book_id = $1)`,
		bookId,
	).Scan(&exists)

	return exists, err
}

// UpsertLinkToken создает или обновляет токен привязки аккаунта к студенту.
func (r *StudentsRepository) UpsertLinkToken(ctx context.Context, bookId int64, token []byte, expiresAt time.Time) (*pgconn.CommandTag, error) {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO link_tokens (book_id, token_hash, expires_at, created_at)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (book_id)
         DO UPDATE SET token_hash = EXCLUDED.token_hash,
                       expires_at = EXCLUDED.expires_at,
                       created_at = NOW()`,
		bookId,
		token,
		expiresAt,
	)

	return &tag, err
}

// DeleteLinkToken удаляет токен привязки по хешу токена и возвращает book_id студента.
func (r *StudentsRepository) DeleteLinkToken(ctx context.Context, tokenHash []byte) (int64, error) {
	var bookId int64

	err := r.db.QueryRow(ctx,
		`DELETE FROM link_tokens
         WHERE token_hash = $1 AND expires_at > NOW()
         RETURNING book_id`,
		tokenHash,
	).Scan(&bookId)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}

	return bookId, err
}

// GetAllStudents возвращает полный список студентов.
func (r *StudentsRepository) GetAllStudents(ctx context.Context) ([]models.Student, error) {
	var students []models.Student

	err := pgxscan.Select(ctx, r.db, &students,
		`SELECT id, book_id, surname, name, middle_name, birth_date, student_group
         FROM students`,
	)

	return students, err
}
