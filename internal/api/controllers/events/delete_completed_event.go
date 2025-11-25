package events

import (
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
	"time"
)

// DeleteCompletedEvent  Удалить выполнение события пользователем
// @Summary      Удаление выполнения события
// @Description  Удаляет отметку о выполнении события конкретным пользователем. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        user_id        path    int     true  "ID пользователя"
// @Param        event_id       path    int     true  "ID события"
// @Success      200  {object}  models.DeleteCompletedEventResponse        "Отметка о выполнении успешно удалена"
// @Failure      400  {object}  models.ErrorResponse           "Некорректный user_id или event_id"
// @Failure      401  {object}  models.ErrorResponse           "Нет прав доступа"
// @Failure      404  {object}  models.ErrorResponse           "Отметка о выполнении не найдена"
// @Failure      500  {object}  models.ErrorResponse           "Ошибка сервера при удалении"
// @Router       /admin/delete_completed_event/{user_id}/{event_id} [delete]
func DeleteCompletedEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userId, eventId int
		var err error
		// Проверяем наличие user_id и event_id в параметрах пути и преобразуем их в целые числа
		{
			// Получаем user_id и event_id из параметров пути
			userIdParam := c.Param("user_id")
			eventIdParam := c.Param("event_id")
			if userIdParam == "" || eventIdParam == "" {
				c.JSON(400, models.ErrorResponse{
					Error:   "user_id or event_id is missing",
					Message: "user_id или event_id отсутствует в параметрах запроса",
				})
				return
			}
			// Преобразуем user_id и event_id в целые числа
			userId, err = strconv.Atoi(userIdParam)
			if err != nil {
				c.JSON(400, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Некорректный формат user_id",
				})
				return
			}
			// Преобразуем event_id в целое число
			eventId, err = strconv.Atoi(eventIdParam)
			if err != nil {
				c.JSON(400, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Некорректный формат event_id",
				})
				return
			}
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос на удаление записи о выполнении события
		tag, err := pool.Exec(ctx,
			`DELETE from completed_events where user_id = $1 and event_id = $2`, userId, eventId)

		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при удалении выполненного события",
			})
			return
		}

		// Проверяем, была ли удалена какая-либо запись
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "Event not found RowsAffected=0",
				Message: "Выполненное событие не найдено",
			})
			return
		}

		// Возвращаем успешный ответ
		c.JSON(200, models.DeleteCompletedEventResponse{
			Successful: true,
			UserID:     userId,
			EventID:    eventId,
		})

		return
	}
}
