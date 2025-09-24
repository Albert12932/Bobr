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
		var CurUser models.User

		if err := c.ShouldBindJSON(&AuthCheck); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid book_id",
				"message": err.Error(),
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)

		defer cancel()

		err := pool.QueryRow(ctx, "select * from users where book_id = $1", AuthCheck.Book_id).Scan(&CurUser.Id, &CurUser.Book_id, &CurUser.Surname, &CurUser.Name, &CurUser.Middle_name, &CurUser.Birth_date, &CurUser.Group)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusOK, models.AuthStatus{
					Status:             "free",
					Display_name:       CurUser.Name,
					Group:              CurUser.Group,
					Link_token:         "Null",
					Link_token_ttl_sec: 0, /* TODO */
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "DATABASE_ERROR",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{
			"ok": "false",
		})
	}
}
