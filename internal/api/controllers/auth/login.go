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

// Login Авторизация пользователя
// @Summary      Авторизация пользователя
// @Description  Проверяет почту и пароль пользователя. При успешной авторизации выдает пару access и refresh токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.LoginRequest  true  "Данные для входа (почта и пароль)"
// @Success      200  {object}  models.LoginResponse  "Успешная авторизация — возвращается пользователь и пара токенов"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный запрос — ошибка парсинга JSON"
// @Failure      401  {object}  models.ErrorResponse  "Неправильный пароль"
// @Failure      404  {object}  models.ErrorResponse  "Пользователь с такой почтой не найден"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка сервера при работе с базой данных или токенами"
// @Router       /auth/login [post]
func Login(loginService *services.LoginService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем email + password из тела запроса
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат запроса",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Вызываем сервис логина
		resp, err := loginService.Login(ctx, req.Email, req.Password)
		if err != nil {

			switch {
			case errors.Is(err, services.ErrUserNotFound):
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Пользователь не найден",
				})
				return

			case errors.Is(err, services.ErrInvalidPassword):
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Неправильный пароль",
				})
				return

			case errors.Is(err, services.ErrRowsAffected):
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка обновления refresh-токена",
				})
				return

			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Внутренняя ошибка сервера",
				})
				return
			}
		}

		// Успех
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, resp)
	}
}
