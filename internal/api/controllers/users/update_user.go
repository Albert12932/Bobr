package users

import (
	"bobri/internal/models"
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

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

		// Извлекаем данные из тела запроса
		var updateData models.UpdateUserRequest
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		// Получаем уровень роли из payload
		var adminRoleLevel int64
		{

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
			adminRoleLevel = payload.RoleLevel
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Проверяем, если обновляется уровень роли, что такая роль существует
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

		// Получаем текущий уровень роли пользователя, которого обновляем
		var curRoleLevel, curBookId int64
		err := pool.QueryRow(ctx, "SELECT role_level, book_id FROM users WHERE id = $1", updateData.UserId).Scan(&curRoleLevel, &curBookId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении текущего уровня роли пользователя",
			})
			return
		}

		// Проверяем права на изменение: нельзя изменить пользователя с уровнем роли выше или равным своему
		if updateData.NewData.RoleLevel >= adminRoleLevel || curRoleLevel >= adminRoleLevel {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Не достаточно прав",
			})
			return
		}

		// Строим динамический SQL запрос на обновление
		var builder sq.UpdateBuilder
		{
			builder = sq.Update("users")

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
		}

		// Генерируем SQL запрос и аргументы
		sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении данных пользователя",
			})
			return
		}

		// Выполняем SQL запрос
		tag, err := pool.Exec(ctx, sqlQuery, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении данных пользователя",
			})
			return
		}
		// Проверяем, что была обновлена ровно одна строка
		if tag.RowsAffected() != 1 {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении данных пользователя (RowsAffected != 1)",
			})
			return
		}

		// Формируем ответ
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
