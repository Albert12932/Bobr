package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
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
func GetAllCompletedEvents(service *services.CompletedEventsService) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		events, err := service.GetAllCompletedEvents(ctx)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении выполненных событий",
			})
			return
		}

		c.JSON(200, events)
	}
}
