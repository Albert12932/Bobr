package controllers

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
	"net/http"
	"strconv"
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

		var eventData models.CreateEventRequest
		if err := c.ShouldBindJSON(&eventData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		builder := sq.Insert("users")
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()

		var eventId int64
		fmt.Println(sqlQuery, args)

		err = pool.QueryRow(ctx, sqlQuery, args...).Scan(&eventId)
		var pgErr *pgconn.PgError

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

		createEventResponse := models.CreateEventResponse{}

		err = pgxscan.Get(ctx, pool, &createEventResponse, "SELECT id, title, description, event_type_code, points, icon_url, event_date, created_at FROM events WHERE id = $1", eventId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении данных созданного события",
			})
			return
		}
		c.JSON(200, createEventResponse)

		return

	}
}

// UpdateEvent  Частичное обновление события
// @Summary      Обновить выбранные поля события
// @Description  Производит частичное обновление данных события по его ID.
// @Description  Обновляются только те поля, которые переданы в теле запроса.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        input          body     models.UpdateEventRequest  true  "ID события и новые значения полей"
// @Success      200  {object}  models.SuccessResponse  "Событие успешно обновлено"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON или ошибка валидации"
// @Failure      401  {object}  models.ErrorResponse  "Неавторизованный доступ — неверный или отсутствующий токен"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка сервера при попытке обновления записи"
// @Router       /admin/update_event [patch]
func UpdateEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		var updateData models.UpdateEventRequest
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		builder := sq.Update("events")

		if updateData.NewData.Title != "" {
			builder = builder.Set("title", updateData.NewData.Title)
		}
		if updateData.NewData.Description != "" {
			builder = builder.Set("description", updateData.NewData.Description)
		}
		if updateData.NewData.Points != 0 {
			builder = builder.Set("points", updateData.NewData.Points)
		}
		if updateData.NewData.IconUrl != "" {
			builder = builder.Set("icon_url", updateData.NewData.IconUrl)
		}
		if updateData.NewData.EventDate != nil && !((*updateData.NewData.EventDate).IsZero()) {
			builder = builder.Set("event_date", updateData.NewData.EventDate)
		}
		if updateData.NewData.EventTypeCode != 0 {
			builder = builder.Set("event_type_code", updateData.NewData.EventTypeCode)
		}

		builder = builder.Where(sq.Eq{"id": updateData.EventId})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()

		fmt.Println(sqlQuery, args)

		_, err = pool.Exec(ctx, sqlQuery, args...)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении события",
			})
			return
		}

		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Событие успешно обновлено",
		})

		return

	}
}

// DeleteEvent  Удаление события по ID
// @Summary      Удалить событие
// @Description  Удаляет событие по его идентификатору. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        id             path    int     true   "ID события для удаления"
// @Success      200  {object}  models.DeleteEventResponse "Событие успешно удалено"
// @Failure      400  {object}  models.ErrorResponse    "Некорректный ID события"
// @Failure      401  {object}  models.ErrorResponse    "Нет прав доступа"
// @Failure      404  {object}  models.ErrorResponse    "Событие с указанным ID не найдено"
// @Failure      500  {object}  models.ErrorResponse    "Ошибка сервера при удалении"
// @Router       /admin/delete_event/{id} [delete]
func DeleteEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		paramEventId := c.Param("id")
		if paramEventId == "" {
			c.JSON(400, models.ErrorResponse{
				Error:   "Event ID is not found",
				Message: "ID события не был передан в параметрах запроса",
			})
			return
		}
		deleteEventId, err := strconv.Atoi(paramEventId)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат ID события",
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tag, err := pool.Exec(ctx, "DELETE FROM events WHERE id = $1", deleteEventId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при удалении события",
			})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "Event not found",
				Message: "Событие не найдено",
			})
			return
		}

		c.JSON(200, models.DeleteEventResponse{
			Successful: true,
			EventID:    deleteEventId,
		})

		return

	}
}

