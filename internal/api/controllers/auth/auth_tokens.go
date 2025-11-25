package auth

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// RefreshToken Обновление access и refresh токенов
// @Summary      Обновление access и refresh токенов
// @Description  Принимает refresh токен, проверяет его валидность и возвращает новую пару токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.RefreshTokenRequest  true  "Refresh токен"
// @Success      200  {object}  models.RefreshTokenResponse  "Успешное обновление токенов"
// @Failure      400  {object}  models.ErrorResponse          "Некорректный JSON"
// @Failure      401  {object}  models.ErrorResponse          "Невалидный или просроченный refresh токен"
// @Failure      500  {object}  models.ErrorResponse          "Ошибка при обновлении/генерации токенов"
// @Router       /auth/refresh [post]
func RefreshToken(service *services.RefreshTokensService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем refresh токен из тела запроса
		var body models.RefreshTokenRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось получить refresh token из тела запроса",
				})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := service.RefreshToken(ctx, body.RefreshToken)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrNoTokensFound):
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   services.ErrNoTokensFound.Error(),
					Message: "Refresh token не найден или просрочен",
				})
			case errors.Is(err, services.ErrNoRowsAffected):
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   services.ErrNoRowsAffected.Error(),
					Message: "Не удалось обновить refresh token",
				})
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при обновлении токенов",
				})
			}
			return

		}

		c.JSON(200, resp)
	}
}
