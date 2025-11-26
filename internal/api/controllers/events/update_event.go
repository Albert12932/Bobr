package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

// UpdateEvent  Частичное обновление события
// @Summary      Обновить выбранные поля события
// @Description  Производит частичное обновление данных события по его ID.
// @Description  Обновляются только те поля, которые переданы в теле запроса.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        input          body     models.UpdateEventRequest  true  "ID события и новые значения полей"
// @Success      200  {object}  models.SuccessResponse  "Событие успешно обновлено"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON или ошибка валидации"
// @Failure      401  {object}  models.ErrorResponse  "Неавторизованный доступ — неверный или отсутствующий токен"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка сервера при попытке обновления записи"
// @Router       /admin/update_event [patch]
func UpdateEvent(eventService *services.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var updateData models.UpdateEventRequest
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат JSON",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := eventService.UpdateEvent(ctx, updateData)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении события",
			})
			return
		}

		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Событие успешно обновлено",
		})
	}
}
