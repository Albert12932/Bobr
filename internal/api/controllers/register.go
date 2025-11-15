package controllers

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"errors"
	"github.com/georgysavva/scany/v2/pgxscan"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// RegisterByToken  Регистрация пользователя по токену
// @Summary      Регистрация пользователя по токену
// @Description  Регистрирует нового пользователя на основе временного токена, выданного после проверки студенческого билета.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.RegisterRequest  true  "Почта, пароль и токен регистрации пользователя"
// @Success      200  {object}  models.RegisterResponse     "Успешная регистрация"
// @Failure      400  {object}  models.ErrorResponse        "Некорректный запрос или формат данных"
// @Failure      401  {object}  models.ErrorResponse        "Токен не найден или истёк"
// @Failure      404  {object}  models.ErrorResponse        "Студент с такой почтой не найден"
// @Failure      409  {object}  models.ErrorResponse        "Пользователь с такой почтой уже существует"
// @Failure      500  {object}  models.ErrorResponse        "Ошибка сервера (база данных, хеширование, транзакция)"
// @Router       /auth/register [post]
func RegisterByToken(pool *pgxpool.Pool, accessJwtMaker *helpers.JWTMaker) gin.HandlerFunc {
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

		// Проверяем есть ли пользователь с такой почтой
		var used bool
		err := pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, body.Email).Scan(&used)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Ошибка при проверке почты"})
			return
		}
		if used {
			c.JSON(http.StatusConflict,
				models.ErrorResponse{
					Error:   "Email already used",
					Message: "Почта уже используется"})
			return
		}

		// Создаем контекст с 5-секундным таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Начинаем транзакцию
		tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while creating transaction"})
			return
		}
		defer func() { _ = tx.Rollback(ctx) }()

		// Проверяем не истек ли токен
		var bookID int
		err = tx.QueryRow(ctx, `DELETE FROM link_tokens
       WHERE token_hash = $1 AND expires_at > now()
       RETURNING book_id`, helpers.HashToken(body.Token)).Scan(&bookID)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusUnauthorized,
					models.ErrorResponse{
						Error:   err.Error(),
						Message: "Token не найден или истёк",
					})
				return
			}
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "DATABASE_ERROR"})
			return
		}

		// Берем данные из students для последующей вставки в users
		var student models.Student
		err = pgxscan.Get(ctx, pool, &student, `
			SELECT book_id, name, surname, middle_name, student_group, birth_date
			FROM students
			WHERE book_id = $1`,
			bookID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusNotFound,
					models.ErrorResponse{
						Error:   "No rows while getting student info",
						Message: "Нет данных о студенте"})
				return
			}
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   "Error while getting student info",
					Message: err.Error()})
			return
		}

		// Генерируем хэш пароля
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost) // cost=10 по умолчанию
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while hashing password"})
			return
		}

		// Вставляем в users данные пользователя
		var userID int64
		roleLevel := int64(10)
		err = tx.QueryRow(ctx, `
			INSERT INTO users (book_id, name, surname, middle_name, student_group, birth_date, password, email, role_level)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`,
			student.BookId, student.Name, student.Surname, student.MiddleName, student.Group, student.BirthDate,
			hash, body.Email, roleLevel).Scan(&userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while inserting user info"})
			return
		}

		// Коммитим транзакцию
		if err := tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError,
				models.ErrorResponse{
					Error:   err.Error(),
					Message: "Error while commiting transaction"})
			return
		}

		accessToken, exp, refreshToken, errResp := GetPairOfTokens(pool, accessJwtMaker,
			models.GetTokensRequest{UserId: userID, RoleLevel: roleLevel})

		if errResp != (models.ErrorResponse{}) {
			c.JSON(http.StatusInternalServerError, errResp)
			return
		}

		// Выдаем ответ в нужном формате
		var resp models.RegisterResponse
		resp.UserSubstructure.ID = userID
		resp.UserSubstructure.BookId = student.BookId
		resp.UserSubstructure.Email = body.Email
		resp.UserSubstructure.FirstName = student.Name
		resp.UserSubstructure.Group = student.Group
		resp.UserSubstructure.RoleLevel = roleLevel
		resp.Auth.AccessToken = accessToken
		resp.Auth.ExpUnix = exp
		resp.Auth.RefreshToken = refreshToken

		// на всякий случай отключим кеш
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, resp)
	}
}
