package users

import (
	"bobri/internal/models"
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

// GetStudents Получение списка студентов
// @Summary      Получение списка студентов
// @Description  Возвращает всех студентов из таблицы students.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer токен" default(Bearer )
// @Success      200  {array}  models.Student             "Список студентов"
// @Failure      500  {object} models.ErrorResponse        "Ошибка при запросе или чтении данных"
// @Router       /admin/students [get]
func GetStudents(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var students []models.Student

		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Выполняем запрос к базе данных и сканируем результаты в слайс students
		err := pgxscan.Select(ctx, pool, &students,
			"select id, book_id, surname, name, middle_name, birth_date, student_group from students")
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении студентов",
			})
			return
		}

		// Возвращаем список студентов в формате JSON
		c.JSON(200, students)
		return
	}
}
