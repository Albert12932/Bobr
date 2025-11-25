package users

import (
	"bobri/internal/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

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
// @Router       /me/completed_events [get]
func GetCompletedEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userId int64
		// Получаем userId из payload контекста
		{
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
			userId = payload.Sub
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос на получение выполненных событий пользователя
		var completedUserEvents []models.UserCompletedEvent
		err := pgxscan.Select(ctx, pool, &completedUserEvents,
			`SELECT event_id, completed_at from completed_events where user_id = $1`, userId)

		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении выполненных событий",
			})
			return
		}

		// Возвращаем список выполненных событий
		c.JSON(200, completedUserEvents)

		return
	}
}
