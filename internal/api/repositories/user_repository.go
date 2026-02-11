package repositories

import (
	"bobri/internal/models"
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
)

// UserRepository отвечает за работу с таблицей users и связанными сущностями.
type UserRepository struct {
	db DBTX
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

// WithDB возвращает копию репозитория, использующую новый DBTX (tx или pool).
func (r *UserRepository) WithDB(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByEmail возвращает пользователя по email.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	err := pgxscan.Get(ctx, r.db, &user,
		`SELECT id,
		        COALESCE(book_id, 0) as book_id,
		        name,
		        surname,
		        middle_name,
		        password,
		        email,
		        role_level
		 FROM users
		 WHERE email = $1`,
		email,
	)
	if err != nil {
		return user, fmt.Errorf("could not get user: %w", err)
	}

	return user, err
}

// GetStudentByBookId возвращает студента по номеру зачетной книжки.
// Этот метод по смыслу не должен быть здесь, но оставляем для совместимости.
// В StudentsRepository он уже реализован.
func (r *UserRepository) GetStudentByBookId(ctx context.Context, bookId int64) (models.Student, error) {
	var student models.Student

	err := pgxscan.Get(ctx, r.db, &student,
		`SELECT id,
		        book_id,
		        surname,
		        name,
		        middle_name,
		        birth_date,
		        student_group
		 FROM students
		 WHERE book_id = $1`,
		bookId,
	)
	if err != nil {
		return student, fmt.Errorf("could not get student: %w", err)
	}

	return student, err
}

// CreateUser создает нового пользователя и возвращает его id.
func (r *UserRepository) CreateUser(ctx context.Context, user models.User) (int64, error) {
	var id int64

	err := r.db.QueryRow(ctx,
		`INSERT INTO users (book_id, name, surname, middle_name, student_group, birth_date,
		                    password, email, role_level)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
         RETURNING id`,
		user.BookId,
		user.Name,
		user.Surname,
		user.MiddleName,
		user.StudentGroup,
		user.BirthDate,
		user.Password,
		user.Email,
		user.RoleLevel,
	).Scan(&id)
	if err != nil {
		return id, fmt.Errorf("could not create user: %w", err)
	}

	return id, nil
}

// CheckUserWithEmailExists проверяет существование пользователя по email.
func (r *UserRepository) CheckUserWithEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email,
	).Scan(&exists)
	if err != nil {
		return exists, fmt.Errorf("could not check user with email exists: %w", err)
	}

	return exists, err
}

// DeleteUser удаляет пользователя по email/userId и возвращает количество удаленных строк.
func (r *UserRepository) DeleteUser(ctx context.Context, userId int64) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM users WHERE id = $1`,
		userId)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), err

}

// GetUsersWithMaxRole возвращает всех пользователей, у которых роль <= maxRole.
func (r *UserRepository) GetUsersWithMaxRole(ctx context.Context, maxRole int64, limit int) ([]models.User, error) {
	var users []models.User

	err := pgxscan.Select(ctx, r.db, &users,
		`SELECT id,
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
		 WHERE role_level <= $1 limit $2`,
		maxRole, limit)

	return users, err
}

// GetProfileByUserID возвращает профиль пользователя, включая суммарные баллы.
func (r *UserRepository) GetProfileByUserID(ctx context.Context, userID int64) (models.ProfileResponse, error) {
	var profile models.ProfileResponse

	err := pgxscan.Get(ctx, r.db, &profile,
		`SELECT COALESCE(book_id, 0) as book_id,
		        name,
		        surname,
		        middle_name,
		        COALESCE(birth_date, TO_DATE('1970-01-01','YYYY-MM-DD')) as birth_date,
		        COALESCE(student_group, '') as student_group,
		        email,
		        role_level
		 FROM users
		 WHERE id = $1`,
		userID,
	)

	if err != nil {
		return models.ProfileResponse{}, err
	}

	points, err := r.GetUserPoints(ctx, userID)
	if err != nil {
		return models.ProfileResponse{}, err
	}

	profile.TotalPoints = points

	return profile, nil
}

// GetUserRoleAndBookId возвращает роль и номер книжки пользователя.
func (r *UserRepository) GetUserRoleAndBookId(ctx context.Context, userID int64) (int64, int64, error) {
	var role int64
	var bookId int64

	err := r.db.QueryRow(ctx,
		`SELECT role_level, COALESCE(book_id, 0)
         FROM users
         WHERE id = $1`,
		userID,
	).Scan(&role, &bookId)

	return role, bookId, err
}

// RoleExists проверяет, существует ли роль с указанным уровнем.
func (r *UserRepository) RoleExists(ctx context.Context, roleLevel int64) (bool, error) {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM roles WHERE level = $1)`,
		roleLevel,
	).Scan(&exists)

	return exists, err
}

