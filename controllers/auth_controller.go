package controllers

import (
	"bobri/models"
	"context"
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
		var exists bool

		if err := c.ShouldBindJSON(&AuthCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid book_id",
				"message": err.Error(),
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)

		defer cancel()

		err := pool.QueryRow(ctx, "select * from students where book_id = $1", AuthCheck.Book_id).Scan(&CurStudent.Id, &CurStudent.Book_id, &CurStudent.Surname, &CurStudent.Name, &CurStudent.Middle_name, &CurStudent.Birth_date, &CurStudent.Group)
		if err != nil {
			if err == pgx.ErrNoRows { /* Если проверили зачетку и ее нет в списке студентов */
				c.JSON(http.StatusConflict, gin.H{
					"ok": "false",
				})
			return
			}

			c.JSON(http.StatusInternalServerError, gin.H{ /* Ошибка при запросе */
				"error":   "DATABASE_ERROR",
				"message": err.Error(),
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)

		err := pool.QueryRow(ctx, "select * from users where book_id = $1", AuthCheck.Book_id).Scan(&exists)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusOK, AuthStatus {
					Status: true,
					Display_name: CurStudent.Name,
					group: CurStudent.Group
					Link_token: "", /* TODO */
					Link_token_ttl_sec 300,
				})
				exists = false
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{ /* Ошибка при запросе */
					"error":   "DATABASE_ERROR",
					"message": err.Error(),
				})
				return
			}
		} else {
			c.JSON(http.StatusConflict, gin.H{
				"ok": "false",
			})
			exists = true
		}
	}
}
