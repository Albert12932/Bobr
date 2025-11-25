package auth

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// Login Авторизация пользователя
// @Summary      Авторизация пользователя
// @Description  Проверяет почту и пароль пользователя.
// @Description  При успешной авторизации выдает пару access и refresh токенов.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.LoginRequest  true  "Данные для входа (почта и пароль)"
// @Success      200  {object}  models.LoginResponse     "Успешная авторизация — возвращается пользователь и пара токенов" example({"user":{"id":1,"first_name":"Иван","surname":"Иванов"},"session":{"auth":{"access_token":"token string", "refresh_token":"token string","expires_at":"2024-01-01T00:00:00Z"}}})
// @Failure      400  {object}  models.ErrorResponse     "Некорректный запрос — ошибка парсинга JSON"
// @Failure      401  {object}  models.ErrorResponse     "Неверная почта или пароль"
// @Failure      404  {object}  models.ErrorResponse     "Пользователь с такой почтой не найден"
// @Failure      500  {object}  models.ErrorResponse     "Ошибка сервера при работе с базой данных или токенами"
// @Router       /auth/login [post]
func Login(pool *pgxpool.Pool, AccessJwtMaker *helpers.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем модель (почту, пароль) из тела запроса
		var loginData models.LoginRequest
		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось получить почту и пароль из тела запроса",
				})
			return
		}

		// Создаем контекст с 5-секундным таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Получаем инфу о пользователе
		var user models.User

		err := pgxscan.Get(ctx, pool, &user, `
		SELECT id, coalesce(book_id, 0) as book_id, name, surname, password, email, role_level
		FROM users
		WHERE email = $1
		`, loginData.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "Wrong email",
					Message: "Не удалось найти пользователя с такой почтой",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while getting user info",
			})
			return
		}

		// Сверяем пароль из тела запроса и из бд
		if err := bcrypt.CompareHashAndPassword(user.Password, []byte(loginData.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Неправильный пароль",
			})
			return
		}

		accessToken, exp, refreshToken, errResponse := GetPairOfTokens(pool, AccessJwtMaker,
			models.GetTokensRequest{UserId: user.Id, RoleLevel: user.RoleLevel})
		if errResponse != (models.ErrorResponse{}) {
			c.JSON(http.StatusInternalServerError,
				errResponse)
			return
		}

		// Возвращаем токен + ответ
		var resp models.LoginResponse
		resp.UserSubstructure.ID = user.Id
		resp.UserSubstructure.Email = user.Email
		resp.UserSubstructure.BookId = user.BookId
		resp.UserSubstructure.FirstName = user.Name
		resp.UserSubstructure.RoleLevel = user.RoleLevel
		resp.UserSubstructure.Group = user.Group
		resp.Auth.AccessToken = accessToken
		resp.Auth.RefreshToken = refreshToken
		resp.Auth.ExpUnix = exp

		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, resp)
	}
}
