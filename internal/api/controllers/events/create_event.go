package events

import (
	"bobri/internal/models"
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// CreateEvent  Создание нового события
// @Summary      Создать событие
// @Description  Создаёт новое событие в системе и возвращает полную информацию о созданной записи. Требует авторизации администратора. Поля помимо title могут не указываться, тогда они будут заменены на стандартные значения
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        input          body     models.CreateEventRequest  true  "Данные нового события"
// @Success      200  {object}  models.CreateEventResponse  "Событие успешно создано"
// @Failure      400  {object}  models.ErrorResponse        "Некорректный JSON или ошибка валидации"
// @Failure      401  {object}  models.ErrorResponse        "Нет доступа — невалидный или отсутствующий токен"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка сервера при создании события"
// @Router       /admin/create_event [post]
func CreateEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем модель из тела запроса
		var eventData models.CreateEventRequest
		if err := c.ShouldBindJSON(&eventData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		var builder sq.InsertBuilder
		// Построение динамического SQL запроса с помощью Squirrel
		{
			builder = sq.Insert("users")
			columns := []string{"title"}
			values := []interface{}{eventData.Title}
			fmt.Println(eventData)
			if eventData.Description != "" {
				columns = append(columns, "description")
				values = append(values, eventData.Description)
			}
			if eventData.Points != 0 {
				columns = append(columns, "points")
				values = append(values, eventData.Points)
			}
			if eventData.IconUrl != "" {
				columns = append(columns, "icon_url")
				values = append(values, eventData.IconUrl)
			}
			if eventData.EventDate != nil && !((*eventData.EventDate).IsZero()) {
				columns = append(columns, "event_date")
				values = append(values, eventData.EventDate)
			}
			if eventData.EventTypeCode != 0 {
				columns = append(columns, "event_type_code")
				values = append(values, eventData.EventTypeCode)
			}
			builder = builder.Into("events")
			builder = builder.Columns(columns...)
			builder = builder.Values(values...)
			builder = builder.Suffix("returning id")
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Генерируем SQL запрос и аргументы
		sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()

		// Выполняем запрос и получаем ID созданного события
		var eventId int64
		err = pool.QueryRow(ctx, sqlQuery, args...).Scan(&eventId)
		var pgErr *pgconn.PgError

		// Обработка ошибок Postgres
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				c.JSON(409, models.ErrorResponse{
					Error:   "already_completed",
					Message: "Событие с таким названием уже существует",
				})
				return
			}

			c.JSON(500, models.ErrorResponse{
				Error:   pgErr.Message,
				Message: "Ошибка создания события",
			})
			return

		}

		// Получаем полную информацию о созданном событии
		createEventResponse := models.CreateEventResponse{}
		err = pgxscan.Get(ctx, pool, &createEventResponse,
			"SELECT id, title, description, event_type_code, points, icon_url, event_date, created_at FROM events WHERE id = $1", eventId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении данных созданного события",
			})
			return
		}

		// Возвращаем успешный ответ с данными созданного события
		c.JSON(200, createEventResponse)

		return

	}
}
