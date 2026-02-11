package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetEvents  Получение списка событий
// @Summary      Получить все события
// @Description  Возвращает полный список событий из базы данных.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer токен авторизации. Формат: Bearer {token}" default(Bearer )
// @Param        limit   query    int    false   "Максимальное количество мероприятий в выдаче"  default(50)
// @Success      200  {array}   models.Event         "Список событий"
// @Failure      500  {object}  models.ErrorResponse "Ошибка сервера при получении списка событий"
// @Router       /admin/events [get]
func GetEvents(eventService *services.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {

		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		events, err := eventService.GetEvents(ctx, limit)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении событий",
			})
			return
		}

		c.JSON(200, events)
	}
}
