package events

import (
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

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
		// Берем модель из тела запроса
		var completeData models.CompleteUserEventRequest

		if err := c.ShouldBindJSON(&completeData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем вставку записи о выполнении события
		_, err := pool.Exec(ctx,
			`INSERT INTO completed_events (user_id, event_id)
			 VALUES ($1, $2)`, completeData.UserId, completeData.EventId)

		var pgErr *pgconn.PgError
		// Обработка ошибок Postgres
		{
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
		}

		// Возвращаем успешный ответ
		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Событие успешно отмечено как выполненное",
		})

		return
	}
}
