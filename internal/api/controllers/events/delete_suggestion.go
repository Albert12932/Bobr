package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DeleteSuggestion  Удаление рекомендации по ID
// @Summary      Удалить рекомендацию по ID
// @Description  Удаляет рекомендацию по заданному ID. Если рекомендация не найдена, возвращает ошибку 404. В случае других ошибок возвращается ошибка 500.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}" default(Bearer )
// @Param        id             path     int64   true  "ID рекомендации для удаления"
// @Success      200  {object}  models.DeleteEventResponse  "Рекомендация успешно удалена"
// @Failure      400  {object}  models.ErrorResponse        "Некорректный формат ID рекомендации"
// @Failure      404  {object}  models.ErrorResponse        "Рекомендация с таким ID не найдено"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка при удалении рекомендации"
// @Router       /admin/delete_suggestion/{id} [delete]
func DeleteSuggestion(service *services.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr := c.Param("id")
		eventId, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат ID рекомендации",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err = service.DeleteSuggestion(ctx, eventId)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoRowsAffected):
				c.JSON(404, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Рекомендация с таким ID не найдено",
				})
			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при удалении рекомендации",
				})
			}
			return
		}

		c.JSON(200, models.DeleteEventResponse{
			Successful: true,
			EventID:    eventId,
		})
	}
}
