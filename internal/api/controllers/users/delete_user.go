package users

import (
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
		var userEmail string
		// Получаем email из параметров пути
		{
			userEmail = c.Param("email")

			if userEmail == "" {
				c.JSON(400, models.ErrorResponse{
					Error:   "empty email",
					Message: "Передан пустой email",
				})
				return
			}
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Удаляем пользователя из базы данных
		tag, err := pool.Exec(ctx, "DELETE FROM users WHERE email = $1", userEmail)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while deleting user",
			})
			return
		}
		// Проверяем, был ли удален пользователь
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "User not found",
				Message: "Не удалось найти пользователя с такой почтой",
			})
			return
		}

		// Возвращаем успешный ответ
		c.JSON(200, models.DeleteUserResponse{
			Successful: true,
			Email:      userEmail,
		})
		return
	}

}
