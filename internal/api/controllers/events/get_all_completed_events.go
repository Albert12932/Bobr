package events

import (
	"bobri/internal/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

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
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос на получение всех выполненных событий
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
		// Возвращаем список всех выполненных событий
		c.JSON(200, completedEvents)

		return
	}
}