// AddCompletedEvent  Отметить событие как выполненное пользователем
// @Summary      Отметить выполнение события
// @Description  Добавляет запись о выполнении события конкретным пользователем. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        input          body    models.CompleteUserEventRequest  true   "ID пользователя и ID события"
// @Success      200  {object}  models.SuccessResponse                   "Событие отмечено как выполненное"
// @Failure      400  {object}  models.ErrorResponse                     "Некорректные данные или пользователь/событие не существуют"
// @Failure      401  {object}  models.ErrorResponse                     "Нет прав доступа"
// @Failure      409  {object}  models.ErrorResponse                     "Событие уже было отмечено ранее"
// @Failure      500  {object}  models.ErrorResponse                     "Ошибка сервера при добавлении записи"
// @Router       /admin/add_completed_event [post]
func AddCompletedEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var completeData models.CompleteUserEventRequest
		if err := c.ShouldBindJSON(&completeData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := pool.Exec(ctx,
			`INSERT INTO completed_events (user_id, event_id)
			 VALUES ($1, $2)`, completeData.UserId, completeData.EventId)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			// Ошибка внешнего ключа
			if pgErr.Code == "23503" {
				c.JSON(400, models.ErrorResponse{
					Error:   "invalid_reference",
					Message: "Пользователь или событие не существует",
				})
				return
			}

			// Дубликат (PK violation, unique violation)
			if pgErr.Code == "23505" {
				c.JSON(409, models.ErrorResponse{
					Error:   "already_completed",
					Message: "Это событие уже отмечено пользователем",
				})
				return
			}

			// Любая другая ошибка Postgres
			c.JSON(500, models.ErrorResponse{
				Error:   pgErr.Message,
				Message: "Ошибка базы данных",
			})
			return
		}
		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Событие успешно отмечено как выполненное",
		})

		return
	}
}

// DeleteCompletedEvent  Удалить выполнение события пользователем
// @Summary      Удаление выполнения события
// @Description  Удаляет отметку о выполнении события конкретным пользователем. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        user_id        path    int     true  "ID пользователя"
// @Param        event_id       path    int     true  "ID события"
// @Success      200  {object}  models.DeleteCompletedEventResponse        "Отметка о выполнении успешно удалена"
// @Failure      400  {object}  models.ErrorResponse           "Некорректный user_id или event_id"
// @Failure      401  {object}  models.ErrorResponse           "Нет прав доступа"
// @Failure      404  {object}  models.ErrorResponse           "Отметка о выполнении не найдена"
// @Failure      500  {object}  models.ErrorResponse           "Ошибка сервера при удалении"
// @Router       /admin/delete_completed_event/{user_id}/{event_id} [delete]
func DeleteCompletedEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIdParam := c.Param("user_id")
		eventIdParam := c.Param("event_id")
		if userIdParam == "" || eventIdParam == "" {
			c.JSON(400, models.ErrorResponse{
				Error:   "user_id or event_id is missing",
				Message: "user_id или event_id отсутствует в параметрах запроса",
			})
			return
		}
		userId, err := strconv.Atoi(userIdParam)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат user_id",
			})
			return
		}

		eventId, err := strconv.Atoi(eventIdParam)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат event_id",
			})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tag, err := pool.Exec(ctx,
			`DELETE from completed_events where user_id = $1 and event_id = $2`, userId, eventId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при удалении выполненного события",
			})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "Event not found RowsAffected=0",
				Message: "Выполненное событие не найдено",
			})
			return
		}

		c.JSON(200, models.DeleteCompletedEventResponse{
			Successful: true,
			UserID:     userId,
			EventID:    eventId,
		})

		return
	}
}

// GetCompletedEvents  Получение выполненных событий пользователя
// @Summary      Получить выполненные события
// @Description  Возвращает список событий, которые пользователь отметил как выполненные.
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Success      200  {array}   models.UserCompletedEvent                  "Список выполненных пользователем событий"
// @Failure      400  {object}  models.ErrorResponse                   "Некорректный user_id"
// @Failure      401  {object}  models.ErrorResponse                   "Нет доступа"
// @Failure      500  {object}  models.ErrorResponse                   "Ошибка сервера при получении выполненных событий"
// @Router       /completed_events/{user_id} [get]
func GetCompletedEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		payloadInterface, existsPayload := c.Get("userPayload")
		if !existsPayload {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   "Payload doesn't exist",
					Message: "Данные о пользователе в JWT Токене (Payload) не найдены",
				})
			return
		}

		// Преобразуем payload в наш тип
		payload := payloadInterface.(*models.Payload)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var completedUserEvents []models.UserCompletedEvent

		err := pgxscan.Select(ctx, pool, &completedUserEvents,
			`SELECT event_id, completed_at from completed_events where user_id = $1`, payload.Sub)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении выполненных событий",
			})
			return
		}
		c.JSON(200, completedUserEvents)

		return
	}
}

// GetAllCompletedEvents  Получение всех выполненных событий пользователей
// @Summary      Получить все выполненные события
// @Description  Возвращает полный список выполненных событий всех пользователей системы.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Success      200  {array}   models.CompletedEvent  "Список выполненных событий всех пользователей"
// @Failure      401  {object}  models.ErrorResponse        "Нет прав доступа"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка сервера при получении выполненных событий"
// @Router       /admin/completed_events [get]
func GetAllCompletedEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var completedEvents []models.CompletedEvent

		err := pgxscan.Select(ctx, pool, &completedEvents,
			`SELECT user_id, event_id, completed_at from completed_events`)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении выполненных событий",
			})
			return
		}
		c.JSON(200, completedEvents)

		return
	}
}
