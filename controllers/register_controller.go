package controllers

import (
	"bobri/helpers"
	"bobri/models" // если не нужен – можно убрать
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RegisterByToken(pool *pgxpool.Pool, jwtMaker *helpers.JWTMaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken := c.Param("token")
		if rawToken == "" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Error while getting token from path", Message: "Token not found"})
			return
		}

		var body models.RegisterReq
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Error while marshaling JSON", Message: err.Error()})
			return
		}

		// можно добавить свою проверку сложности пароля здесь
		if len(body.Password) < 8 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Error while creating password", Message: "Weak password"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while creating transaction", Message: err.Error()})
			return
		}
		defer func() { _ = tx.Rollback(ctx) }()

		// 1) Валидируем токен: ищем книгу по ХЭШУ токена и сроку
		var bookID int
		err = tx.QueryRow(ctx, `DELETE FROM link_tokens
       WHERE token_hash = $1 AND expires_at > now()
       RETURNING book_id`,
			helpers.HashToken(rawToken)).Scan(&bookID)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error: "INVALID_OR_EXPIRED_TOKEN", Message: "Token не найден или истёк",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "DATABASE_ERROR", Message: err.Error()})
			return
		}
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE book_id = $1)`, bookID).Scan(&exists); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while checking if user exists", Message: err.Error()})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, models.ErrorResponse{Error: "Error while creating user", Message: "user exists"})
			return
		}
		var (
			firstName, lastName, middleName, studentGroup string
			birth_date                                    time.Time
		)
		// ПОДГОНИ ПОД СВОИ ИМЕНА КОЛОНОК!
		err = tx.QueryRow(ctx, `
			SELECT name, surname, middle_name, student_group, birth_date
			FROM students
			WHERE book_id = $1
		`, bookID).Scan(&firstName, &lastName, &middleName, &studentGroup, &birth_date)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "No rows while getting student info", Message: "No rows"})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while getting student info", Message: err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost) // cost=10 по умолчанию
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while hashing password", Message: err.Error()})
			return
		}

		// 5) Создаём пользователя
		var userID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO users (book_id, name, surname, middle_name, student_group, birth_date, password)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`, bookID, firstName, lastName, middleName, studentGroup, birth_date, (hash)).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while inserting user info", Message: err.Error()})
			return
		}

		// 7) Коммит
		if err := tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while commiting transaction", Message: err.Error()})
			return
		}

		// 8) Выдаём JWT
		accessToken, exp, err := jwtMaker.Issue(userID, bookID, firstName, lastName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error while jwt", Message: err.Error()})
			return
		}

		// 9) Ответ в нужном формате
		var resp models.RegisterResp
		resp.OK = true
		resp.User.ID = userID
		resp.User.FirstName = firstName
		resp.User.LastName = lastName
		resp.Session.Auth.Token = accessToken
		resp.Session.Auth.ExpiresAt = exp

		// на всякий случай отключим кеш
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, resp)
	}
}
