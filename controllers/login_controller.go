package controllers

import (
	"bobri/helpers"
	"bobri/models"
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
					Error:   "Error while marshaling JSON",
					Message: err.Error(),
				})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var (
			id       int64
			book_id  int
			name     string
			surname  string
			password string
		)
		// Получаем инфу о студенте
		err := pool.QueryRow(ctx, `
		SELECT id, book_id, name, surname, password
		FROM users
		WHERE book_id = $1
		`, loginData.Book_id).Scan(&id, &book_id, &name, &surname, &password)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   "Error while getting user info",
					Message: "Wrong book_id",
				})
				return
			}
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Error while getting user info",
				Message: err.Error(),
			})
			return
		}
		// Сверяем пароль из тела запроса и из бд
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(loginData.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Error while comparing passwordhash",
				Message: err.Error(),
			})
			return
		}
		// Создаем токен
		accessToken, exp, err := jwtMaker.Issue(id, book_id, name, surname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "JWT_ERROR",
				"message": err.Error(),
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
