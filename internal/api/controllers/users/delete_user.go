package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DeleteUser  Удаление пользователя по user_id
// @Summary      Удалить пользователя
// @Description  Удаляет пользователя по его user_id. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        email          path    string  true   "UserId удаляемого пользователя"
// @Success      200  {object}  models.DeleteUserResponse    "Пользователь успешно удалён"
// @Failure      400  {object}  models.ErrorResponse "Передан пустой id пользователя"
// @Failure      401  {object}  models.ErrorResponse "Нет прав доступа к операции"
// @Failure      404  {object}  models.ErrorResponse "Пользователь с таким user_id не найден"
// @Failure      500  {object}  models.ErrorResponse "Ошибка сервера при удалении"
// @Router       /admin/delete_user/{user_id} [delete]
func DeleteUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIdStr := c.Param("user_id")
		if userIdStr == "" {
			c.JSON(400, models.ErrorResponse{
				Error:   "empty argument",
				Message: "Передан пустой user_id",
			})
			return
		}
		userId, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   "invalid argument",
				Message: "Передан некорректный user_id",
			})
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err = userService.DeleteUser(ctx, userId)
		if err != nil {
			switch {
			case errors.Is(err, services.ErrUserNotFound):
				c.JSON(404, models.ErrorResponse{
					Error:   "User not found",
					Message: "Пользователь с такой почтой/user_id не найден",
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
			UserId:     userId,
		})
	}
}
