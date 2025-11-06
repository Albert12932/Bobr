package controllers

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// Login @Summary      Авторизация пользователя
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
		var (
			id        int64
			name      string
			surname   string
			password  []byte
			mail      string
			roleLevel int
		)
		err := pool.QueryRow(ctx, `
		SELECT id, name, surname, password, mail, role_level
		FROM users
		WHERE mail = $1
		`, loginData.Mail).Scan(&id, &name, &surname, &password, &mail, &roleLevel)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   "Wrong mail",
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
		if err := bcrypt.CompareHashAndPassword(password, []byte(loginData.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Неправильный пароль",
			})
			return
		}

		// Создаем access токен
		accessToken, exp, err := AccessJwtMaker.Issue(id, int64(roleLevel))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "JWT_ERROR",
			})
			return
		}

		// Удаляем прошлые refresh токены пользователя TODO fingerprints
		Tag, err := pool.Exec(ctx, `
			DELETE FROM refresh_tokens
			WHERE user_id = $1
		`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось удалить старые refresh token"})
			return
		}

		// Создаем refresh token
		refreshToken, err := helpers.NewRefreshToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось создать refresh token",
			})
			return
		}

		// Сохраняем новый refresh token в бд
		Tag, err = pool.Exec(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3)
		`, id, helpers.HashToken(refreshToken), time.Now().Add(30*24*time.Hour)) // 30 дней`)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось сохранить refresh token"})
			return
		}
		if Tag.RowsAffected() != 1 {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   "RowsAffected != 1",
					Message: "Не удалось сохранить refresh token"})
			return
		}

		// Возвращаем токен + ответ
		var resp models.LoginResponse
		resp.UserSubstructure.ID = id
		resp.UserSubstructure.Mail = mail
		resp.UserSubstructure.FirstName = name
		resp.UserSubstructure.RoleLevel = roleLevel // TODO
		resp.Session.Auth.AccessToken = accessToken
		resp.Session.Auth.RefreshToken = refreshToken
		resp.Session.Auth.ExpiresAt = exp

		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, resp)
	}
}

func getRefreshTokenUserId(pool *pgxpool.Pool, refreshToken string) (int64, error) {

	// Контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userId int64

	// Проверяем валидность refresh токена
	err := pool.QueryRow(ctx, "SELECT user_id from refresh_tokens where token_hash = $1 and expires_at > now()", helpers.HashToken(refreshToken)).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

// RefreshToken @Summary      Обновление access и refresh токенов
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
func RefreshToken(pool *pgxpool.Pool, AccessJwtMaker *helpers.JWTMaker) gin.HandlerFunc {
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

		// Проверяем валидность refresh токена и получаем userId
		userId, err := getRefreshTokenUserId(pool, body.RefreshToken)
		if err != nil && userId == 0 {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось найти refreshToken в базе или он истек",
				})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось проверить валидность refresh token",
				})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var roleLevel int64

		err = pool.QueryRow(ctx, "SELECT role_level from users where id = $1", userId).Scan(&roleLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка получения роли при обновлении AccessToken-а",
			})
			return
		}

		// Создаем новый access токен
		accessToken, exp, err := AccessJwtMaker.Issue(userId, roleLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "Error creating access token",
			})
			return
		}

		// Создаем новый refresh токен
		newRefreshToken, err := helpers.NewRefreshToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось создать новый refresh token",
			})
			return
		}

		// Обновляем refresh токен в бд
		Tag, err := pool.Exec(ctx, `
			UPDATE refresh_tokens
			SET token_hash = $1, expires_at = $2
			WHERE user_id = $3
		`, helpers.HashToken(newRefreshToken), time.Now().Add(30*24*time.Hour), userId) // 30 дней
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось обновить refresh token",
				})
			return
		}
		if Tag.RowsAffected() != 1 {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   "RowsAffected != 1",
					Message: "Не удалось обновить refresh token",
				})
			return
		}

		// Возвращаем токены + ответ
		var resp models.RefreshTokenResponse
		resp.UserID = userId
		resp.Auth.AccessToken = accessToken
		resp.Auth.RefreshToken = newRefreshToken
		resp.Auth.ExpiresAt = exp

		c.JSON(200, resp)
	}
}
