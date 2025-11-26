package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidReference       = errors.New("пользователь или событие не существуют")
	ErrAlreadyCompleted       = errors.New("событие уже отмечено пользователем")
	ErrCompletedEventNotFound = errors.New("выполненное событие не найдено")
)

type CompletedEventsService struct {
	repo *repositories.CompletedEventsRepository
	db   *pgxpool.Pool
}

func NewCompletedEventsService(repo *repositories.CompletedEventsRepository, db *pgxpool.Pool) *CompletedEventsService {
	return &CompletedEventsService{repo: repo, db: db}
}

func (s *CompletedEventsService) AddCompletedEvent(ctx context.Context, userId, eventId int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = s.repo.AddCompletedEventTx(ctx, tx, userId, eventId)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
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
func (s *CompletedEventsService) GetAllCompletedEvents(ctx context.Context) ([]models.CompletedEvent, error) {
	return s.repo.GetAllCompletedEvents(ctx)
}
func (s *CompletedEventsService) GetCompletedEvents(ctx context.Context, userId int64) (models.CompletedEventsFullResponse, error) {
	return s.repo.GetCompletedEventsWithStats(ctx, userId)
}
