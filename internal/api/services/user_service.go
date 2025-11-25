package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) DeleteUser(ctx context.Context, email string) error {
	rows, err := s.userRepo.DeleteUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}
func (s *UserService) GetUsers(ctx context.Context, maxRole int64) ([]models.User, error) {
	return s.userRepo.GetUsersWithMaxRole(ctx, maxRole)
}
func (s *UserService) GetProfile(ctx context.Context, userID int64) (models.ProfileResponse, error) {
	return s.userRepo.GetProfileByUserID(ctx, userID)
}
func (s *UserService) UpdateUser(ctx context.Context, adminRole int64, req models.UpdateUserRequest) (models.UpdateUserResponse, error) {

	// Проверяем роль
	if req.NewData.RoleLevel != 0 {
		exists, err := s.userRepo.RoleExists(ctx, req.NewData.RoleLevel)
		if err != nil {
			return models.UpdateUserResponse{}, err
		}
		if !exists {
			return models.UpdateUserResponse{}, errors.New("такой роли не существует")
		}
	}

	// Получаем текущую роль и book_id
	curRole, _, err := s.userRepo.GetUserRoleAndBookId(ctx, req.UserId)
	if err != nil {
		return models.UpdateUserResponse{}, err
	}

	// Проверяем полномочия
	if req.NewData.RoleLevel >= adminRole || curRole >= adminRole {
		return models.UpdateUserResponse{}, errors.New("недостаточно прав")
	}

	// Строим SQL через Squirrel
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

	sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return models.UpdateUserResponse{}, err
	}

	rows, err := s.userRepo.UpdateUser(ctx, sqlQuery, args)
	if err != nil {
		return models.UpdateUserResponse{}, err
	}
	if rows != 1 {
		return models.UpdateUserResponse{}, errors.New("RowsAffected != 1")
	}

	var resp models.UpdateUserResponse
	resp.Successful = true
	resp.UserID = req.UserId
	resp.New = req.NewData

	return resp, nil
}

// GetLeaderboard возвращает топ пользователей по количеству очков.
// limit – сколько записей вернуть (если некорректный, подрежем до разумных границ).
func (s *UserService) GetLeaderboard(ctx context.Context, limit int) ([]models.UserWithPoints, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	return s.userRepo.GetLeaderboard(ctx, limit)
}
