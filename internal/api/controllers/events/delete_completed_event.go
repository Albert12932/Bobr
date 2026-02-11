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

// DeleteCompletedEvent  Удалить выполнение события пользователем
// @Summary      Удаление выполнения события
// @Description  Удаляет отметку о выполнении события конкретным пользователем. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer токен авторизации. Формат: Bearer {token}" default(Bearer )
// @Param        user_id        path    int     true  "ID пользователя"
// @Param        event_id       path    int     true  "ID события"
// @Success      200  {object}  models.DeleteCompletedEventResponse        "Отметка о выполнении успешно удалена"
// @Failure      400  {object}  models.ErrorResponse           "Некорректный user_id или event_id"
// @Failure      401  {object}  models.ErrorResponse           "Нет прав доступа"
// @Failure      404  {object}  models.ErrorResponse           "Отметка о выполнении не найдена"
// @Failure      500  {object}  models.ErrorResponse           "Ошибка сервера при удалении"
// @Router       /admin/delete_completed_event/{user_id}/{event_id} [delete]
func DeleteCompletedEvent(service *services.CompletedEventsService) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIdStr := c.Param("user_id")
		eventIdStr := c.Param("event_id")

		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат user_id",
			})
			return
		}

		eventId, err := strconv.ParseInt(eventIdStr, 10, 64)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат event_id",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err = service.DeleteCompletedEvent(ctx, userId, eventId)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrCompletedEventNotFound):
				c.JSON(404, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Выполненное событие не найдено",
				})
				return

			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при удалении выполненного события",
				})
				return
			}
		}

		c.JSON(200, models.DeleteCompletedEventResponse{
			Successful: true,
			UserID:     userId,
			EventID:    eventId,
		})
	}
}
