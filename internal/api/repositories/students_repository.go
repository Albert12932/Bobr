package repositories

import (
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type StudentsRepository struct {
	db *pgxpool.Pool
}

func NewStudentsRepository(db *pgxpool.Pool) *StudentsRepository {
	return &StudentsRepository{db: db}
}

func (r *StudentsRepository) GetStudentByBookId(ctx context.Context, bookId int64) (models.Student, error) {
	var student models.Student

	err := r.db.QueryRow(ctx, `SELECT id, book_id, surname, name, middle_name, birth_date, student_group
			   FROM students WHERE book_id = $1`,
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

func (r *StudentsRepository) UserExistsByBookId(ctx context.Context, bookId int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE book_id = $1)`,
		bookId,
	).Scan(&exists)
	return exists, err
}

func (r *StudentsRepository) UpsertLinkTokenTx(ctx context.Context, tx pgx.Tx, bookId int64, token []byte, expiresAt time.Time) (pgconn.CommandTag, error) {
	tag, err := tx.Exec(ctx, `
			INSERT INTO link_tokens (book_id, token_hash, expires_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (book_id)
			DO UPDATE SET token_hash = EXCLUDED.token_hash,
			              expires_at = EXCLUDED.expires_at,
			              created_at = now()
		`, bookId, token, expiresAt)

	return tag, err
}

func (r *StudentsRepository) DeleteLinkTokenTx(ctx context.Context, tx pgx.Tx, tokenHash []byte) (int64, error) {
	var bookId int64
	err := tx.QueryRow(ctx, `DELETE FROM link_tokens
       WHERE token_hash = $1 AND expires_at > now()
       RETURNING book_id`, tokenHash).Scan(&bookId)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return bookId, err
}

// GetAllStudents возвращает всех студентов
func (r *StudentsRepository) GetAllStudents(ctx context.Context) ([]models.Student, error) {
	var students []models.Student
	err := pgxscan.Select(ctx, r.db, &students,
		`SELECT id, book_id, surname, name, middle_name, birth_date, student_group FROM students`)
	return students, err
}
