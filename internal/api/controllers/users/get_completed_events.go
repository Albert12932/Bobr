package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

// GetCompletedEvents Получение выполненных событий пользователя
// @Summary      Получить выполненные события пользователя
// @Description  Возвращает список выполненных пользователем событий, а также статистику по категориям:
// @Description  - Хакатоны (type = 1)
// @Description  - Статьи (type = 2)
// @Description  - Олимпиады (type = 3)
// @Description  - Проекты (type = 4)
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header   string  true  "Bearer токен в формате: Bearer {token}"
// @Success      200  {object}  models.CompletedEventsFullResponse
// @Failure      500  {object}  models.ErrorResponse  "Ошибка при получении данных"
// @Router       /me/completed_events [get]
func GetCompletedEvents(service *services.CompletedEventsService) gin.HandlerFunc {
	return func(c *gin.Context) {

		payload := c.MustGet("userPayload").(*models.Payload)
		userId := payload.Sub

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := service.GetCompletedEvents(ctx, userId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении выполненных событий",
			})
			return
		}

		c.JSON(200, resp)
	}
}
