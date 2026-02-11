package auth

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckStudent Проверка студенческого в системе
// @Summary      Проверка студенческого в системе
// @Description  Проверяет наличие студенческого билета в системе.
// @Description  Если студент не зарегистрирован, генерирует временный токен для регистрации.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.AuthBookRequest  true  "Номер студенческого билета"
// @Success      200  {object}  models.AuthStatus    "Студенческий найден, выдан регистрационный токен"
// @Failure      400  {object}  models.ErrorResponse "Некорректный запрос — ошибка парсинга JSON"
// @Failure      404  {object}  models.ErrorResponse "Студенческий не найден в базе"
// @Failure      409  {object}  models.ErrorResponse "Пользователь с таким номером уже зарегистрирован"
// @Failure      500  {object}  models.ErrorResponse "Ошибка при работе с базой данных"
// @Failure      500  {object}  models.ErrorResponse "Ошибка при генерации или сохранении токена"
// @Router       /auth/check [post]
func CheckStudent(service *services.StudentsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body models.AuthBookRequest

		// Берем номер студенческого из тела запроса
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Не получилось получить номер студенческого из тела запроса",
				})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := service.CheckStudent(ctx, body.BookId)

		// Обработка ошибок
		if err != nil {
			switch {
			case errors.Is(err, services.ErrStudentByBookIdNotFound):
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Студенческого с таким номером не найдено",
				})
				return
			case errors.Is(err, services.ErrUserAlreadyExists):
				c.JSON(http.StatusConflict, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Пользователь с таким номером студенческого уже зарегистрирован",
				})
				return
			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка сервера",
				})
				return
			}
		}
		c.JSON(200, resp)
	}
}
