package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSuggests  Получение списка рекомендаций
// @Summary      Получить список рекомендаций для событий
// @Description  Возвращает список рекомендаций для событий. Если произошла ошибка при получении данных, возвращается код ошибки 500.
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Success      200  {array}   models.Event  "Список рекомендаций событий"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка при получении рекомендаций"
// @Router       /get_suggests [get]
func GetSuggests(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		events, err := userService.GetSuggestions(ctx)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении рекомендаций",
			})
			return
		}

		c.JSON(200, events)
	}
}
