package repositories

import (
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := pgxscan.Get(ctx, r.db, &user, `
		SELECT id, coalesce(book_id, 0) as book_id, name, surname, password, email, role_level
		FROM users
		WHERE email = $1
		`, email)
	return user, err
}
func (r *UserRepository) GetStudentByBookId(ctx context.Context, bookId int64) (models.Student, error) {
	var student models.Student
	err := pgxscan.Get(ctx, r.db, &student, `
		SELECT id, book_id, surname, name, middle_name, birth_date, student_group
		FROM students
		WHERE book_id = $1
		`, bookId)
	return student, err
}

func (r *UserRepository) CreateUserTx(ctx context.Context, tx pgx.Tx, user models.User) (int64, error) {
	var userID int64
	err := tx.QueryRow(ctx, `
			INSERT INTO users (book_id, name, surname, middle_name, student_group, birth_date, password, email, role_level)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`,
		user.BookId, user.Name, user.Surname, user.MiddleName, user.StudentGroup, user.BirthDate,
		user.Password, user.Email, user.RoleLevel).Scan(&userID)
	return userID, err
}

func (r *UserRepository) CheckUserWithEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
		`, email).Scan(&exists)
	return exists, err
}
func (r *UserRepository) DeleteUserByEmail(ctx context.Context, email string) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM users WHERE email = $1`, email)

	if err != nil {
		return 0, err
	}

	return tag.RowsAffected(), nil
}
func (r *UserRepository) GetUsersWithMaxRole(ctx context.Context, maxRole int64) ([]models.User, error) {
	var users []models.User

	err := pgxscan.Select(ctx, r.db, &users, `
		SELECT 
			id,
			COALESCE(book_id, 0) as book_id,
			surname,
			name,
			middle_name,
			COALESCE(birth_date, '1970-01-01'::timestamp) as birth_date,
			COALESCE(student_group, '') as student_group,
			password,
			email,
			role_level
		FROM users
		WHERE role_level <= $1
	`, maxRole)

	return users, err
}
func (r *UserRepository) GetProfileByUserID(ctx context.Context, userID int64) (models.ProfileResponse, error) {
	var profile models.ProfileResponse

	err := pgxscan.Get(ctx, r.db, &profile, `
		SELECT 
			COALESCE(book_id, 0) AS book_id,
			name,
			surname,
			middle_name,
			COALESCE(birth_date, TO_DATE('1970-01-01','YYYY-MM-DD')) AS birth_date,
			COALESCE(student_group, '') AS student_group,
			email,
			role_level
		FROM users
		WHERE id = $1
	`, userID)

	points, err := r.GetUserPoints(ctx, userID)
	if err != nil {
		return models.ProfileResponse{}, err
	}

	profile.TotalPoints = points

	return profile, err
}
func (r *UserRepository) GetUserRoleAndBookId(ctx context.Context, userId int64) (int64, int64, error) {
	var roleLevel int64
	var bookId int64
	err := r.db.QueryRow(ctx,
		"SELECT role_level, COALESCE(book_id, 0) FROM users WHERE id = $1",
		userId,
	).Scan(&roleLevel, &bookId)
	return roleLevel, bookId, err
}

func (r *UserRepository) RoleExists(ctx context.Context, roleLevel int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM roles WHERE level = $1)",
		roleLevel,
	).Scan(&exists)
	return exists, err
}

func (r *UserRepository) UpdateUser(ctx context.Context, sqlQuery string, args []interface{}) (int64, error) {
	tag, err := r.db.Exec(ctx, sqlQuery, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
func (r *UserRepository) GetUserPoints(ctx context.Context, userId int64) (int64, error) {
	var points int64
	err := r.db.QueryRow(ctx, `
        SELECT total_points
        FROM user_points
        WHERE user_id = $1
    `, userId).Scan(&points)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil // у нового пользователя 0 баллов
	}
	return points, err
}
func (r *UserRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.UserWithPoints, error) {
	var users []models.UserWithPoints

	query := `
        SELECT
            u.id AS user_id,
            u.name,
            u.surname,
            up.total_points,
            ROW_NUMBER() OVER (ORDER BY up.total_points DESC) AS position
        FROM user_points up
        JOIN users u ON up.user_id = u.id
        ORDER BY up.total_points DESC
        LIMIT $1
    `

	err := pgxscan.Select(ctx, r.db, &users, query, limit)
	return users, err
}
