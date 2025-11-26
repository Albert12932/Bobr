package auth

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"time"
)

// ResetPassword Запрос на сброс пароля
// @Summary      Запрос на сброс пароля
// @Description  Отправляет на почту пользователя ссылку для сброса пароля.
// @Description  Если пользователь с указанной почтой существует — ему придёт письмо с временной ссылкой на установку нового пароля.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.ResetPasswordRequest  true  "Почта пользователя"
// @Success      200  {object}  map[string]string  "Инструкция отправлена на почту"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка при поиске пользователя или отправке письма"
// @Router       /auth/reset_password [post]
func ResetPassword(service *services.ResetPasswordService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body models.ResetPasswordRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный JSON",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := service.ResetPassword(ctx, body.Email)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				c.JSON(404, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Пользователь с такой почтой не найден",
				})
			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при сбросе пароля",
				})
			}
			return
		}

		c.JSON(200, gin.H{
			"ok":      true,
			"message": "Инструкция по сбросу пароля отправлена",
			"email":   body.Email,
		})
	}
}

// SetNewPassword Установка нового пароля
// @Summary      Установка нового пароля
// @Description  Устанавливает новый пароль по токену из письма. Токен действителен ограниченное время (15 минут).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.SetNewPasswordRequest  true  "Токен и новый пароль"
// @Success      200  {object}  models.SuccessResponse  "Пароль успешно обновлён"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON"
// @Failure      401  {object}  models.ErrorResponse  "Невалидный или истекший токен"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка при обновлении пароля"
// @Router       /auth/set_new_password [post]
func SetNewPassword(service *services.ResetPasswordService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var body models.SetNewPasswordRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат тела запроса",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := service.SetNewPassword(ctx, body.Token, body.NewPassword)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrInvalidResetToken):
				c.JSON(401, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Невалидный или истёкший токен",
				})
				return

			case errors.Is(err, services.ErrWeakPassword):
				c.JSON(400, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Пароль должен быть не менее 8 символов",
				})
				return

			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при обновлении пароля",
				})
				return
			}
		}

		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Пароль успешно обновлён",
		})
	}
}
