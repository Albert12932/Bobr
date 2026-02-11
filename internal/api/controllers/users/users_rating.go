package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetLeaderboard Получить топ пользователей по количеству очков
// @Summary      Лидерборд пользователей
// @Description  Возвращает список пользователей, отсортированный по количеству набранных очков.
// @Tags		 user
// @Produce      json
// @Security     BearerAuth
// @Param        limit   query    int    false   "Максимальное количество пользователей в выдаче"  default(50)
// @Success      200     {array}  models.UserWithPoints   "Список пользователей с их количеством очков"
// @Failure      500     {object} models.ErrorResponse    "Ошибка при получении лидерборда"
// @Router       /leaderboard [get]
func GetLeaderboard(service *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		users, err := service.GetLeaderboard(ctx, limit)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении лидерборда",
			})
			return
		}

		c.JSON(200, users)
	}
}
