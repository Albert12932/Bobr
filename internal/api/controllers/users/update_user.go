package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateUser обновляет данные пользователя
// @Summary Обновление данных пользователя
// @Description Обновляет данные пользователя с проверкой прав доступа. Только пользователи с более высоким уровнем прав могут изменять данные пользователей с более низким уровнем прав.
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен в формате: Bearer {token} default(Bearer )"
// @Param request body models.UpdateUserRequest true "Данные для обновления пользователя"
// @Success 200 {object} models.UpdateUserResponse "Успешное обновление данных пользователя"
// @Failure 400 {object} models.ErrorResponse "Ошибка в формате JSON"
// @Failure 403 {object} models.ErrorResponse "Недостаточно прав"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/update_user [patch]
func UpdateUser(service *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req models.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный JSON",
			})
			return
		}

		// Достаём payload
		payloadRaw, ok := c.Get("userPayload")
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "payload missing",
				Message: "Ошибка извлечения данных из токена",
			})
			return
		}
		payload := payloadRaw.(*models.Payload)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := service.UpdateUser(ctx, payload.RoleLevel, req) // TODO redactor admin
		if err != nil {
			if err.Error() == "недостаточно прав" {
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Недостаточно прав",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении пользователя",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
