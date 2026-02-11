package repositories

import (
	"bobri/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CompletedEventsRepository Репозиторий получает DBTX — это может быть pool или tx
type CompletedEventsRepository struct {
	db DBTX
}

func NewCompletedEventsRepository(db DBTX) *CompletedEventsRepository {
	return &CompletedEventsRepository{db: db}
}

// WithDB позволяет репозиторию работать поверх транзакции
func (r *CompletedEventsRepository) WithDB(db DBTX) *CompletedEventsRepository {
	return &CompletedEventsRepository{db: db}
}

// AddCompletedEvent — добавляет событие + обновляет очки.
func (r *CompletedEventsRepository) AddCompletedEvent(ctx context.Context, userId, eventId int64) error {

	//Получаем очки события
	var points int
	err := r.db.QueryRow(ctx, `SELECT points FROM events WHERE id = $1`, eventId).Scan(&points)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("event points not found for eventId %d: %w", eventId, err)
		}
		return err
	}

	// 2. Добавляем выполненное событие
	tag, err := r.db.Exec(ctx,
		`INSERT INTO completed_events (user_id, event_id) VALUES ($1, $2)`, userId, eventId,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return fmt.Errorf("could not insert completed event: %w", err)
	}
	// 3. Обновляем очки пользователя
	tag, err = r.db.Exec(ctx,
		`INSERT INTO user_points (user_id, total_points) VALUES ($1, $2) ON CONFLICT (user_id)
         DO UPDATE SET total_points = user_points.total_points + EXCLUDED.total_points`, userId, points,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return fmt.Errorf("could not insert update user points: %w", err)
	}

	return nil
}

// DeleteCompletedEvent удаляет связь user_id + event_id
func (r *CompletedEventsRepository) DeleteCompletedEvent(ctx context.Context, userId, eventId int64) (pgconn.CommandTag, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM completed_events 
         WHERE user_id = $1 AND event_id = $2`,
		userId, eventId,
	)
	if err != nil || tag.RowsAffected() == 0 {
		return pgconn.CommandTag{}, fmt.Errorf("could not delete completed event: %w", err)
	}
	return tag, nil
}

// GetAllCompletedEvents Получить все события
func (r *CompletedEventsRepository) GetAllCompletedEvents(ctx context.Context, limit int) ([]models.CompletedEvent, error) {
	var result []models.CompletedEvent

	err := pgxscan.Select(ctx, r.db, &result,
		`SELECT user_id, event_id, completed_at FROM completed_events limit $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get completed events: %w", err)
	}

	return result, err
}

// GetCompletedEventsWithStats Получить события пользователя + статистику
func (r *CompletedEventsRepository) GetCompletedEventsWithStats(ctx context.Context, userId int64) (models.CompletedEventsFullResponse, error) {
	var resp models.CompletedEventsFullResponse

	// 1. Получаем выполненные события
	err := pgxscan.Select(ctx, r.db, &resp.Events,
		`SELECT e.id AS id, e.title, e.description, e.event_type_code, e.points, e.icon_url, e.event_date, e.created_at,
	ce.completed_at
     FROM events e
     JOIN completed_events ce ON e.id = ce.event_id
     WHERE ce.user_id = $1
       AND ce.completed_at IS NOT NULL;`,
		userId,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, fmt.Errorf("completed events not found: %w", err)
		}
		return resp, fmt.Errorf("could not get completed events: %w", err)
	}

	// 2. Получаем статистику по типам
	var rows []struct {
		EventType int `db:"event_type_code"`
		Count     int `db:"count"`
	}

	err = pgxscan.Select(ctx, r.db, &rows,
		`SELECT e.event_type_code, COUNT(*) 
         FROM completed_events ce
         JOIN events e ON ce.event_id = e.id
         WHERE ce.user_id = $1
         GROUP BY e.event_type_code`,
		userId,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return resp, fmt.Errorf("completed events not found: %w", err)
		}
		return resp, fmt.Errorf("could not get completed events: %w", err)
	}

	// распределяем статистику
	for _, row := range rows {
		switch row.EventType {
		case 1:
			resp.Stats.Hackathons = row.Count
		case 2:
			resp.Stats.Articles = row.Count
		case 3:
			resp.Stats.Olympiads = row.Count
		case 4:
			resp.Stats.Projects = row.Count
		}
	}

	return resp, nil
}
