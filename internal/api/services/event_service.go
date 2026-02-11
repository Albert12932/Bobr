package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEventAlreadyExists = errors.New("событие с таким названием уже существует")
	ErrEventNotFound      = errors.New("событие не найдено")
)

type EventService struct {
	events *repositories.EventRepository
	uow    *repositories.UoW
}

func NewEventService(repo *repositories.EventRepository, uow *repositories.UoW) *EventService {
	return &EventService{
		events: repo,
		uow:    uow,
	}
}

// CreateEvent создает новое событие.
func (s *EventService) CreateEvent(ctx context.Context, data models.CreateEventRequest) (models.CreateEventResponse, error) {
	var result models.CreateEventResponse

	err := s.uow.WithinTransaction(ctx, func(ctx context.Context, tx repositories.DBTX) error {
		id, err := s.events.WithDB(tx).CreateEvent(ctx, data)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrEventAlreadyExists
			}
			return err
		}

		result, err = s.events.WithDB(tx).GetEventById(ctx, id)
		return err
	})

	return result, err
}

// DeleteEvent удаляет событие по id.
func (s *EventService) DeleteEvent(ctx context.Context, eventId int64) error {
	tag, err := s.events.DeleteEvent(ctx, eventId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

// GetEvents возвращает список всех событий.
func (s *EventService) GetEvents(ctx context.Context, limit int) ([]models.Event, error) {
	return s.events.GetEvents(ctx, limit)
}

// UpdateEvent обновляет событие.
func (s *EventService) UpdateEvent(ctx context.Context, req models.UpdateEventRequest) error {
	return s.events.UpdateEvent(ctx, req)
}

func (s *EventService) SuggestEvent(ctx context.Context, eventId, expiresAtHours int64) error {
	tag, err := s.events.CreateSuggest(ctx, eventId, time.Now().Add(time.Hour*time.Duration(expiresAtHours)))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func (s *EventService) DeleteSuggestion(ctx context.Context, eventId int64) error {
	_, err := s.events.DeleteSuggest(ctx, eventId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoRowsAffected
		}
		return err
	}
	return nil
}
