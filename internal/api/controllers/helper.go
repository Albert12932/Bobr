package controllers

import (
	"bobri/internal/models"
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

// DeleteUser  Удаление пользователя по email
// @Summary      Удалить пользователя
// @Description  Удаляет пользователя по его email. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        email          path    string  true   "Email удаляемого пользователя"
// @Success      200  {object}  models.DeleteUserResponse    "Пользователь успешно удалён"
// @Failure      400  {object}  models.ErrorResponse "Передан пустой email"
// @Failure      401  {object}  models.ErrorResponse "Нет прав доступа к операции"
// @Failure      404  {object}  models.ErrorResponse "Пользователь с таким email не найден"
// @Failure      500  {object}  models.ErrorResponse "Ошибка сервера при удалении"
// @Router       /admin/delete_user/{email} [delete]
func DeleteUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userEmail := c.Param("email")

		if userEmail == "" {
			c.JSON(400, models.ErrorResponse{
				Error:   "empty email",
				Message: "Передан пустой email",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		tag, err := pool.Exec(ctx, "DELETE FROM users WHERE email = $1", userEmail)
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
				Message: "Не удалось найти пользователя с такой почтой",
			})
			return
		}
		c.JSON(200, models.DeleteUserResponse{
			Successful: true,
			Email:      userEmail,
		})
		return
	}

}

// GetStudents Получение списка студентов
// @Summary      Получение списка студентов
// @Description  Возвращает всех студентов из таблицы students.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer )
// @Success      200  {array}  models.Student             "Список студентов"
// @Failure      500  {object} models.ErrorResponse        "Ошибка при запросе или чтении данных"
// @Router       /admin/students [get]
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

// GetUsers Получение списка пользователей
// @Summary      Получение списка пользователей
// @Description  Возвращает всех зарегистрированных пользователей из таблицы users.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer )
// @Success      200  {array}  models.User               "Список пользователей"
// @Failure      500  {object} models.ErrorResponse       "Ошибка при запросе или чтении данных"
// @Router       /admin/users [get]
func GetUsers(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

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
		err := pgxscan.Select(ctx, pool, &users,
			"SELECT id, COALESCE(book_id, 0) as book_id, surname, name, middle_name, coalesce(birth_date, '1970-01-01'::timestamp) as birth_date, coalesce(student_group, '') as student_group, password, email, role_level FROM users where role_level <= $1", payload.RoleLevel)

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

// UpdateUser обновляет данные пользователя
// @Summary Обновление данных пользователя
// @Description Обновляет данные пользователя с проверкой прав доступа. Только пользователи с более высоким уровнем прав могут изменять данные пользователей с более низким уровнем прав.
// @Tags admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer токен в формате: Bearer {token} default(Bearer )"
// @Param request body models.UpdateUserRequest true "Данные для обновления пользователя"
// @Success 200 {object} models.UpdateUserResponse "Успешное обновление данных пользователя"
// @Failure 400 {object} models.ErrorResponse "Ошибка в формате JSON"
// @Failure 403 {object} models.ErrorResponse "Недостаточно прав"
// @Failure 500 {object} models.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/update_user [patch]
func UpdateUser(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO "can't scan into dest[1]: cannot scan NULL into *int64", обновление админа
		var updateData models.UpdateUserRequest
		if err := c.ShouldBindJSON(&updateData); err != nil {
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
		adminRoleLevel := payload.RoleLevel

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if updateData.NewData.RoleLevel != 0 {
			var exists bool
			err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM roles WHERE level = $1)", updateData.NewData.RoleLevel).Scan(&exists)
			if err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не удалось проверить наличие роли",
				})
				return
			}
		}

		var curRoleLevel, curBookId int64

		err := pool.QueryRow(ctx, "SELECT role_level, book_id FROM users WHERE id = $1", updateData.UserId).Scan(&curRoleLevel, &curBookId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении текущего уровня роли пользователя",
			})
			return
		}

		if updateData.NewData.RoleLevel >= adminRoleLevel || curRoleLevel >= adminRoleLevel {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Не достаточно прав",
			})
			return
		}

		builder := sq.Update("users")

		if updateData.NewData.BookId != 0 {
			builder = builder.Set("book_id", updateData.NewData.BookId)
		}
		if updateData.NewData.Name != "" {
			builder = builder.Set("name", updateData.NewData.Name)
		}
		if updateData.NewData.Surname != "" {
			builder = builder.Set("surname", updateData.NewData.Surname)
		}
		if updateData.NewData.MiddleName != "" {
			builder = builder.Set("middle_name", updateData.NewData.MiddleName)
		}
		if updateData.NewData.StudentGroup != "" {
			builder = builder.Set("student_group", updateData.NewData.StudentGroup)
		}
		if updateData.NewData.Email != "" {
			builder = builder.Set("email", updateData.NewData.Email)
		}
		if updateData.NewData.RoleLevel != 0 {
			builder = builder.Set("role_level", updateData.NewData.RoleLevel)
		}

		builder = builder.Where(sq.Eq{"id": updateData.UserId})

		sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении данных пользователя",
			})
			return
		}

		tag, err := pool.Exec(ctx, sqlQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении данных пользователя",
			})
			return
		}
		if tag.RowsAffected() != 1 {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении данных пользователя (RowsAffected != 1)",
			})
			return
		}

		var resp models.UpdateUserResponse

		resp.UserID = updateData.UserId
		resp.Successful = true
		resp.New.BookId = updateData.NewData.BookId
		resp.New.Name = updateData.NewData.Name
		resp.New.Surname = updateData.NewData.Surname
		resp.New.MiddleName = updateData.NewData.MiddleName
		resp.New.StudentGroup = updateData.NewData.StudentGroup
		resp.New.Email = updateData.NewData.Email
		resp.New.RoleLevel = updateData.NewData.RoleLevel

		c.JSON(200, resp)

		return

	}
}

// GetEvents  Получение списка событий
// @Summary      Получить все события
// @Description  Возвращает полный список событий из базы данных.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Success      200  {array}   models.Event         "Список событий"
// @Failure      500  {object}  models.ErrorResponse "Ошибка сервера при получении списка событий"
// @Router       /admin/events [get]
func GetEvents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var events []models.Event

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		err := pgxscan.Select(ctx, pool, &events, "select id, title, description, event_type_code, points, icon_url, event_date, created_at from events")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении событий",
			})
			return
		}

		c.JSON(200, events)
		return
	}
}
