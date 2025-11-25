package events

import (
	"bobri/internal/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// GetEvents  Получение списка событий
// @Summary      Получить все события
// @Description  Возвращает полный список событий из базы данных.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Success      200  {array}   models.Event         "Список событий"
// @Failure      500  {object}  models.ErrorResponse "Ошибка сервера при получении списка событий"
// @Router       /admin/events [get]
func GetEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var events []models.Event

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос к базе данных и сканируем результаты в слайс events
		err := pgxscan.Select(ctx, pool, &events, "select id, title, description, event_type_code, points, icon_url, event_date, created_at from events")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении событий",
			})
			return
		}

		// Возвращаем список событий в формате JSON
		c.JSON(200, events)
		return
	}
}
