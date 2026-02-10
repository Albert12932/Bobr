package repositories

import (
	"bobri/internal/models"
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgconn"
)

// EventRepository Репозиторий использует гибкий интерфейс DBTX.
// Это позволяет одинаково работать с пулом и транзакцией.
type EventRepository struct {
	db DBTX
}

// NewEventRepository инициализирует репозиторий событий с переданным подключением (пулом или транзакцией).
func NewEventRepository(db DBTX) *EventRepository {
	return &EventRepository{db: db}
}

// WithDB Нужен для передачи транзакции в репозиторий
func (r *EventRepository) WithDB(db DBTX) *EventRepository {
	return &EventRepository{db: db}
}

// CreateEvent Создать событие
func (r *EventRepository) CreateEvent(ctx context.Context, data models.CreateEventRequest) (int64, error) {
	builder := sq.Insert("events")

	columns := []string{"title"}
	values := []interface{}{data.Title}

	if data.Description != "" {
		columns = append(columns, "description")
		values = append(values, data.Description)
	}
	if data.Points != 0 {
		columns = append(columns, "points")
		values = append(values, data.Points)
	}
	if data.IconUrl != "" {
		columns = append(columns, "icon_url")
		values = append(values, data.IconUrl)
	}
	if data.EventDate != nil && !data.EventDate.IsZero() {
		columns = append(columns, "event_date")
		values = append(values, data.EventDate)
	}
	if data.EventTypeCode != 0 {
		columns = append(columns, "event_type_code")
		values = append(values, data.EventTypeCode)
	}

	builder = builder.Columns(columns...).Values(values...).Suffix("RETURNING id")

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, err
	}

	var id int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)

	return id, err
}

// GetEventById Получить событие по id
func (r *EventRepository) GetEventById(ctx context.Context, id int64) (models.CreateEventResponse, error) {
	var result models.CreateEventResponse

	err := pgxscan.Get(ctx, r.db, &result,
		`SELECT id, title, description, event_type_code, points,
		        icon_url, event_date, created_at
         FROM events WHERE id = $1`,
		id,
	)

	return result, err
}

// DeleteEvent Удалить событие
func (r *EventRepository) DeleteEvent(ctx context.Context, eventId int64) (pgconn.CommandTag, error) {
	return r.db.Exec(ctx,
		`DELETE FROM events WHERE id = $1`,
		eventId,
	)
}

// GetEvents Получить список всех событий
func (r *EventRepository) GetEvents(ctx context.Context) ([]models.Event, error) {
	var events []models.Event

	err := pgxscan.Select(ctx, r.db, &events,
		`SELECT id, title, description, event_type_code, points,
		        icon_url, event_date, created_at
		 FROM events`,
	)

	return events, err
}

// UpdateEvent Обновить данные события
func (r *EventRepository) UpdateEvent(ctx context.Context, req models.UpdateEventRequest) error {
	builder := sq.Update("events")

	if req.NewData.Title != "" {
		builder = builder.Set("title", req.NewData.Title)
	}
	if req.NewData.Description != "" {
		builder = builder.Set("description", req.NewData.Description)
	}
	if req.NewData.Points != 0 {
		builder = builder.Set("points", req.NewData.Points)
	}
	if req.NewData.IconUrl != "" {
		builder = builder.Set("icon_url", req.NewData.IconUrl)
	}
	if req.NewData.EventDate != nil && !req.NewData.EventDate.IsZero() {
		builder = builder.Set("event_date", req.NewData.EventDate)
	}
	if req.NewData.EventTypeCode != 0 {
		builder = builder.Set("event_type_code", req.NewData.EventTypeCode)
	}

	builder = builder.Where(sq.Eq{"id": req.EventId})

	query, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *EventRepository) CreateSuggest(ctx context.Context, eventId int64, expiresAt time.Time) (pgconn.CommandTag, error) {

	tag, err := r.db.Exec(ctx, `INSERT INTO suggest_events (event_id, expires_at) values ($1, $2)`, eventId, expiresAt)
	return tag, err
}

func (r *EventRepository) DeleteSuggest(ctx context.Context, eventId int64) (pgconn.CommandTag, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM suggest_events WHERE event_id = $1`, eventId)
	if err != nil {
		return tag, err
	}
	return tag, err
}