// UpdateUser обновляет данные пользователя согласно UpdateUserRequest.
// Возвращает количество измененных строк.
func (r *UserRepository) UpdateUser(ctx context.Context, req models.UpdateUserRequest) (int64, error) {
	builder := sq.Update("users")

	if req.NewData.BookId != 0 {
		builder = builder.Set("book_id", req.NewData.BookId)
	}
	if req.NewData.Name != "" {
		builder = builder.Set("name", req.NewData.Name)
	}
	if req.NewData.Surname != "" {
		builder = builder.Set("surname", req.NewData.Surname)
	}
	if req.NewData.MiddleName != "" {
		builder = builder.Set("middle_name", req.NewData.MiddleName)
	}
	if req.NewData.StudentGroup != "" {
		builder = builder.Set("student_group", req.NewData.StudentGroup)
	}
	if req.NewData.Email != "" {
		builder = builder.Set("email", req.NewData.Email)
	}
	if req.NewData.RoleLevel != 0 {
		builder = builder.Set("role_level", req.NewData.RoleLevel)
	}

	builder = builder.Where(sq.Eq{"id": req.UserId})

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("could not convert to sql query: %w", err)
	}

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("could not update user: %w", err)
	}

	return tag.RowsAffected(), nil
}

// GetUserPoints возвращает количество баллов пользователя.
func (r *UserRepository) GetUserPoints(ctx context.Context, userID int64) (int64, error) {
	var points int64

	err := r.db.QueryRow(ctx,
		`SELECT total_points
         FROM user_points
         WHERE user_id = $1`,
		userID,
	).Scan(&points)
	if err != nil {
		return 0, fmt.Errorf("could not get user points: %w", err)
	}

	return points, err
}

// GetLeaderboard возвращает список пользователей с максимальными баллами.
func (r *UserRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.UserWithPoints, error) {
	var users []models.UserWithPoints

	query :=
		`SELECT u.id as user_id,
		        u.name,
		        u.surname,
		        up.total_points,
		        ROW_NUMBER() OVER (ORDER BY up.total_points DESC) as position
         FROM user_points up
         JOIN users u ON up.user_id = u.id
         ORDER BY up.total_points DESC
         LIMIT $1`

	err := pgxscan.Select(ctx, r.db, &users, query, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get leaderboard: %w", err)
	}

	return users, err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userId int64, hash []byte) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password = $1 WHERE id = $2`,
		hash, userId,
	)
	return fmt.Errorf("could not update password: %w", err)
}

func (r *UserRepository) GetSuggests(ctx context.Context) ([]models.Event, error) {
	var suggests []models.Event

	err := pgxscan.Select(ctx, r.db, &suggests, `WITH deleted AS (
    DELETE FROM suggest_events
        WHERE expires_at < NOW()
        RETURNING event_id
	)
	SELECT e.*
	FROM events e
	WHERE e.id IN (SELECT event_id FROM suggest_events WHERE event_id NOT IN (SELECT event_id FROM deleted));`)
	if err != nil {
		return suggests, fmt.Errorf("could not get suggests: %w", err)
	}

	return suggests, nil
}
