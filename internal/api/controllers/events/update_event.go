package events

import (
	"bobri/internal/models"
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// UpdateEvent  Частичное обновление события
// @Summary      Обновить выбранные поля события
// @Description  Производит частичное обновление данных события по его ID.
// @Description  Обновляются только те поля, которые переданы в теле запроса.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        Authorization  header   string  true  "Bearer токен авторизации. Формат: Bearer {token}"
// @Param        input          body     models.UpdateEventRequest  true  "ID события и новые значения полей"
// @Success      200  {object}  models.SuccessResponse  "Событие успешно обновлено"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON или ошибка валидации"
// @Failure      401  {object}  models.ErrorResponse  "Неавторизованный доступ — неверный или отсутствующий токен"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка сервера при попытке обновления записи"
// @Router       /admin/update_event [patch]
func UpdateEvent(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем модель из тела запроса
		var updateData models.UpdateEventRequest
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Error while marshaling JSON",
			})
			return
		}

		var builder sq.UpdateBuilder
		// Построение динамического SQL запроса с помощью Squirrel
		{
			builder := sq.Update("events")

			if updateData.NewData.Title != "" {
				builder = builder.Set("title", updateData.NewData.Title)
			}
			if updateData.NewData.Description != "" {
				builder = builder.Set("description", updateData.NewData.Description)
			}
			if updateData.NewData.Points != 0 {
				builder = builder.Set("points", updateData.NewData.Points)
			}
			if updateData.NewData.IconUrl != "" {
				builder = builder.Set("icon_url", updateData.NewData.IconUrl)
			}
			if updateData.NewData.EventDate != nil && !((*updateData.NewData.EventDate).IsZero()) {
				builder = builder.Set("event_date", updateData.NewData.EventDate)
			}
			if updateData.NewData.EventTypeCode != 0 {
				builder = builder.Set("event_type_code", updateData.NewData.EventTypeCode)
			}

			builder = builder.Where(sq.Eq{"id": updateData.EventId})
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Генерируем SQL запрос и аргументы
		sqlQuery, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()

		// Выполняем запрос
		_, err = pool.Exec(ctx, sqlQuery, args...)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении события",
			})
			return
		}

		// Возвращаем успешный ответ
		c.JSON(200, models.SuccessResponse{
			Successful: true,
			Message:    "Событие успешно обновлено",
		})

		return

	}
}
