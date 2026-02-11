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

// DeleteEvent  Удаление события по ID
// @Summary      Удалить событие
// @Description  Удаляет событие по его идентификатору. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}" default(Bearer )
// @Param        id             path    int     true   "ID события для удаления"
// @Success      200  {object}  models.DeleteEventResponse "Событие успешно удалено"
// @Failure      400  {object}  models.ErrorResponse    "Некорректный ID события"
// @Failure      401  {object}  models.ErrorResponse    "Нет прав доступа"
// @Failure      404  {object}  models.ErrorResponse    "Событие с указанным ID не найдено"
// @Failure      500  {object}  models.ErrorResponse    "Ошибка сервера при удалении"
// @Router       /admin/delete_event/{id} [delete]
func DeleteEvent(service *services.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {

		idStr := c.Param("id")
		eventId, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат ID события",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err = service.DeleteEvent(ctx, eventId)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrEventNotFound):
				c.JSON(404, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Событие с таким ID не найдено",
				})
			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при удалении события",
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
