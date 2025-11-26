package services

import (
	"bobri/internal/api/repositories"
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrEventAlreadyExists = errors.New("событие с таким названием уже существует")
	ErrEventNotFound      = errors.New("событие не найдено")
)

type EventService struct {
	eventRepo *repositories.EventRepository
}

func NewEventService(repo *repositories.EventRepository) *EventService {
	return &EventService{eventRepo: repo}
}

func (s *EventService) CreateEvent(ctx context.Context, data models.CreateEventRequest) (models.CreateEventResponse, error) {
	id, err := s.eventRepo.CreateEvent(ctx, data)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.CreateEventResponse{}, ErrEventAlreadyExists
		}

		return models.CreateEventResponse{}, err
	}

	event, err := s.eventRepo.GetEventById(ctx, id)
	if err != nil {
		return models.CreateEventResponse{}, err
	}

	return event, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, eventId int64) error {
	tag, err := s.eventRepo.DeleteEvent(ctx, eventId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}
func (s *EventService) GetEvents(ctx context.Context) ([]models.Event, error) {
	return s.eventRepo.GetEvents(ctx)
}
func (s *EventService) UpdateEvent(ctx context.Context, req models.UpdateEventRequest) error {
	return s.eventRepo.UpdateEvent(ctx, req)
}
