package services

import (
	"bobri/internal/models"
	"fmt"
	"net/smtp"
	"strings"
)

type EmailProvider struct {
	fromEmail string
	password  string
}

func NewEmailProvider(emailAuth models.EmailAuth) *EmailProvider {
	return &EmailProvider{
		fromEmail: emailAuth.EmailFrom,
		password:  emailAuth.EmailPass,
	}
}

// SendResetPassword отправляет письмо со ссылкой для сброса пароля
func (e *EmailProvider) SendResetPassword(email, rawToken string) error {
	from := e.fromEmail
	pass := e.password

	htmlBody := `
<h2>Здравствуйте!</h2>
<p>Вы запросили сброс пароля для своего аккаунта в системе <b>Beaver</b>.</p>

<p>Если вы не запрашивали сброс пароля — просто проигнорируйте это письмо.</p>

<p>
    Token: {{TOKEN}}
</p>

<hr>
<p style="font-size:12px;color:gray;">
С уважением,<br>
Команда поддержки <b>Beaver</b>
</p>`

	// Подставляем ссылку
	htmlBody = strings.ReplaceAll(htmlBody, "{{TOKEN}}", rawToken)

	// Формируем письмо
	msg := []byte(fmt.Sprintf(
		"Subject: Сброс пароля\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n"+
			"\r\n%s",
		from, email, htmlBody))

	smtpAddr := "smtp.gmail.com:587"
	auth := smtp.PlainAuth("", from, pass, "smtp.gmail.com")

	return smtp.SendMail(smtpAddr, auth, from, []string{email}, msg)
}
