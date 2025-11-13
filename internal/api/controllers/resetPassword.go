package controllers

import (
	"bobri/internal/models"
	"bobri/pkg/helpers"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// ResetPassword Запрос на сброс пароля
// @Summary      Запрос на сброс пароля
// @Description  Отправляет на почту пользователя ссылку для сброса пароля.
// @Description  Если пользователь с указанной почтой существует — ему придёт письмо с временной ссылкой на установку нового пароля.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.ResetPasswordRequest  true  "Почта пользователя"
// @Success      200  {object}  models.ResetPasswordResponse  "Инструкция отправлена на почту"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка при поиске пользователя или отправке письма"
// @Router       /auth/reset_password [post]
func ResetPassword(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Берем почту из тела запроса в json
		var body models.ResetPasswordRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат тела запроса",
			})
			return
		}

		// Создаем контекст с таймаутом 5 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Проверяем, что пользователь с такой почтой существует
		var userId int64
		err := pool.QueryRow(ctx,
			`SELECT id FROM users WHERE mail = $1`, body.Mail).Scan(&userId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при поиске пользователя по почте",
			})
			return
		}

		// Генерируем сырой токен
		rawToken, err := helpers.GenerateTokenRaw(32)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при генерации токена сброса пароля",
			})
			return
		}

		// Хэшируем токен перед сохранением в БД
		tokenHash := helpers.HashToken(rawToken)
		expiresAt := time.Now().Add(15 * time.Minute)
		createdAt := time.Now()

		// Сброс пароля
		_, err = pool.Exec(ctx, `INSERT into reset_password_tokens (user_id, mail, token_hash, expires_at, created_at) 
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id)
			DO UPDATE SET token_hash = EXCLUDED.token_hash,
			expires_at = EXCLUDED.expires_at,
			created_at = now();`,
			userId, body.Mail, tokenHash, expiresAt, createdAt)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при записи токена сброса пароля",
			})
			return
		}

		from := os.Getenv("FROM_EMAIL")
		pass := os.Getenv("EMAIL_PASSWORD")
		to := body.Mail

		htmlBody := `
			<h2>Здравствуйте!</h2>
			<p>Вы запросили сброс пароля для своего аккаунта в системе <b>Beaver</b>.</p>

			<p>Если вы не запрашивали сброс пароля — просто проигнорируйте это письмо.</p>
			<p>{{RESET_TOKEN}}</p>
			
			<hr>
			<p style="font-size:12px;color:gray;">
			С уважением,<br>
			Команда поддержки <b>Beaver</b>
			</p>`

		// Подставляем ссылку TODO
		htmlBody = strings.ReplaceAll(htmlBody, "{{RESET_LINK}}", "http://localhost:3000/reset-password")
		htmlBody = strings.ReplaceAll(htmlBody, "{{RESET_TOKEN}}", rawToken)

		// Формируем письмо с заголовками
		msg := []byte(fmt.Sprintf(
			"Subject: Сброс пароля для вашего аккаунта\r\n"+
				"MIME-Version: 1.0\r\n"+
				"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
				"From: %s\r\n"+
				"To: %s\r\n"+
				"\r\n%s", from, to, htmlBody))

		smtpAddr := "smtp.gmail.com:587"
		auth := smtp.PlainAuth("", from, pass, "smtp.gmail.com")

		err = smtp.SendMail(smtpAddr, auth, from, []string{to}, msg)
		if err != nil {
			log.Fatal("Ошибка при отправке:", err)
		}
		log.Println("Письмо отправлено успешно")

		c.JSON(200, models.ResetPasswordResponse{
			OK:      true,
			Mail:    body.Mail,
			Message: "Инструкция по сбросу пароля отправлена на указанную почту",
		})
		return
	}
}

// SetNewPassword Установка нового пароля
// @Summary      Установка нового пароля
// @Description  Устанавливает новый пароль по токену из письма. Токен действителен ограниченное время (15 минут).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  models.SetNewPasswordRequest  true  "Токен и новый пароль"
// @Success      200  {object}  models.SetNewPasswordResponse  "Пароль успешно обновлён"
// @Failure      400  {object}  models.ErrorResponse  "Некорректный JSON"
// @Failure      401  {object}  models.ErrorResponse  "Невалидный или истекший токен"
// @Failure      500  {object}  models.ErrorResponse  "Ошибка при обновлении пароля"
// @Router       /auth/set_new_password [post]
func SetNewPassword(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Берем токен и новый пароль из тела запроса в json
		var body models.SetNewPasswordRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Некорректный формат тела запроса",
			})
			return
		}

		// Создаем контекст с таймаутом 5 секунд
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Хэшируем полученный токен
		tokenHash := helpers.HashToken(body.Token)
		var userId int64

		// Проверяем валидность токена
		err := pool.QueryRow(ctx,
			`SELECT user_id FROM reset_password_tokens 
			WHERE token_hash = $1 AND expires_at > now()`, tokenHash).Scan(&userId)
		if err != nil {
			c.JSON(401, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Невалидный или истекший токен сброса пароля",
			})
			return
		}

		// Проверяем, что длина пароля не меньше 8 символов
		if len(body.NewPassword) < 8 {
			c.JSON(http.StatusBadRequest,
				models.ErrorResponse{
					Error:   "Weak password",
					Message: "Пароль должен быть не менее 8 символов"})
			return
		}

		// Хэшируем новый пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при хэшировании нового пароля",
			})
			return
		}

		// Обновляем пароль пользователя в базе
		_, err = pool.Exec(ctx,
			`UPDATE users SET password = $1 WHERE id = $2`,
			hashedPassword, userId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при обновлении пароля пользователя",
			})
			return
		}

		// Удаляем использованный токен сброса пароля
		_, err = pool.Exec(ctx,
			`DELETE FROM reset_password_tokens WHERE user_id = $1`, userId)
		if err != nil {
			c.JSON(500, models.ErrorResponse{
				Error:   err.Error(),
				Message: "Ошибка при удалении токена сброса пароля",
			})
			return
		}

		// Возвращаем успешный ответ
		c.JSON(200, models.SetNewPasswordResponse{
			OK:      true,
			Message: "Пароль успешно обновлен",
		})
		return
	}
}
