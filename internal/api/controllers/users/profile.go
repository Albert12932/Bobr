package users

import (
	"bobri/internal/models"
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

// GetProfile Получение профиля пользователя
// @Summary      Получение профиля пользователя
// @Description  Возвращает данные о пользователе
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer )
// @Success      200  {object} models.ProfileResponse               "Данные о пользователе"
// @Failure      500  {object} models.ErrorResponse       "Ошибка при запросе или чтении данных"
// @Router       /me/profile [get]
func GetProfile(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

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

		fmt.Println(payload.Sub)
		fmt.Println("Тут id")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var profileData models.ProfileResponse

		err := pgxscan.Get(ctx, pool, &profileData, `select coalesce(book_id, 0) as book_id, name, surname, middle_name, coalesce(birth_date, TO_DATE('01.01.1970', 'DD:MM:YYYY')) as birth_date, coalesce(student_group, '') as student_group, email, role_level from users where id = $1`, payload.Sub)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при сканировании пользователей",
			})
		}

		c.JSON(200, profileData)
		return
	}

}
