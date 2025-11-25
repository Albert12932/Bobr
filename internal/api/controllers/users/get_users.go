package users

import (
	"bobri/internal/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
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
func GetUsers(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

		var roleLevel int64
		// Получаем уровень роли из payload
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
			roleLevel = payload.RoleLevel
		}

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос к базе данных и сканируем результаты в слайс users
		err := pgxscan.Select(ctx, pool, &users,
			"SELECT id, COALESCE(book_id, 0) as book_id, surname, name, middle_name, coalesce(birth_date, '1970-01-01'::timestamp) as birth_date, coalesce(student_group, '') as student_group, password, email, role_level FROM users where role_level <= $1", roleLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при сканировании пользователей",
			})
		}

		// Возвращаем список пользователей в формате JSON
		c.JSON(200, users)
		return
	}

}
