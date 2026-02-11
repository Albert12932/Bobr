package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateEvent  Создание нового события
// @Summary      Создать событие
// @Description  Создаёт новое событие в системе и возвращает полную информацию о созданной записи. Требует авторизации администратора. Поля помимо title могут не указываться, тогда они будут заменены на стандартные значения
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}" default(Bearer )
// @Param        input          body     models.CreateEventRequest  true  "Данные нового события"
// @Success      200  {object}  models.CreateEventResponse  "Событие успешно создано"
// @Failure      400  {object}  models.ErrorResponse        "Некорректный JSON или ошибка валидации"
// @Failure      401  {object}  models.ErrorResponse        "Нет доступа — невалидный или отсутствующий токен"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка сервера при создании события"
// @Router       /admin/create_event [post]
func CreateEvent(service *services.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var body models.CreateEventRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный JSON",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		event, err := service.CreateEvent(ctx, body)
		if err != nil {

			switch {
			case errors.Is(err, services.ErrEventAlreadyExists):
				c.JSON(409, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Событие с таким названием уже существует",
				})
				return
			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка сервера при создании события",
				})
				return
			}
		}

		c.JSON(200, event)
	}
}
