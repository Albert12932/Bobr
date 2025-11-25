package auth

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
		resp.Auth.ExpUnix = exp

		c.JSON(200, resp)
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

func GetPairOfTokens(pool *pgxpool.Pool, AccessJwtMaker *helpers.JWTMaker, userData models.GetTokensRequest) (AccessToken string, AccessTokenExp int64, RefreshToken string, errResp models.ErrorResponse) {

	// Создаем контекст с таймаутом 5 сек
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Создаем access токен
	accessToken, exp, err := AccessJwtMaker.Issue(userData.UserId, userData.RoleLevel)
	if err != nil {
		return "", 0, "",
			models.ErrorResponse{
				Error:   err.Error(),
				Message: "JWT_ERROR"}
	}

	// Удаляем прошлые refresh токены пользователя TODO fingerprints
	Tag, err := pool.Exec(ctx, `
			DELETE FROM refresh_tokens
			WHERE user_id = $1
		`, userData.UserId)
	if err != nil {
		return "", 0, "",
			models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось удалить старые refresh token"}
	}

	// Создаем refresh token
	refreshToken, err := helpers.NewRefreshToken()
	if err != nil {
		return "", 0, "", models.ErrorResponse{
			Error:   err.Error(),
			Message: "Не удалось создать refresh token",
		}
	}

	// Сохраняем новый refresh token в бд
	Tag, err = pool.Exec(ctx, `
			INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
			VALUES ($1, $2, $3)
		`, userData.UserId, helpers.HashToken(refreshToken), time.Now().Add(30*24*time.Hour)) // 30 дней`)
	if err != nil {
		return "", 0, "",
			models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось сохранить refresh token"}
	}
	if Tag.RowsAffected() != 1 {
		return "", 0, "",
			models.ErrorResponse{
				Error:   "RowsAffected != 1",
				Message: "Не удалось сохранить refresh token"}
	}

	return accessToken, exp, refreshToken, models.ErrorResponse{}
}
