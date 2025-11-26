package repositories

import (
	"bobri/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompletedEventsRepository struct {
	db *pgxpool.Pool
}

func NewCompletedEventsRepository(db *pgxpool.Pool) *CompletedEventsRepository {
	return &CompletedEventsRepository{db: db}
}

// AddCompletedEventTx — добавляет запись о выполнении события и обновляет очки
func (r *CompletedEventsRepository) AddCompletedEventTx(ctx context.Context, tx pgx.Tx, userId, eventId int64) error {
	// 1. Получаем количество очков события
	var points int
	err := tx.QueryRow(ctx,
		`SELECT points FROM events WHERE id = $1`,
		eventId,
	).Scan(&points)
	if err != nil {
		return fmt.Errorf("failed to get event points: %w", err)
	}

	// 2. Добавляем запись о выполнении события
	_, err = tx.Exec(ctx,
		`INSERT INTO completed_events (user_id, event_id)
         VALUES ($1, $2)`,
		userId, eventId)
	if err != nil {
		return fmt.Errorf("failed to insert completed event: %w", err)
	}

	// 3. Обновляем (или создаём) сумму очков пользователя
	_, err = tx.Exec(ctx,
		`INSERT INTO user_points (user_id, total_points)
         VALUES ($1, $2)
         ON CONFLICT (user_id)
         DO UPDATE SET total_points = user_points.total_points + EXCLUDED.total_points`,
		userId, points)
	if err != nil {
		return fmt.Errorf("failed to update user_points: %w", err)
	}

	return nil
}

// DeleteCompletedEvent — удаляет связь user_id + event_id
func (r *CompletedEventsRepository) DeleteCompletedEvent(ctx context.Context, userId, eventId int64) (pgconn.CommandTag, error) {
	return r.db.Exec(ctx,
		`DELETE FROM completed_events 
         WHERE user_id = $1 AND event_id = $2`,
		userId, eventId)
}
func (r *CompletedEventsRepository) GetAllCompletedEvents(ctx context.Context) ([]models.CompletedEvent, error) {
	var result []models.CompletedEvent

	err := pgxscan.Select(ctx, r.db, &result,
		`SELECT user_id, event_id, completed_at FROM completed_events`)

	return result, err
}
func (r *CompletedEventsRepository) GetCompletedEventsWithStats(ctx context.Context, userId int64) (models.CompletedEventsFullResponse, error) {
	var resp models.CompletedEventsFullResponse

	// 1. Получаем список выполненных событий
	err := pgxscan.Select(ctx, r.db, &resp.Events,
		`SELECT event_id, completed_at 
         FROM completed_events 
         WHERE user_id = $1`,
		userId,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		resp.Events = []models.UserCompletedEvent{}
	} else if err != nil {
		return resp, err
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
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return resp, err
	}

	// распределяем по структуре Stats
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
