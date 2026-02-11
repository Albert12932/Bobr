package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"
)

var (
	ErrInvalidReference       = errors.New("пользователь или событие не существуют")
	ErrAlreadyCompleted       = errors.New("событие уже отмечено пользователем")
	ErrCompletedEventNotFound = errors.New("выполненное событие не найдено")
)

type CompletedEventsService struct {
	repo *repositories.CompletedEventsRepository
	uow  *repositories.UoW
}

// NewCompletedEventsService создает сервис выполненных событий.
func NewCompletedEventsService(repo *repositories.CompletedEventsRepository, uow *repositories.UoW) *CompletedEventsService {
	return &CompletedEventsService{
		repo: repo,
		uow:  uow,
	}
}

// AddCompletedEvent добавляет выполненное событие пользователю.
func (s *CompletedEventsService) AddCompletedEvent(ctx context.Context, userId, eventId int64) error {
	return s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		// вызываем репозиторий через WithDB(tx)

		err := s.repo.WithDB(tx).AddCompletedEvent(ctx, userId, eventId)
		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteCompletedEvent удаляет отметку о выполнении события.
func (s *CompletedEventsService) DeleteCompletedEvent(ctx context.Context, userId, eventId int64) error {
	tag, err := s.repo.DeleteCompletedEvent(ctx, userId, eventId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrCompletedEventNotFound
	}

	return nil
}

// GetAllCompletedEvents возвращает список всех выполненных событий.
func (s *CompletedEventsService) GetAllCompletedEvents(ctx context.Context, limit int) ([]models.CompletedEvent, error) {
	return s.repo.GetAllCompletedEvents(ctx, limit)
}

// GetCompletedEvents возвращает выполненные события пользователя с агрегированной статистикой.
func (s *CompletedEventsService) GetCompletedEvents(ctx context.Context, userId int64) (models.CompletedEventsFullResponse, error) {
	return s.repo.GetCompletedEventsWithStats(ctx, userId)
}
