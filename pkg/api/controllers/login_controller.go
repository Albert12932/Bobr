package controllers

import (
	"bobri/pkg/helpers"
	"bobri/pkg/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

func Login(pool *pgxpool.Pool, jwtMaker *helpers.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем модель (студенческий, пароль) из тела запроса
		var loginData models.LoginRequest
		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось получить студенческий и пароль тела запроса",
				})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var (
			id       int64
			bookId   int
			name     string
			surname  string
			password []byte
		)
		// Получаем инфу о пользователе
		err := pool.QueryRow(ctx, `
		SELECT id, book_id, name, surname, password
		FROM users
		WHERE book_id = $1
		`, loginData.BookId).Scan(&id, &bookId, &name, &surname, &password)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   "Wrong book_id",
					Message: "Не удалось найти пользователя с таким студенческим",
				})
				return
			}
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
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
		// Создаем токен
		accessToken, exp, err := jwtMaker.Issue(id, bookId, name, surname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "JWT_ERROR",
			})
			return
		}
		// Возвращаем токен + ответ
		var resp models.AuthResp
		resp.OK = true
		resp.User.ID = id
		resp.User.FirstName = name
		resp.User.Surname = surname
		resp.Session.Auth.Token = accessToken
		resp.Session.Auth.ExpiresAt = exp

		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, resp)
	}
}
