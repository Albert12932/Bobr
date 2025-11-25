package events

import (
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"strconv"
	"time"
)

// DeleteEvent  Удаление события по ID
// @Summary      Удалить событие
// @Description  Удаляет событие по его идентификатору. Требует прав администратора.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true   "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        id             path    int     true   "ID события для удаления"
// @Success      200  {object}  models.DeleteEventResponse "Событие успешно удалено"
// @Failure      400  {object}  models.ErrorResponse    "Некорректный ID события"
// @Failure      401  {object}  models.ErrorResponse    "Нет прав доступа"
// @Failure      404  {object}  models.ErrorResponse    "Событие с указанным ID не найдено"
// @Failure      500  {object}  models.ErrorResponse    "Ошибка сервера при удалении"
// @Router       /admin/delete_event/{id} [delete]
func DeleteEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var eventId int
		var err error
		// Получаем eventId события из параметров пути и преобразуем его в целое число
		{
			paramEventId := c.Param("id")
			if paramEventId == "" {
				c.JSON(400, models.ErrorResponse{
					Error:   "Event ID is not found",
					Message: "ID события не был передан в параметрах запроса",
				})
				return
			}
			// Преобразуем ID события в целое число
			eventId, err = strconv.Atoi(paramEventId)
			if err != nil {
				c.JSON(400, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Некорректный формат ID события",
				})
				return
			}
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос на удаление события
		tag, err := pool.Exec(ctx, "DELETE FROM events WHERE id = $1", eventId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при удалении события",
			})
			return
		}

		// Проверяем, было ли удалено какое-либо событие
		if tag.RowsAffected() == 0 {
			c.JSON(404, models.ErrorResponse{
				Error:   "Event not found",
				Message: "Событие не найдено",
			})
			return
		}

		// Возвращаем успешный ответ
		c.JSON(200, models.DeleteEventResponse{
			Successful: true,
			EventID:    eventId,
		})

		return

	}
}
