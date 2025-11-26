package auth

import (
	"bobri/internal/api/services"
	"bobri/internal/models"
	"context"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterByToken  Регистрация пользователя по токену
// @Summary      Регистрация пользователя по токену
// @Description  Регистрирует нового пользователя на основе временного токена, выданного после проверки студенческого билета.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.RegisterRequest  true  "Почта, пароль и регистрационный токен"
// @Success      201  {object}  models.RegisterResponse  "Успешная регистрация"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный формат JSON или токен истёк"
// @Failure      404  {object}  models.ErrorResponse  "Студент с таким номером студенческого не найден"
// @Failure      409  {object}  models.ErrorResponse  "Пользователь с такой почтой уже существует"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка сервера (база данных, хеширование, транзакция)"
// @Router       /auth/register [post]
func RegisterByToken(service *services.RegisterService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем пароль и почту из тела запроса в json
		var body models.RegisterRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Неверный формат тела запроса"})
			return
		}

		// Проверяем, что длина пароля не меньше 8 символов
		if len(body.Password) < 8 {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   "Weak password",
					Message: "Пароль должен быть не менее 8 символов"})
			return
		}

		// Проверяем валидность почты регулярным выражением
		validMail := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(body.Email)
		if !validMail {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   "Wrong email",
					Message: "Неправильный формат почты"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := service.RegisterUser(ctx, body.Email, body.Password, body.Token)
		if err != nil {

			switch {
			case errors.Is(err, services.ErrUserWithEmailAlreadyExists):
				c.JSON(http.StatusConflict, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Пользователь с такой почтой уже существует",
				})
				return

			case errors.Is(err, services.ErrTokenNotFound):
				c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Токен регистрации не найден или истёк",
				})
				return

			case errors.Is(err, services.ErrStudentByBookIdNotFound):
				c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Студент с таким номером студенческого не найден",
				})
				return

			case errors.Is(err, services.ErrRowsAffected):
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при обновлении токена регистрации",
				})
				return

			default:
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   err.Error(),
					Message: "Внутренняя ошибка сервера",
				})
				return
			}
		}

		c.JSON(http.StatusCreated, resp)
	}
}
