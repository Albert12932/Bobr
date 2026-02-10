package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) DeleteUser(ctx context.Context, userId int64) error {
	rows, err := s.userRepo.DeleteUser(ctx, userId)
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
	// проверяем существование роли
	if req.NewData.RoleLevel != 0 {
		exists, err := s.userRepo.RoleExists(ctx, req.NewData.RoleLevel)
		if err != nil {
			return models.UpdateUserResponse{}, err
		}
		if !exists {
			return models.UpdateUserResponse{}, errors.New("такой роли не существует")
		}
	}

	// получаем текущую роль
	curRole, _, err := s.userRepo.GetUserRoleAndBookId(ctx, req.UserId)
	if err != nil {
		return models.UpdateUserResponse{}, err
	}

	// проверяем полномочия
	if req.NewData.RoleLevel >= adminRole || curRole >= adminRole {
		return models.UpdateUserResponse{}, errors.New("недостаточно прав")
	}

	// выполняем обновление через репозиторий
	rows, err := s.userRepo.UpdateUser(ctx, req)
	if err != nil {
		return models.UpdateUserResponse{}, err
	}

	if rows != 1 {
		return models.UpdateUserResponse{}, errors.New("RowsAffected != 1")
	}

	return models.UpdateUserResponse{
		Successful: true,
		UserID:     req.UserId,
		New:        req.NewData,
	}, nil
}

func (s *UserService) GetLeaderboard(ctx context.Context, limit int) ([]models.UserWithPoints, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	return s.userRepo.GetLeaderboard(ctx, limit)
}

func (s *UserService) GetSuggestions(ctx context.Context) ([]models.Event, error) {
	return s.userRepo.GetSuggests(ctx)
}
