package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
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
func DeleteUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		email := c.Param("email")
		if email == "" {
			c.JSON(400, models.ErrorResponse{
				Error:   "empty email",
				Message: "Передан пустой email",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := userService.DeleteUser(ctx, email)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				c.JSON(404, models.ErrorResponse{
					Error:   "User not found",
					Message: "Пользователь с такой почтой не найден",
				})
				return
			default:
				c.JSON(500, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при удалении пользователя",
				})
				return
			}
		}

		c.JSON(200, models.DeleteUserResponse{
			Successful: true,
			Email:      email,
		})
	}
}
