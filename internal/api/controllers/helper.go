package controllers

import (
	"bobri/internal/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

// DeleteUser @Summary      Удаление пользователя
// @Description  Удаляет пользователя по номеру студенческого билета.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer <token>)
// @Param        input  body  models.AuthBookRequest  true  "Номер студенческого"
// @Success      200  {object}  models.DeleteUserResponse  "Пользователь успешно удалён" example({"deleted":true,"book_id":123456})
// @Failure      400  {object}  models.ErrorResponse        "Некорректный JSON"
// @Failure      404  {object}  models.ErrorResponse        "Пользователь не найден"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка при удалении пользователя"
// @Router       /helper/delete_user [delete]
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
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer <token>)
// @Success      200  {array}  models.Student             "Список студентов" example([{"id":1,"book_id":123456,"surname":"Иванов","name":"Иван","middle_name":"Иванович","birth_date":"2000-01-01T00:00:00Z","student_group":"ШАД-111"}, {"id":2,"book_id":654321,"surname":"Петров","name":"Пётр","middle_name":"Петрович","birth_date":"1999-12-31T00:00:00Z","student_group":"ШАД-111"}])
// @Failure      500  {object} models.ErrorResponse        "Ошибка при запросе или чтении данных"
// @Router       /helper/students [get]
func GetStudents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var students []models.Student

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err := pgxscan.Select(ctx, pool, &students,
			"select id, book_id, surname, name, middle_name, birth_date, student_group from students")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении студентов",
			})
			return
		}

		c.JSON(200, students)
		return
	}

}

// GetUsers @Summary      Получение списка пользователей
// @Description  Возвращает всех зарегистрированных пользователей из таблицы users.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer <token>)
// @Success      200  {array}  models.User               "Список пользователей" example([{"id":1,"book_id":123456,"surname":"Иванов","name":"Иван","middle_name":"Иванович","birth_date":"2000-01-01T00:00:00Z","student_group":"ШАД-111","password":"hashed_password","mail":"string mail"}, {"id":2,"book_id":654321,"surname":"Петров","name":"Пётр","middle_name":"Петрович","birth_date":"1999-12-31T00:00:00Z","student_group":"ШАД-111","password":"hashed_password","mail":"string mail"}])
// @Failure      500  {object} models.ErrorResponse       "Ошибка при запросе или чтении данных"
// @Router       /helper/users [get]
func GetUsers(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()
		// TODO where level_role < payload.roleLevel
		err := pgxscan.Select(ctx, pool, &users,
			"SELECT id, COALESCE(book_id, 0) as book_id, surname, name, middle_name, coalesce(birth_date, '1970-01-01'::timestamp) as birth_date, coalesce(student_group, '') as student_group, password, mail, role_level FROM users")

		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при сканировании пользователей",
			})
		}

		c.JSON(200, users)
		return
	}

}

// SetRole @Summary      Установка уровня роли пользователя
// @Description  Устанавливает уровень роли для указанного пользователя. Требуются права администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer <token>)
// @Param        request  body  models.SetRoleRequest  true  "Данные для установки роли"
// @Success      200  {object}  models.SetRoleResponse  "Роль успешно установлена"
// @Failure      400  {object}  models.ErrorResponse    "Некорректный JSON"
// @Failure      403  {object}  models.ErrorResponse    "Недостаточно прав"
// @Failure      500  {object}  models.ErrorResponse    "Ошибка сервера"
// @Router       /helper/set_role [post]
func SetRole(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		var setRoleRequest models.SetRoleRequest
		if err := c.ShouldBindJSON(&setRoleRequest); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		payloadInterface, existsPayload := c.Get("userPayload")
		if !existsPayload {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   "Payload doesn't exist",
					Message: "Данные о пользователе в JWT Токене (Payload) не найдены",
				})
			return
		}

		// Преобразуем payload в наш тип
		payload := payloadInterface.(*models.Payload)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var roleLevel int64

		err := pool.QueryRow(ctx, "SELECT role_level FROM users  where id = $1", payload.Sub).Scan(&roleLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось получить роль пользователя",
			})
			return
		}

		var exists bool
		err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM roles WHERE level = $1)", roleLevel).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Не удалось проверить наличие роли",
			})
			return
		}

		if roleLevel <= setRoleRequest.RoleLevel {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Не достаточно прав",
			})
			return
		}

		Tag, err := pool.Exec(ctx, "UPDATE users set role_level = $1 where id = $2", setRoleRequest.RoleLevel, setRoleRequest.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при установки роли пользователя",
			})
			return
		}
		if Tag.RowsAffected() != 1 {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при установки роли пользователя (RowsAffected != 1)",
			})
			return
		}

		c.JSON(200, models.SetRoleResponse{
			Successful: true,
			UserID:     setRoleRequest.UserId,
			RoleLevel:  setRoleRequest.RoleLevel,
		})
		return
	}
}
