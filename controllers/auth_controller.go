package controllers

import (
	"bobri/helpers"
	"bobri/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

func AuthCheck(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var AuthCheck models.Auth
		var CurStudent models.Student

		if err := c.ShouldBindJSON(&AuthCheck); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "Error while BindJSON",
				Message: err.Error(),
			})
			return
		}

		ctxStud, cancelStud := context.WithTimeout(c.Request.Context(), 3*time.Second)

		defer cancelStud()

		err := pool.QueryRow(ctxStud, `SELECT id, book_id, surname, name, middle_name, birth_date, "student_group"
			   FROM students WHERE book_id = $1`,
			AuthCheck.Book_id,
		).Scan(
			&CurStudent.Id,
			&CurStudent.Book_id,
			&CurStudent.Surname,
			&CurStudent.Name,
			&CurStudent.Middle_name,
			&CurStudent.Birth_date,
			&CurStudent.Group,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) { // Если зачетки нет
				c.JSON(http.StatusNotFound, gin.H{"ok": "false"})
				return
			}
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{ // Ошибка при поиске зачетки
				Error:   "Error while checking student book_id",
				Message: err.Error()})
			return
		}

		ctxUser, cancelUser := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancelUser()
		var exists bool
		err = pool.QueryRow(ctxUser,
			`SELECT EXISTS(SELECT 1 FROM users WHERE book_id = $1)`,
			AuthCheck.Book_id,
		).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Error while checking if user exists",
				Message: err.Error(),
			})
			return
		}
		if exists {
			// Уже зарегистрирован
			c.JSON(http.StatusConflict, gin.H{"ok": "false"})
			return
		}
		rawToken, err := helpers.GenerateTokenRaw(32) // 256 бит энтропии
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Error while generating token",
				Message: err.Error(),
			})
			return
		}
		tokenHash := helpers.HashToken(rawToken)
		expiresAt := time.Now().Add(helpers.LinkTokenTTL)
		tx, err := pool.BeginTx(ctxUser, pgx.TxOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Error while creating transaction",
				Message: err.Error(),
			})
			return
		}
		defer func() { _ = tx.Rollback(ctxUser) }()

		_, err = tx.Exec(ctxUser, `
			INSERT INTO link_tokens (book_id, token_hash, expires_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (book_id)
			DO UPDATE SET token_hash = EXCLUDED.token_hash,
			              expires_at = EXCLUDED.expires_at,
			              created_at = now()
		`, AuthCheck.Book_id, tokenHash, expiresAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Error while updating token information",
				Message: err.Error(),
			})
			return
		}
		if err := tx.Commit(ctxUser); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Error while commiting transaction",
				Message: err.Error(),
			})
			return
		}
		// 4) Отдаём "free" + сам token и TTL (в секундах)
		c.JSON(http.StatusOK, models.AuthStatus{
			Status:             "free",
			Display_name:       CurStudent.Name,
			Group:              CurStudent.Group,
			Link_token:         rawToken,                            // сырой токен
			Link_token_ttl_sec: int(helpers.LinkTokenTTL.Seconds()), // 300
		})
	}
}
