package users

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"github.com/gin-gonic/gin"
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
func GetStudents(service *services.StudentsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		students, err := service.GetStudents(ctx)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при получении студентов",
			})
			return
		}

		c.JSON(200, students)
	}
}
