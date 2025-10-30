package controllers

import (
	"bobri/pkg/models"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func DeleteUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userData struct {
			BookId int `json:"book_id"`
		}
		if err := c.ShouldBindJSON(&userData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   "Error while marshaling JSON",
				Message: err.Error(),
			})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		tag, err := pool.Exec(ctx, "DELETE FROM users WHERE book_id = $1", userData.BookId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   "Error while deleting user",
				Message: err.Error(),
			})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "User not found",
				Message: "No user with the given book_id",
			})
			return
		}
		c.JSON(200, gin.H{"deleted": true, "book_id": userData.BookId})
		return
	}

}

func GetStudents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.Student

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		rows, err := pool.Query(ctx, "select id, book_id, surname, name, middle_name, birth_date, student_group from students")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   "Error while querying students",
				Message: err.Error(),
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var user models.Student
			err := rows.Scan(&user.Id, &user.BookId, &user.Surname, &user.Name, &user.MiddleName, &user.BirthDate, &user.Group)
			if err != nil {
				c.JSON(500, models.ErrorResponse{
					Error:   "Error while scanning student row",
					Message: err.Error(),
				})
				return
			}
			users = append(users, user)
		}
		c.JSON(200, users)
		return
	}

}
