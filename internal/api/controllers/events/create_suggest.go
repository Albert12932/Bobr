package events

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateSuggest  Создание рекомендации для события
// @Summary      Создать рекомендацию для события
// @Description  Создаёт рекомендацию для события с указанием времени истечения в часах с момента создания. Возвращает ID события. Требует авторизации пользователя.
//
//	Поля помимо EventId могут быть не указаны, и они будут обработаны по умолчанию.
//
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}" default(Bearer )
// @Param        input          body     models.CreateSuggestRequest  true  "Данные для создания рекомендации"
// @Success      200  {string}  string  "ID события, для которого была создана рекомендация"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка сервера при создании рекомендации"
// @Router       /admin/create_suggest [post]
func CreateSuggest(service *services.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var body models.CreateSuggestRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный JSON",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := service.SuggestEvent(ctx, body.EventId, body.ExpiresAtHours)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка сервера при создании рекомендации",
			})
			return
		}

		c.JSON(200, models.SuccessResponse{Successful: true, Message: fmt.Sprintf("Event_id: %d", body.EventId)})
	}
}
