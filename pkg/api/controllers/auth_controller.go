package controllers

import (
	"bobri/pkg/helpers"
	"bobri/pkg/models"
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

		// Берем номер студенческого из тела запроса
		if err := c.ShouldBindJSON(&AuthCheck); err != nil {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не получилось получить номер студенческого из тела запроса",
				})
			return
		}

		ctxStud, cancelStud := context.WithTimeout(c.Request.Context(), 3*time.Second)

		defer cancelStud()

		// Получаем все данные студента
		err := pool.QueryRow(ctxStud, `SELECT id, book_id, surname, name, middle_name, birth_date, "student_group"
			   FROM students WHERE book_id = $1`,
			AuthCheck.BookId,
		).Scan(
			&CurStudent.Id,
			&CurStudent.BookId,
			&CurStudent.Surname,
			&CurStudent.Name,
			&CurStudent.MiddleName,
			&CurStudent.BirthDate,
			&CurStudent.Group,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) { // Если студенческого нет
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   pgx.ErrNoRows.Error(),
					Message: "Студенческого с таким номером не найдено",
				})
				return
			}
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while checking student book_id"})
			return
		}

		// Проверяем, есть ли пользователь с таким номером студенческого
		ctxUser, cancelUser := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancelUser()
		var exists bool
		err = pool.QueryRow(ctxUser,
			`SELECT EXISTS(SELECT 1 FROM users WHERE book_id = $1)`,
			AuthCheck.BookId,
		).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while checking if user exists",
			})
			return
		}
		if exists {
			// Уже зарегистрирован
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "User already registered",
				Message: "Пользователь с таким номером студенческого уже зарегистрирован",
			})
			return
		}

		// Генерируем токен для регистрации
		rawToken, err := helpers.GenerateTokenRaw(32) // 256 бит энтропии
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while generating token",
			})
			return
		}

		tokenHash := helpers.HashToken(rawToken)
		expiresAt := time.Now().Add(helpers.LinkTokenTTL)

		// Создаем транзакцию для создания токена и добавления его в бд
		tx, err := pool.BeginTx(ctxUser, pgx.TxOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while creating transaction",
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
		`, AuthCheck.BookId, tokenHash, expiresAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while updating token information",
				})
			return
		}

		// Коммитим транзакцию
		if err := tx.Commit(ctxUser); err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while commiting transaction",
				})
			return
		}
		// 4) Отдаём "free" + сам token и TTL (в секундах)
		c.JSON(http.StatusOK, models.AuthStatus{
			Status:          "free",
			DisplayName:     CurStudent.Name,
			Group:           CurStudent.Group,
			LinkToken:       rawToken,                            // сырой токен
			LinkTokenTtlSec: int(helpers.LinkTokenTTL.Seconds()), // 300
		})
	}
}
