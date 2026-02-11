package services

import (
	"bobri/internal/models"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
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

func (e *EmailProvider) SendResetPassword(email, rawToken string) error {
	from := e.fromEmail
	pass := e.password

	htmlBody := `
<h2>Здравствуйте!</h2>
<p>Вы запросили сброс пароля для своего аккаунта в системе <b>Beaver</b>.</p>

<p>Если вы не запрашивали сброс пароля - просто проигнорируйте это письмо.</p>

<p>
    Token: {{TOKEN}}
</p>

<hr>
<p style="font-size:12px;color:gray;">
С уважением,<br>
Команда поддержки <b>Beaver</b>
</p>`

	htmlBody = strings.ReplaceAll(htmlBody, "{{TOKEN}}", rawToken)

	msg := []byte(fmt.Sprintf(
		"Subject: Сброс пароля\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"From: %s\r\n"+
			"To: %s\r\n\r\n%s",
		from, email, htmlBody,
	))

	// адрес Gmail SMTP
	smtpAddr := "smtp.gmail.com:587"

	// сетевой диалер с тайм-аутом
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	// создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// dial с контекстом
	conn, err := dialer.DialContext(ctx, "tcp", smtpAddr)
	if err != nil {
		return fmt.Errorf("smtp dial error: %w", err)
	}

	client, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		return fmt.Errorf("smtp client error: %w", err)
	}
	defer func(client *smtp.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	// STARTTLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{
			ServerName: "smtp.gmail.com",
		}
		if err := client.StartTLS(config); err != nil {
			return fmt.Errorf("starttls error: %w", err)
		}
	}

	auth := smtp.PlainAuth("", from, pass, "smtp.gmail.com")
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth error: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from error: %w", err)
	}

	if err := client.Rcpt(email); err != nil {
		return fmt.Errorf("rcpt error: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data open error: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("data write error: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("data close error: %w", err)
	}

	return client.Quit()
}
