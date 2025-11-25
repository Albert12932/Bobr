package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"time"
)

// AddCompletedEvent  Отметить событие как выполненное пользователем
// @Summary      Отметить выполнение события
// @Description  Добавляет запись о выполнении события конкретным пользователем. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        input          body    models.CompleteUserEventRequest  true   "ID пользователя и ID события"
// @Success      200  {object}  models.SuccessResponse                   "Событие отмечено как выполненное"
// @Failure      400  {object}  models.ErrorResponse                     "Некорректные данные или пользователь/событие не существуют"
// @Failure      401  {object}  models.ErrorResponse                     "Нет прав доступа"
// @Failure      409  {object}  models.ErrorResponse                     "Событие уже было отмечено ранее"
// @Failure      500  {object}  models.ErrorResponse                     "Ошибка сервера при добавлении записи"
// @Router       /admin/add_completed_event [post]
func AddCompletedEvent(service *services.CompletedEventsService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var body models.CompleteUserEventRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный JSON",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := service.AddCompletedEvent(ctx, body.UserId, body.EventId)
		if err != nil {

			switch {
			// TODO SQLSTATE 23505
			case errors.Is(err, services.ErrInvalidReference):
				c.JSON(400, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Пользователь или событие не существуют",
				})
				return

			case errors.Is(err, services.ErrAlreadyCompleted):
				c.JSON(409, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Это событие уже отмечено пользователем",
				})
				return

			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка базы данных",
				})
				return
			}
		}

		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Событие успешно отмечено как выполненное",
		})
	}
}
