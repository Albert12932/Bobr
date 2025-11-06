package controllers

import (
	"bobri/pkg/models"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// DeleteUser @Summary      Удаление пользователя
// @Description  Удаляет пользователя по номеру студенческого билета.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        input  body  models.AuthBookRequest  true  "Номер студенческого"
// @Success      200  {object}  models.DeleteUserResponse  "Пользователь успешно удалён" example({"deleted":true,"book_id":123456})
// @Failure      400  {object}  models.ErrorResponse        "Некорректный JSON"
// @Failure      404  {object}  models.ErrorResponse        "Пользователь не найден"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка при удалении пользователя"
// @Router       /helper/deleteUser [delete]
func DeleteUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userData models.AuthBookRequest
		if err := c.ShouldBindJSON(&userData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		tag, err := pool.Exec(ctx, "DELETE FROM users WHERE book_id = $1", userData.BookId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while deleting user",
			})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "User not found",
				Message: "Не удалось найти пользователя с таким студенческим",
			})
			return
		}
		c.JSON(200, models.DeleteUserResponse{
			Deleted: true,
			BookId:  userData.BookId,
		})
		return
	}

}

// GetStudents @Summary      Получение списка студентов
// @Description  Возвращает всех студентов из таблицы students.
// @Tags         students
// @Produce      json
// @Success      200  {array}  models.Student             "Список студентов" example([{"id":1,"book_id":123456,"surname":"Иванов","name":"Иван","middle_name":"Иванович","birth_date":"2000-01-01T00:00:00Z","student_group":"ШАД-111"}, {"id":2,"book_id":654321,"surname":"Петров","name":"Пётр","middle_name":"Петрович","birth_date":"1999-12-31T00:00:00Z","student_group":"ШАД-111"}])
// @Failure      500  {object} models.ErrorResponse        "Ошибка при запросе или чтении данных"
// @Router       /helper/students [get]
func GetStudents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var students []models.Student

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		rows, err := pool.Query(ctx, "select id, book_id, surname, name, middle_name, birth_date, student_group from students")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while querying students",
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var student models.Student
			err := rows.Scan(&student.Id, &student.BookId, &student.Surname, &student.Name, &student.MiddleName, &student.BirthDate, &student.Group)
			if err != nil {
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while scanning student row",
				})
				return
			}
			students = append(students, student)
		}
		c.JSON(200, students)
		return
	}

}

// GetUsers @Summary      Получение списка пользователей
// @Description  Возвращает всех зарегистрированных пользователей из таблицы users.
// @Tags         users
// @Produce      json
// @Success      200  {array}  models.User               "Список пользователей" example([{"id":1,"book_id":123456,"surname":"Иванов","name":"Иван","middle_name":"Иванович","birth_date":"2000-01-01T00:00:00Z","student_group":"ШАД-111","password":"hashed_password","mail":"string mail"}, {"id":2,"book_id":654321,"surname":"Петров","name":"Пётр","middle_name":"Петрович","birth_date":"1999-12-31T00:00:00Z","student_group":"ШАД-111","password":"hashed_password","mail":"string mail"}])
// @Failure      500  {object} models.ErrorResponse       "Ошибка при запросе или чтении данных"
// @Router       /helper/users [get]
func GetUsers(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		rows, err := pool.Query(ctx, "select id, book_id, surname, name, middle_name, birth_date, student_group, password, mail from users")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while querying students",
			})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var user models.User
			err := rows.Scan(&user.Id, &user.BookId, &user.Surname, &user.Name, &user.MiddleName, &user.BirthDate, &user.Group, &user.Password, &user.Mail)
			if err != nil {
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while scanning user row",
				})
				return
			}
			users = append(users, user)
		}
		c.JSON(200, users)
		return
	}

}
