package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
func GetUsers(usersService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {

		payloadInterface, ok := c.Get("userPayload")
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "payload missing",
				Message: "Не удалось получить payload",
			})
			return
		}
		payload := payloadInterface.(*models.Payload)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		users, err := usersService.GetUsers(ctx, payload.RoleLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении пользователей",
			})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}
