package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetProfile Получение профиля пользователя
// @Summary      Получение профиля пользователя
// @Description  Возвращает данные о пользователе
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer )
// @Success      200  {object} models.ProfileResponse               "Данные о пользователе"
// @Failure      500  {object} models.ErrorResponse       "Ошибка при запросе или чтении данных"
// @Router       /me/profile [get]
func GetProfile(profileService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		payloadInterface, ok := c.Get("userPayload")
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "payload missing",
				Message: "Не удалось получить данные пользователя из токена",
			})
			return
		}
		payload := payloadInterface.(*models.Payload)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		profile, err := profileService.GetProfile(ctx, payload.Sub) //TODO points sum
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении профиля пользователя",
			})
			return
		}

		c.JSON(http.StatusOK, profile)
	}
}
